package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
	"github.com/ibldzn/kolek-rpa/internal/repayment"
)

type Server struct {
	Key                 string
	UpdateKolek         func(context.Context, int, string, []string) []kolek.ProcessResult
	SetRepaymentAccount func(context.Context, string, string) (repayment.Result, error)
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/kolek", s.auth(s.kolek))
	mux.HandleFunc("/api/v1/repayment-account", s.auth(s.repaymentAccount))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { writeError(w, http.StatusNotFound, "not found") })
	return mux
}

func (s Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authorization, "Bearer ")
		if !ok || s.Key == "" || subtle.ConstantTimeCompare([]byte(token), []byte(s.Key)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func (s Server) kolek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request struct {
		Kolek      int      `json:"kolek"`
		ChangeType string   `json:"change_type"`
		Accounts   []string `json:"accounts"`
	}
	if !decode(w, r, &request) {
		return
	}
	if request.Kolek < 1 || request.Kolek > 5 {
		writeError(w, http.StatusBadRequest, "kolek must be 1..5")
		return
	}
	if request.ChangeType != "Manual" && request.ChangeType != "Automatic" {
		writeError(w, http.StatusBadRequest, "change_type must be Manual or Automatic")
		return
	}
	if len(request.Accounts) == 0 {
		writeError(w, http.StatusBadRequest, "accounts must be non-empty")
		return
	}
	if s.UpdateKolek == nil {
		writeError(w, http.StatusInternalServerError, "server configuration unavailable")
		return
	}
	started := time.Now()
	results := s.UpdateKolek(r.Context(), request.Kolek, request.ChangeType, request.Accounts)
	data := struct {
		Processed int           `json:"processed"`
		Success   int           `json:"success"`
		Skipped   int           `json:"skipped"`
		Failed    int           `json:"failed"`
		Results   []kolekResult `json:"results"`
	}{Results: make([]kolekResult, 0, len(results))}
	for _, item := range results {
		log.Printf("operation=kolek input_account=%q primary_account=%q branch=%q target=%d change_type=%s result=%s", item.InputAccount, item.PrimaryAccount, item.BranchCode, request.Kolek, request.ChangeType, item.Status)
		entry := kolekResult{InputAccount: item.InputAccount, PrimaryAccount: item.PrimaryAccount, Branch: item.BranchCode,
			NewKolek: item.NewKolek, Status: item.Status, Reason: item.Reason}
		if item.HasOld {
			entry.OldKolekBI, entry.OldKolekBPR = &item.OldKolekBI, &item.OldKolekBPR
		}
		if item.Err != nil {
			entry.Error = safeKolekError(item.Err)
		}
		data.Results = append(data.Results, entry)
		switch item.Status {
		case kolek.ProcessSuccess:
			data.Success++
		case kolek.ProcessSkipped:
			data.Skipped++
		case kolek.ProcessFailed:
			data.Failed++
		}
	}
	data.Processed = len(data.Results)
	log.Printf("operation=kolek target=%d change_type=%s processed=%d success=%d skipped=%d failed=%d duration=%s", request.Kolek, request.ChangeType, data.Processed, data.Success, data.Skipped, data.Failed, time.Since(started))
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "data": data})
}

type kolekResult struct {
	InputAccount   string              `json:"input_account"`
	PrimaryAccount string              `json:"primary_account,omitempty"`
	Branch         string              `json:"branch,omitempty"`
	OldKolekBI     *int                `json:"old_kolek_bi,omitempty"`
	OldKolekBPR    *int                `json:"old_kolek_bpr,omitempty"`
	NewKolek       int                 `json:"new_kolek"`
	Status         kolek.ProcessStatus `json:"status"`
	Reason         string              `json:"reason,omitempty"`
	Error          string              `json:"error,omitempty"`
}

func safeKolekError(err error) string {
	if errors.Is(err, fincloud.ErrDataNotFound) {
		return "alternate lookup: data not found"
	}
	if errors.Is(err, fincloud.ErrInvalidCredentials) {
		return "Fincloud login failed"
	}
	if strings.Contains(err.Error(), "account must contain only digits") || strings.Contains(err.Error(), "invalid primary loan account") ||
		strings.Contains(err.Error(), "FINCLOUD_LOOKUP_LOCATION_ID is required") {
		return err.Error()
	}
	return "Fincloud operation failed"
}

func (s Server) repaymentAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request struct {
		LoanAccount   string `json:"loan_account"`
		SavingAccount string `json:"saving_account"`
	}
	if !decode(w, r, &request) {
		return
	}
	request.LoanAccount, request.SavingAccount = strings.TrimSpace(request.LoanAccount), strings.TrimSpace(request.SavingAccount)
	if request.LoanAccount == "" || request.SavingAccount == "" {
		writeError(w, http.StatusBadRequest, "loan_account and saving_account are required")
		return
	}
	if s.SetRepaymentAccount == nil {
		writeError(w, http.StatusInternalServerError, "server configuration unavailable")
		return
	}
	started := time.Now()
	result, err := s.SetRepaymentAccount(r.Context(), request.LoanAccount, request.SavingAccount)
	if err != nil {
		status, message := repaymentError(err)
		log.Printf("operation=repayment_account input_loan_account=%q saving_account=%q result=failed duration=%s", request.LoanAccount, request.SavingAccount, time.Since(started))
		writeError(w, status, message)
		return
	}
	log.Printf("operation=repayment_account input_loan_account=%q primary_loan_account=%q branch=%q saving_account=%q result=success duration=%s", result.InputLoanAccount, result.PrimaryLoanAccount, result.Branch, result.SavingAccount, time.Since(started))
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "data": result})
}

func repaymentError(err error) (int, string) {
	switch {
	case errors.Is(err, repayment.ErrInvalidInput):
		return http.StatusBadRequest, "invalid loan or saving account"
	case errors.Is(err, repayment.ErrConfig), errors.Is(err, fincloud.ErrMissingAPIClient):
		return http.StatusInternalServerError, "server configuration unavailable"
	case errors.Is(err, fincloud.ErrInactiveSavingAccount):
		return http.StatusUnprocessableEntity, "saving account is not active"
	case errors.Is(err, fincloud.ErrInvalidSavingAccount):
		return http.StatusUnprocessableEntity, "saving account inquiry did not return a valid account"
	case errors.Is(err, fincloud.ErrDataNotFound):
		return http.StatusUnprocessableEntity, "loan or saving account not found"
	case strings.Contains(err.Error(), "account must contain only digits"), strings.Contains(err.Error(), "invalid primary loan account"):
		return http.StatusBadRequest, "invalid loan account"
	case strings.Contains(err.Error(), "FINCLOUD_LOOKUP_LOCATION_ID is required"):
		return http.StatusInternalServerError, "server configuration unavailable"
	case strings.Contains(err.Error(), "alternate lookup"):
		return http.StatusBadGateway, "Fincloud alternate loan lookup failed"
	case strings.Contains(err.Error(), "loan inquiry"):
		return http.StatusBadGateway, "Fincloud loan inquiry failed"
	case strings.Contains(err.Error(), "saving account inquiry"):
		return http.StatusBadGateway, "Fincloud saving account inquiry failed"
	case strings.Contains(err.Error(), "submit"):
		return http.StatusBadGateway, "Fincloud repayment update failed"
	case strings.Contains(err.Error(), "login"):
		return http.StatusBadGateway, "Fincloud login failed"
	default:
		return http.StatusBadGateway, "Fincloud operation failed"
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"status": "error", "error": map[string]string{"message": message}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
