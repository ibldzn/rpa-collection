package kolek

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
)

func TestBranchAndGrouping(t *testing.T) {
	for account, want := range map[string]string{
		"3000010000000010": "001",
		"3000070000000010": "007",
	} {
		got, err := branchFromPrimary(account)
		if err != nil || got != want {
			t.Fatalf("branchFromPrimary(%q) = %q, %v; want %q", account, got, err, want)
		}
	}
	if _, err := branchFromPrimary("123"); err == nil {
		t.Fatal("short primary account accepted")
	}
	targets := []LoanTarget{
		{PrimaryAccount: "3000020000000001", BranchCode: "002", inputIndex: 0},
		{PrimaryAccount: "3000010000000002", BranchCode: "001", inputIndex: 1},
		{PrimaryAccount: "3000020000000003", BranchCode: "002", inputIndex: 2},
		{PrimaryAccount: "3000010000000004", BranchCode: "001", inputIndex: 3},
	}
	branches, groups := groupByBranch(targets)
	if !reflect.DeepEqual(branches, []string{"001", "002"}) {
		t.Fatalf("branches = %v", branches)
	}
	for branch, want := range map[string][]string{
		"001": {"3000010000000002", "3000010000000004"},
		"002": {"3000020000000001", "3000020000000003"},
	} {
		var got []string
		for _, item := range groups[branch] {
			got = append(got, item.PrimaryAccount)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("branch %s accounts = %v, want %v", branch, got, want)
		}
	}
}

func TestRunResolvesGroupsAndSubmits(t *testing.T) {
	var mu sync.Mutex
	logins := make(map[string]int)
	updates := make(map[string]url.Values)
	var lookupAccounts []string
	var inquiryAccounts []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/access/login":
			if err := r.ParseForm(); err != nil {
				t.Error(err)
				return
			}
			location := r.Form.Get("locationid")
			mu.Lock()
			logins[location]++
			mu.Unlock()
			if location == "007" {
				http.Error(w, "rejected", http.StatusForbidden)
				return
			}
			fmt.Fprintf(w, `{"status":"ok","data":{"result":{"sessionid":%q}}}`, "session-"+location)
		case "/pinjaman/inquiry/rekening/cari":
			if r.Header.Get("sessionid") != "session-lookup" || r.URL.Query().Get("cabang") != "ALL" {
				t.Errorf("alternate lookup used wrong session or query")
			}
			alt := r.URL.Query().Get("noalt")
			mu.Lock()
			lookupAccounts = append(lookupAccounts, alt)
			mu.Unlock()
			if alt == "0130101394" {
				fmt.Fprint(w, `{"status":"ok","data":{"result":[{"id":"3000010000000010"}]}}`)
			} else {
				fmt.Fprint(w, `{"status":"ok","data":{"result":[]}}`)
			}
		case "/pinjaman/updateManualKolek/pembuatan/cari":
			account := r.URL.Query().Get("norekening")
			branch := account[3:6]
			if got := r.Header.Get("sessionid"); got != "session-"+branch {
				t.Errorf("inquiry session = %q for %s", got, account)
			}
			mu.Lock()
			inquiryAccounts = append(inquiryAccounts, account)
			mu.Unlock()
			if strings.HasSuffix(account, "0003") {
				fmt.Fprint(w, `{"status":"error"}`)
				return
			}
			if strings.HasSuffix(account, "0004") {
				account = "wrong-account"
			}
			bi, bpr := 5, 2
			updateBI, updateBPR := "Automatic", "Manual"
			switch {
			case strings.HasSuffix(account, "0002"):
				bi, bpr = 1, 1
				updateBI, updateBPR = "manual", "MANUAL"
			case strings.HasSuffix(account, "0005"):
				bi, bpr = 1, 1
			case strings.HasSuffix(account, "0006"):
				bi, bpr = 1, 1
				updateBI = ""
			}
			fmt.Fprintf(w, `{"status":"ok","data":{"result":{"norekening":%q,"namanasabah":"Ratna Juwita","nopk":"PL001000073837","appdate":{"date":"2026-08-27 00:00:00.000000"},"datarekening":{"kolekbimanual":0,"kolekbprmanual":1,"kolekbiauto":3,"kolekbprauto":4,"kolekbi":%d,"kolekbpr":%d,"updatekolekbi":%q,"updatekolekbpr":%q,"dpd":4423,"totalassetvalue":0,"totalcollateralvalue":0}}}}`, account, bi, bpr, updateBI, updateBPR)
		case "/pinjaman/updateManualKolek/pembuatan/pinjaman":
			if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type = %q", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Error(err)
				return
			}
			account := r.Form.Get("norekening")
			if got := r.Header.Get("sessionid"); got != "session-"+account[3:6] {
				t.Errorf("update session = %q for %s", got, account)
			}
			mu.Lock()
			updates[account] = r.Form
			mu.Unlock()
			fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			http.NotFound(w, r)
		}
	})
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Result(), nil
	})

	config := Config{
		BaseURL:          "http://fincloud.test",
		Credentials:      fincloud.Credentials{Username: "user", Password: "password", RoleID: "role"},
		LookupLocationID: "lookup",
		HTTPClient:       &http.Client{Transport: transport},
	}
	inputs := []string{
		"3000020000000001", " 0130101394 ", "3000010000000002",
		"3000010000000005", "3000010000000006",
		"3000020000000003", "3000020000000004", "3000070000000010",
		"3000070000000011", "123", "0230109999",
	}
	results := Run(context.Background(), 1, "Manual", inputs, config)
	if len(results) != len(inputs) {
		t.Fatalf("results = %d, want %d", len(results), len(inputs))
	}
	byInput := make(map[string]ProcessResult)
	for _, result := range results {
		byInput[strings.TrimSpace(result.InputAccount)] = result
	}
	for input, want := range map[string]ProcessStatus{
		"3000020000000001": ProcessSuccess,
		"0130101394":       ProcessSuccess,
		"3000010000000002": ProcessSkipped,
		"3000010000000005": ProcessSuccess,
		"3000010000000006": ProcessFailed,
		"3000020000000003": ProcessFailed,
		"3000020000000004": ProcessFailed,
		"3000070000000010": ProcessFailed,
		"3000070000000011": ProcessFailed,
		"123":              ProcessFailed,
		"0230109999":       ProcessFailed,
	} {
		if result := byInput[input]; result.Status != want {
			t.Errorf("%s status = %s, error %v; want %s", input, result.Status, result.Err, want)
		}
	}
	if result := byInput["0130101394"]; result.PrimaryAccount != "3000010000000010" || result.BranchCode != "001" || result.OldKolekBI != 5 || result.OldKolekBPR != 2 {
		t.Errorf("alternate account result = %+v", result)
	}
	if result := byInput["3000020000000001"]; result.PrimaryAccount != result.InputAccount || result.BranchCode != "002" {
		t.Errorf("primary account changed: %+v", result)
	}
	if result := byInput["3000010000000002"]; !result.HasOld || result.Err != nil {
		t.Errorf("skipped account result = %+v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	for location, count := range logins {
		if count != 1 {
			t.Errorf("%s logins = %d, want 1", location, count)
		}
	}
	if len(logins) != 4 {
		t.Errorf("login locations = %v", logins)
	}
	if !reflect.DeepEqual(lookupAccounts, []string{"0130101394", "0230109999"}) {
		t.Errorf("alternate lookups = %v", lookupAccounts)
	}
	if !reflect.DeepEqual(inquiryAccounts, []string{
		"3000010000000010", "3000010000000002", "3000010000000005", "3000010000000006",
		"3000020000000001", "3000020000000003", "3000020000000004",
	}) {
		t.Errorf("inquiry order = %v", inquiryAccounts)
	}
	if len(updates) != 3 {
		t.Fatalf("updates = %d, want 3", len(updates))
	}
	if _, ok := updates["3000010000000005"]; !ok {
		t.Error("change type difference with same collectability did not submit update")
	}
	form := updates["3000010000000010"]
	for key, want := range map[string]string{
		"jenistransaksi":          "Update Manual Kolektibilitas BI & Internal",
		"norekening":              "3000010000000010",
		"namanasabah":             "Ratna Juwita",
		"nopk":                    "PL001000073837",
		"tgl_transaksi":           "2026-8-27",
		"total_collateralvalue":   "0",
		"total_assetvalue":        "0",
		"dpd":                     "4423",
		"nilai_kolekbilama":       "5",
		"nilai_kolekbprlama":      "2",
		"nilai_kolekbi":           "1",
		"nilai_kolekbpr":          "1",
		"jenisperubahan_kolekbi":  "Manual",
		"jenisperubahan_kolekbpr": "Manual",
		"status_dokumen":          "Diajukan",
	} {
		if got := form.Get(key); got != want {
			t.Errorf("form %s = %q, want %q", key, got, want)
		}
	}
}

