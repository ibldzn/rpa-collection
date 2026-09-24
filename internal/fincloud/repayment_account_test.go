package fincloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/ibldzn/kolek-rpa/internal/fincloudapi"
)

const autodebitRemovalLoanResponse = `{
  "status":"ok",
  "data":{"result":{
    "lokasi":"001",
    "id":"3000010000000011",
    "nopk":"PL001000073838",
    "namanasabah":"Sariman",
    "aliasnama":"Sariman",
    "statusrekening":"Aktif",
    "tgl_pencairan":{"date":"2013-04-18 00:00:00.000000","timezone_type":3,"timezone":"UTC"},
    "rec_dibuat_oleh":"IT",
    "noalt":"0130101415",
    "produk_jenispinjaman":"Kredit Angsuran",
    "produkid":"301 - Kredit Pegawai Aktif",
    "idproduk":"301",
    "currency":"IDR",
    "plafondlimit":30000000,
    "jmlpokok_pinjaman":30000000,
    "jangkawaktu":"60 Month",
    "produk_jenisangsuran":"in Arrear",
    "tgl_angsuran":18,
    "sid_sifatkredit2":"9",
    "sid_jenispenggunaan":"3",
    "sid_sumberdanapelunasan":"1",
    "sid_golongankredit":"NU",
    "sid_orientasipenggunaan":"3",
    "sid_sektorekonomi":"009000",
    "sid_sektorekonomi2":"1020",
    "sid_sifatkredit":"9",
    "datapenjamin":"Tidak Ada",
    "mengetahuisuamiistri":"Ya",
    "sid_jenisusaha":"4",
    "bungaflat":18,
    "noperjanjiankredit":"1059454/IV/2013",
    "persendendatunggakan":0,
    "titipan":0,
    "periode":"2013-04-18  -  2018-04-18",
    "produk_sukubunga":28.88,
    "produk_perubahansukubunga":"Bunga Fix",
    "pejabatkredit":null,
    "pejabatkreditdua":null,
    "tujuankredit":null,
    "tglterakhir_bayarpokokdanbunga":null,
    "tglbayarpokokbunga_berikutnya":{"date":"2014-04-18 00:00:00.000000","timezone_type":3,"timezone":"UTC"},
    "tgljtterakhir":{"date":"2018-04-18 00:00:00.000000","timezone_type":3,"timezone":"UTC"},
    "tgljtberikutnya":null,
    "outstandingpinjaman":18950000,
    "tunggakanpokok":18950000,
    "accrue":0,
    "decimalpoint":2,
    "tunggakanbunga":19350000,
    "dendatunggakan":0,
    "dpd":4514,
    "kolekbi":5,
    "kolekbpr":5,
    "updatekolekbi":"Automatic",
    "totalcollateralvalue":0,
    "totalassetvalue":0,
    "nocif":"00100000367",
    "jenisnasabah":"Perorangan",
    "tglbukacif":{"date":"2015-04-18 00:00:00.000000","timezone_type":3,"timezone":"UTC"},
    "status_dokumen":"Aktif",
    "dataalamat_ktp_alamat1":"Kp Rawa Bogo Rt04 Rw01",
    "dataalamat_ktp_alamat2":null,
    "dataalamat_ktp_rt":"000",
    "dataalamat_ktp_rw":"000",
    "dataalamat_ktp_kelurahan":"Jatimekar",
    "dataalamat_ktp_kecamatan":"Jati Asih",
    "dataalamat_ktp_kota":"0102",
    "dataalamat_ktp_propinsi":"0",
    "dataalamat_ktp_kodepos":"17422",
    "norektab_pencairanpinjaman":"001000OPER - Internal Account 001 (Kantor Pusat Operasional)",
    "norektab_bayarangsuran":null
  }}
}`

