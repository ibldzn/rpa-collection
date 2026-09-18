package fincloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	manualKolekInquiryPath = "/pinjaman/updateManualKolek/pembuatan/cari"
	manualKolekUpdatePath  = "/pinjaman/updateManualKolek/pembuatan/pinjaman"
)

type ManualKolekInquiry struct {
	AccountNumber   string
	CustomerName    string
	AgreementNumber string
	TransactionDate string
	KolekBI         int
	KolekBPR        int
	DPD             int
	AssetValue      json.Number
	CollateralValue json.Number
}

func (c *Client) InquiryManualKolek(ctx context.Context, account string) (*ManualKolekInquiry, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodGet, manualKolekInquiryPath, nil)
		if err != nil {
			return nil, err
		}
		query := req.URL.Query()
		query.Set("norekening", account)
		req.URL.RawQuery = query.Encode()
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("manual kolek inquiry %s: %w", account, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manual kolek inquiry %s: HTTP %d", account, resp.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
		Data   struct {
			Result struct {
				AccountNumber   string `json:"norekening"`
				CustomerName    string `json:"namanasabah"`
				AgreementNumber string `json:"nopk"`
				AppDate         struct {
					Date string `json:"date"`
				} `json:"appdate"`
				Loan struct {
					KolekBI         *int        `json:"kolekbi"`
					KolekBPR        *int        `json:"kolekbpr"`
					DPD             *int        `json:"dpd"`
					AssetValue      json.Number `json:"totalassetvalue"`
					CollateralValue json.Number `json:"totalcollateralvalue"`
				} `json:"datarekening"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("manual kolek inquiry %s: decode: %w", account, err)
	}
	result := body.Data.Result
	if body.Status != "ok" {
		return nil, fmt.Errorf("manual kolek inquiry %s: status %q", account, body.Status)
	}
	if result.AccountNumber != account {
		return nil, fmt.Errorf("manual kolek inquiry %s: returned account %q", account, result.AccountNumber)
	}
	if result.CustomerName == "" || result.AgreementNumber == "" || result.Loan.KolekBI == nil ||
		result.Loan.KolekBPR == nil || result.Loan.DPD == nil ||
		result.Loan.AssetValue == "" || result.Loan.CollateralValue == "" {
		return nil, fmt.Errorf("manual kolek inquiry %s: missing required fields", account)
	}
	date, err := time.Parse("2006-01-02 15:04:05.999999", result.AppDate.Date)
	if err != nil {
		return nil, fmt.Errorf("manual kolek inquiry %s: invalid appdate: %w", account, err)
	}
	return &ManualKolekInquiry{
		AccountNumber:   result.AccountNumber,
		CustomerName:    result.CustomerName,
		AgreementNumber: result.AgreementNumber,
		TransactionDate: fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day()),
		KolekBI:         *result.Loan.KolekBI,
		KolekBPR:        *result.Loan.KolekBPR,
		DPD:             *result.Loan.DPD,
		AssetValue:      result.Loan.AssetValue,
		CollateralValue: result.Loan.CollateralValue,
	}, nil
}

func (c *Client) SubmitManualKolek(ctx context.Context, inquiry ManualKolekInquiry, target int) error {
	form := url.Values{
		"jenistransaksi":          {"Update Manual Kolektibilitas BI & Internal"},
		"norekening":              {inquiry.AccountNumber},
		"namanasabah":             {inquiry.CustomerName},
		"nopk":                    {inquiry.AgreementNumber},
		"tgl_transaksi":           {inquiry.TransactionDate},
		"total_collateralvalue":   {inquiry.CollateralValue.String()},
		"total_assetvalue":        {inquiry.AssetValue.String()},
		"dpd":                     {strconv.Itoa(inquiry.DPD)},
		"nilai_kolekbilama":       {strconv.Itoa(inquiry.KolekBI)},
		"nilai_kolekbprlama":      {strconv.Itoa(inquiry.KolekBPR)},
		"nilai_kolekbi":           {strconv.Itoa(target)},
		"nilai_kolekbpr":          {strconv.Itoa(target)},
		"jenisperubahan_kolekbi":  {"Automatic"},
		"jenisperubahan_kolekbpr": {"Automatic"},
		"status_dokumen":          {"Diajukan"},
	}
	encoded := form.Encode()
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodPost, manualKolekUpdatePath, strings.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
	if err != nil {
		return fmt.Errorf("manual kolek update %s: %w", inquiry.AccountNumber, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("manual kolek update %s: HTTP %d", inquiry.AccountNumber, resp.StatusCode)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("manual kolek update %s: decode: %w", inquiry.AccountNumber, err)
	}
	if body.Status != "ok" {
		return fmt.Errorf("manual kolek update %s: status %q", inquiry.AccountNumber, body.Status)
	}
	return nil
}
