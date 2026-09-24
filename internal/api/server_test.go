package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
	"github.com/ibldzn/kolek-rpa/internal/repayment"
)

func TestAPIAuthenticationAndHealth(t *testing.T) {
	handler := testServer().Handler()
	for _, test := range []struct {
		token  string
		status int
	}{
		{"", 401}, {"wrong", 401}, {"secret", 200},
	} {
		got := request(t, handler, http.MethodPost, "/api/v1/kolek", `{"kolek":1,"change_type":"Manual","accounts":["3000010000000010"]}`, test.token)
		if got.Code != test.status {
			t.Errorf("token %q: HTTP %d, want %d", test.token, got.Code, test.status)
		}
	}
	got := request(t, handler, http.MethodGet, "/health", "", "")
	if got.Code != 200 || !strings.Contains(got.Body.String(), `"status":"ok"`) {
		t.Fatalf("health: %d %s", got.Code, got.Body.String())
	}
	got = request(t, handler, http.MethodGet, "/api/v1/kolek", "", "secret")
	if got.Code != 405 || got.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("method: %d %s", got.Code, got.Body.String())
	}
}

func TestKolekRequestsAndPartialSuccess(t *testing.T) {
	handler := testServer().Handler()
	for _, test := range []struct {
		body   string
		status int
	}{
		{`{"kolek":5,"change_type":"Manual","accounts":["3000010000000010"]}`, 200},
		{`{`, 400},
		{`{"kolek":0,"change_type":"Manual","accounts":["x"]}`, 400},
		{`{"kolek":6,"change_type":"Manual","accounts":["x"]}`, 400},
		{`{"kolek":5,"change_type":"Other","accounts":["x"]}`, 400},
		{`{"kolek":5,"change_type":"Manual","accounts":[]}`, 400},
		{`{"kolek":5,"change_type":"Manual","accounts":["x"],"unknown":true}`, 400},
	} {
		got := request(t, handler, http.MethodPost, "/api/v1/kolek", test.body, "secret")
		if got.Code != test.status {
			t.Errorf("body %q: HTTP %d, want %d: %s", test.body, got.Code, test.status, got.Body.String())
		}
	}
	got := request(t, handler, http.MethodPost, "/api/v1/kolek", `{"kolek":5,"change_type":"Manual","accounts":["a","b","c"]}`, "secret")
	if got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	var body struct {
		Status string `json:"status"`
		Data   struct {
			Processed int `json:"processed"`
			Success   int `json:"success"`
			Skipped   int `json:"skipped"`
			Failed    int `json:"failed"`
			Results   []struct {
				Error string `json:"error"`
			} `json:"results"`
		} `json:"data"`
	}
	if err := json.Unmarshal(got.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Processed != 3 || body.Data.Success != 1 || body.Data.Skipped != 1 || body.Data.Failed != 1 || body.Data.Results[2].Error != "alternate lookup: data not found" {
		t.Fatalf("partial result: %+v", body)
	}
}

func TestRepaymentAccountRequests(t *testing.T) {
	handler := testServer().Handler()
	for _, test := range []struct {
		body    string
		status  int
		primary string
	}{
		{`{"loan_account":"3000010000000011","saving_account":"001000OPER"}`, 200, "3000010000000011"},
		{`{"loan_account":"3000010000000011","saving_account":"001123456789"}`, 200, "3000010000000011"},
		{`{"loan_account":"0130101415","saving_account":"001123456789"}`, 200, "3000010000000011"},
		{`{`, 400, ""},
		{`{"saving_account":"001000OPER"}`, 400, ""},
		{`{"loan_account":"3000010000000011"}`, 400, ""},
		{`{"loan_account":"3000010000000011","saving_account":"inactive"}`, 422, ""},
		{`{"loan_account":"3000010000000011","saving_account":"upstream"}`, 502, ""},
	} {
		got := request(t, handler, http.MethodPost, "/api/v1/repayment-account", test.body, "secret")
		if got.Code != test.status {
			t.Errorf("body %s: HTTP %d, want %d: %s", test.body, got.Code, test.status, got.Body.String())
			continue
		}
		if test.primary != "" && !strings.Contains(got.Body.String(), `"primary_loan_account":"`+test.primary+`"`) {
			t.Errorf("missing resolved primary: %s", got.Body.String())
		}
		if got.Header().Get("Content-Type") != "application/json" {
			t.Error("missing JSON Content-Type")
		}
	}
}

func testServer() Server {
	return Server{
		Key: "secret",
		UpdateKolek: func(_ context.Context, target int, _ string, accounts []string) []kolek.ProcessResult {
			if len(accounts) == 1 {
				return []kolek.ProcessResult{{LoanTarget: kolek.LoanTarget{InputAccount: accounts[0]}, Status: kolek.ProcessSuccess, NewKolek: target}}
			}
			return []kolek.ProcessResult{
				{LoanTarget: kolek.LoanTarget{InputAccount: "a"}, Status: kolek.ProcessSuccess, NewKolek: target},
				{LoanTarget: kolek.LoanTarget{InputAccount: "b"}, Status: kolek.ProcessSkipped, NewKolek: target, Reason: "already set"},
				{LoanTarget: kolek.LoanTarget{InputAccount: "c"}, Status: kolek.ProcessFailed, NewKolek: target, Err: fincloud.ErrDataNotFound},
			}
		},
		SetRepaymentAccount: func(_ context.Context, loan, saving string) (repayment.Result, error) {
			if saving == "inactive" {
				return repayment.Result{}, fincloud.ErrInactiveSavingAccount
			}
			if saving == "upstream" {
				return repayment.Result{}, errors.New("sessionid=private upstream response")
			}
			primary := loan
			if loan == "0130101415" {
				primary = "3000010000000011"
			}
			return repayment.Result{InputLoanAccount: loan, PrimaryLoanAccount: primary, Branch: "001", SavingAccount: saving, SavingAccountName: "Sariman", SavingStatus: "Aktif", Currency: "IDR"}, nil
		},
	}
}

func request(t *testing.T, handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	got := httptest.NewRecorder()
	handler.ServeHTTP(got, req)
	if strings.Contains(got.Body.String(), "sessionid=private") {
		t.Fatal("upstream secret leaked")
	}
	return got
}