func TestSetLoanRepaymentAccountBuildsAndSubmitsForm(t *testing.T) {
	var order []string
	var encoded string
	webTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case autodebitRemovalInquiryPath:
			order = append(order, "loan")
			if req.Method != http.MethodGet || req.URL.Query().Get("norekening") != "3000010000000011" {
				t.Errorf("loan inquiry = %s %s", req.Method, req.URL.String())
			}
			return textResponse(req, http.StatusOK, autodebitRemovalLoanResponse), nil
		case autodebitRemovalSubmitPath:
			order = append(order, "submit")
			if req.Method != http.MethodPost {
				t.Errorf("submit method = %s", req.Method)
			}
			if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type = %q", got)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			encoded = string(body)
			return textResponse(req, http.StatusOK, `{"status":"ok"}`), nil
		default:
			return nil, fmt.Errorf("unexpected web path %q", req.URL.Path)
		}
	})
	apiTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		order = append(order, "saving")
		if req.URL.Path != "/saving/inq/balance" || req.URL.Query().Get("accountNumber") != "001000OPER" {
			t.Errorf("saving inquiry = %s", req.URL.String())
		}
		return textResponse(req, http.StatusOK, `{"responseCode":"00","description":"SUCCESS","data":{"accountNumber":"001000OPER","customerName":"Internal Account 001 (Kantor Pusat Operasional)","documentStatus":"Aktif","currency":"IDR"}}`), nil
	})
	client := newAutodebitRemovalTestClient(t, webTransport, apiTransport)

	if err := client.SetLoanRepaymentAccount(context.Background(), "3000010000000011", "001000OPER"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"loan", "saving", "submit"}) {
		t.Fatalf("request order = %v", order)
	}
	form, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"id":                            "3000010000000011",
		"status_dokumen":                "Diajukan",
		"norektab_bayarangsuran":        "001000OPER",
		"tabbayar_namapemilik":          "Internal Account 001 (Kantor Pusat Operasional)",
		"tabbayar_status":               "Aktif",
		"tabbayar_currency":             "IDR",
		"tgl_pencairan":                 "2013-4-18",
		"tglbayarpokokbunga_berikutnya": "2014-4-18",
		"tgljtterakhir":                 "2018-4-18",
		"tglbukacif":                    "2015-4-18",
		"plafondlimit":                  "30000000",
		"bungaflat":                     "18",
		"produk_sukubunga":              "28.88",
		"persendendatunggakan":          "0",
		"noperjanjiankredit":            "1059454/IV/2013",
		"norektab_pencairanpinjaman":    "001000OPER - Internal Account 001 (Kantor Pusat Operasional)",
	} {
		if got := form.Get(key); got != want {
			t.Errorf("form %s = %q, want %q", key, got, want)
		}
	}
	for _, key := range []string{
		"pejabatkredit", "pejabatkreditdua", "tujuankredit",
		"tglterakhir_bayarpokokdanbunga", "tgljtberikutnya", "dataalamat_ktp_alamat2",
	} {
		if _, ok := form[key]; ok {
			t.Errorf("nullable field %s was submitted as %q", key, form.Get(key))
		}
	}
	for _, want := range []string{
		"produk_jenispinjaman=Kredit+Angsuran",
		"noperjanjiankredit=1059454%2FIV%2F2013",
		"tabbayar_namapemilik=Internal+Account+001+%28Kantor+Pusat+Operasional%29",
	} {
		if !strings.Contains(encoded, want) {
			t.Errorf("encoded form missing %q: %s", want, encoded)
		}
	}
	for _, bogus := range []string{"<nil>", "null", "0001-1-1", "3e%2B07", "3e+07"} {
		if strings.Contains(encoded, bogus) {
			t.Errorf("encoded form contains bogus value %q", bogus)
		}
	}
}