func TestAutomaticChangeTypeUpdatesMatchingCollectability(t *testing.T) {
	const account = "3000010000000010"
	var form url.Values
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		switch request.URL.Path {
		case "/pinjaman/updateManualKolek/pembuatan/cari":
			fmt.Fprintf(recorder, `{"status":"ok","data":{"result":{"norekening":%q,"namanasabah":"Ratna Juwita","nopk":"PL001000073837","appdate":{"date":"2026-08-27 00:00:00.000000"},"datarekening":{"kolekbi":1,"kolekbpr":1,"updatekolekbi":"Manual","updatekolekbpr":"Manual","dpd":4423,"totalassetvalue":0,"totalcollateralvalue":0}}}}`, account)
		case "/pinjaman/updateManualKolek/pembuatan/pinjaman":
			if err := request.ParseForm(); err != nil {
				t.Error(err)
			}
			form = request.Form
			fmt.Fprint(recorder, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected endpoint %s", request.URL.Path)
			recorder.WriteHeader(http.StatusNotFound)
		}
		return recorder.Result(), nil
	})
	client, err := newClient(Config{
		BaseURL:     "http://fincloud.test",
		Credentials: fincloud.Credentials{Username: "user", Password: "password"},
		HTTPClient:  &http.Client{Transport: transport},
	}, "001")
	if err != nil {
		t.Fatal(err)
	}
	result := processLoan(context.Background(), client, LoanTarget{PrimaryAccount: account}, 1, "Automatic")
	if result.Status != ProcessSuccess || result.Err != nil {
		t.Fatalf("result = %+v, want SUCCESS", result)
	}
	for _, key := range []string{"jenisperubahan_kolekbi", "jenisperubahan_kolekbpr"} {
		if got := form.Get(key); got != "Automatic" {
			t.Errorf("%s = %q, want Automatic", key, got)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (transport roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}
