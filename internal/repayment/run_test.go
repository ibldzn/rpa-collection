package repayment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func reply(r *http.Request, body string) *http.Response {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}
}

func TestSetLoanRepaymentAccountResolvesLoanAndUsesLoanBranch(t *testing.T) {
	for _, test := range []struct {
		loan, saving string
		wantLogins   []string
	}{
		{"3000010000000011", "001000OPER", []string{"001"}},
		{"3000010000000011", "002123456789", []string{"001"}},
		{"0130101415", "002123456789", []string{"lookup", "001"}},
	} {
		t.Run(test.loan+"/"+test.saving, func(t *testing.T) {
			var logins []string
			var submits int
			web := roundTrip(func(r *http.Request) (*http.Response, error) {
				switch r.URL.Path {
				case "/admin/access/login":
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					location := r.Form.Get("locationid")
					logins = append(logins, location)
					return reply(r, fmt.Sprintf(`{"status":"ok","data":{"result":{"sessionid":%q}}}`, "session-"+location)), nil
				case "/pinjaman/inquiry/rekening/cari":
					if r.Header.Get("sessionid") != "session-lookup" || r.URL.Query().Get("cabang") != "ALL" {
						t.Error("wrong alternate lookup location")
					}
					return reply(r, `{"status":"ok","data":{"result":[{"id":"3000010000000011"}]}}`), nil
				case "/pinjaman/pendaftaranPenghapusanAutodebit/pembuatan/cari":
					if r.Header.Get("sessionid") != "session-001" || r.URL.Query().Get("norekening") != "3000010000000011" {
						t.Error("wrong loan inquiry location/account")
					}
					return reply(r, `{"status":"ok","data":{"result":{"id":"3000010000000011"}}}`), nil
				case "/pinjaman/pendaftaranPenghapusanAutodebit/pembuatan/pinjaman":
					submits++
					if r.Header.Get("sessionid") != "session-001" {
						t.Error("wrong mutation location")
					}
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if r.Form.Get("norektab_bayarangsuran") != test.saving || r.Form.Get("tabbayar_namapemilik") != "Sariman" || r.Form.Get("status_dokumen") != "Diajukan" {
						t.Error("wrong mutation form", r.Form)
					}
					return reply(r, `{"status":"ok"}`), nil
				}
				t.Errorf("unexpected path %s", r.URL.Path)
				return reply(r, `{}`), nil
			})
			api := roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Query().Get("accountNumber") != test.saving {
					t.Error("wrong saving inquiry")
				}
				return reply(r, fmt.Sprintf(`{"responseCode":"00","data":{"accountNumber":%q,"customerName":"Sariman","documentStatus":"Aktif","currency":"IDR"}}`, test.saving)), nil
			})
			cfg := Config{
				Kolek:      kolek.Config{BaseURL: "http://web.test", Credentials: fincloud.Credentials{Username: "user", Password: "pass", RoleID: "role"}, LookupLocationID: "lookup", HTTPClient: &http.Client{Transport: web}},
				APIBaseURL: "http://api.test", APISecretKey: "secret", APIHTTPClient: &http.Client{Transport: api},
			}
			result, err := SetLoanRepaymentAccount(context.Background(), test.loan, test.saving, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if result.PrimaryLoanAccount != "3000010000000011" || result.Branch != "001" || result.SavingAccount != test.saving || submits != 1 {
				t.Fatalf("result %+v submits %d", result, submits)
			}
			if strings.Join(logins, ",") != strings.Join(test.wantLogins, ",") {
				t.Fatalf("logins %v, want %v", logins, test.wantLogins)
			}
		})
	}
}

func TestSetLoanRepaymentAccountInvalidAndMissingAlternate(t *testing.T) {
	cfg := Config{Kolek: kolek.Config{BaseURL: "http://web.test", Credentials: fincloud.Credentials{Username: "user", Password: "pass", RoleID: "role"}, LookupLocationID: "lookup"}, APIBaseURL: "http://api.test", APISecretKey: "secret"}
	if _, err := SetLoanRepaymentAccount(context.Background(), "invalid", "001000OPER", cfg); err == nil {
		t.Fatal("invalid loan accepted")
	}
	web := roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/admin/access/login" {
			return reply(r, `{"status":"ok","data":{"result":{"sessionid":"session"}}}`), nil
		}
		return reply(r, `{"status":"ok","data":{"result":[]}}`), nil
	})
	cfg.Kolek.HTTPClient = &http.Client{Transport: web}
	if _, err := SetLoanRepaymentAccount(context.Background(), "0130101415", "001000OPER", cfg); !errors.Is(err, fincloud.ErrDataNotFound) {
		t.Fatalf("lookup error = %v", err)
	}
}