func TestBuildAutodebitRemovalFormUsesSuppliedSaving(t *testing.T) {
	response := strings.Replace(autodebitRemovalLoanResponse, `"norektab_bayarangsuran":null`, `"norektab_bayarangsuran":"999999OPER - old"`, 1)
	if !strings.Contains(response, "999999OPER") {
		t.Fatal("stale repayment account was not added to the fixture")
	}
	var body struct {
		Data struct {
			Result autodebitRemovalLoan `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(response), &body); err != nil {
		t.Fatal(err)
	}
	form, err := buildAutodebitRemovalForm(body.Data.Result, "002123456789", fincloudapi.SavingBalanceInquiryResponse{
		AccountNumber:  "002123456789",
		CustomerName:   "Branch Office Customer",
		DocumentStatus: "Aktif",
		Currency:       "IDR",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := form.Get("norektab_bayarangsuran"); got != "002123456789" {
		t.Fatalf("repayment account = %q", got)
	}
	if strings.Contains(form.Encode(), "999999OPER") {
		t.Fatalf("old repayment account leaked into form: %s", form.Encode())
	}
}

func TestSetLoanRepaymentAccountValidatesInputs(t *testing.T) {
	for _, test := range []struct {
		name, loan, saving string
	}{
		{name: "empty loan", saving: "001000OPER"},
		{name: "empty saving", loan: "3000010000000011"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := (&Client{}).SetLoanRepaymentAccount(context.Background(), test.loan, test.saving); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestSetLoanRepaymentAccountStopsOnLoanError(t *testing.T) {
	for _, test := range []struct {
		name       string
		statusCode int
		body       string
		want       string
	}{
		{name: "HTTP", statusCode: http.StatusBadGateway, body: "unavailable", want: "HTTP 502"},
		{name: "API status", statusCode: http.StatusOK, body: `{"status":"error"}`, want: `status "error"`},
		{name: "missing result", statusCode: http.StatusOK, body: `{"status":"ok","data":{}}`, want: "missing result"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var apiCalls int
			client := newAutodebitRemovalTestClient(t,
				roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return textResponse(req, test.statusCode, test.body), nil
				}),
				roundTripFunc(func(req *http.Request) (*http.Response, error) {
					apiCalls++
					return nil, fmt.Errorf("unexpected saving inquiry")
				}),
			)
			if err := client.SetLoanRepaymentAccount(context.Background(), "3000010000000011", "001000OPER"); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v", err)
			}
			if apiCalls != 0 {
				t.Fatalf("saving inquiries = %d", apiCalls)
			}
		})
	}
}

func TestSetLoanRepaymentAccountStopsOnSavingError(t *testing.T) {
	var submits int
	client := newAutodebitRemovalTestClient(t,
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == autodebitRemovalSubmitPath {
				submits++
			}
			return textResponse(req, http.StatusOK, autodebitRemovalLoanResponse), nil
		}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return textResponse(req, http.StatusOK, `{"responseCode":"99","description":"not found"}`), nil
		}),
	)
	if err := client.SetLoanRepaymentAccount(context.Background(), "3000010000000011", "001000OPER"); !errors.Is(err, ErrDataNotFound) {
		t.Fatalf("error = %v", err)
	}
	if submits != 0 {
		t.Fatalf("submits = %d", submits)
	}
}

func TestSetLoanRepaymentAccountValidatesSavingInquiryBeforeSubmit(t *testing.T) {
	for _, test := range []struct {
		name, account, customer, status, currency string
		wantErr                                   error
	}{
		{"normal saving", "001123456789", "Sariman", "Aktif", "IDR", nil},
		{"cross branch saving", "002123456789", "Sariman", " aktif ", "IDR", nil},
		{"mismatched account", "009999999999", "Sariman", "Aktif", "IDR", ErrInvalidSavingAccount},
		{"missing customer", "001123456789", "", "Aktif", "IDR", ErrInvalidSavingAccount},
		{"missing status", "001123456789", "Sariman", "", "IDR", ErrInvalidSavingAccount},
		{"missing currency", "001123456789", "Sariman", "Aktif", "", ErrInvalidSavingAccount},
		{"inactive", "001123456789", "Sariman", "Tutup", "IDR", ErrInactiveSavingAccount},
	} {
		t.Run(test.name, func(t *testing.T) {
			var submits int
			web := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == autodebitRemovalSubmitPath {
					submits++
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					for key, want := range map[string]string{
						"norektab_bayarangsuran": "001123456789",
						"tabbayar_namapemilik":   "Sariman",
						"tabbayar_status":        test.status,
						"tabbayar_currency":      "IDR",
					} {
						if key == "norektab_bayarangsuran" && test.name == "cross branch saving" {
							want = "002123456789"
						}
						if got := req.Form.Get(key); got != want {
							t.Errorf("%s = %q, want %q", key, got, want)
						}
					}
					return textResponse(req, http.StatusOK, `{"status":"ok"}`), nil
				}
				return textResponse(req, http.StatusOK, autodebitRemovalLoanResponse), nil
			})
			api := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got := req.URL.Query().Get("accountNumber"); got != "001123456789" && got != "002123456789" {
					t.Errorf("requested saving = %q", got)
				}
				body := fmt.Sprintf(`{"responseCode":"00","data":{"accountNumber":%q,"customerName":%q,"documentStatus":%q,"currency":%q}}`, test.account, test.customer, test.status, test.currency)
				return textResponse(req, http.StatusOK, body), nil
			})
			client := newAutodebitRemovalTestClient(t, web, api)
			requested := "001123456789"
			if test.name == "cross branch saving" {
				requested = "002123456789"
			}
			err := client.SetLoanRepaymentAccount(context.Background(), "3000010000000011", requested)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			wantSubmits := 1
			if test.wantErr != nil {
				wantSubmits = 0
			}
			if submits != wantSubmits {
				t.Fatalf("submits = %d, want %d", submits, wantSubmits)
			}
		})
	}
}

func TestSetLoanRepaymentAccountPropagatesSubmitError(t *testing.T) {
	client := newAutodebitRemovalTestClient(t,
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == autodebitRemovalSubmitPath {
				return textResponse(req, http.StatusOK, `{"status":"error"}`), nil
			}
			return textResponse(req, http.StatusOK, autodebitRemovalLoanResponse), nil
		}),
		roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return textResponse(req, http.StatusOK, `{"responseCode":"00","description":"SUCCESS","data":{"accountNumber":"001000OPER","customerName":"Internal Account","documentStatus":"Aktif","currency":"IDR"}}`), nil
		}),
	)
	if err := client.SetLoanRepaymentAccount(context.Background(), "3000010000000011", "001000OPER"); err == nil || !strings.Contains(err.Error(), `status "error"`) {
		t.Fatalf("error = %v", err)
	}
}

func newAutodebitRemovalTestClient(t *testing.T, webTransport, apiTransport http.RoundTripper) *Client {
	t.Helper()
	api, err := fincloudapi.NewClient(
		fincloudapi.WithBaseURL("http://api.test"),
		fincloudapi.WithSecretKey("secret"),
		fincloudapi.WithHTTPClient(&http.Client{Transport: apiTransport}),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := newTestClient(t, webTransport)
	client.api = api
	return client
}
