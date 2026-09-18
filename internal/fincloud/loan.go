package fincloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ibldzn/fincloud-base/internal/fincloudapi"
)

func (c *Client) GetLoanAccountFromAltNumber(ctx context.Context, altNumber string) (string, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodGet, "/pinjaman/inquiry/rekening/cari", nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("cabang", "ALL")
		q.Set("noalt", altNumber)
		q.Set("pagesize", "50")
		req.URL.RawQuery = q.Encode()

		return req, nil
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrDataFetchFailed
	}

	var intermediate struct {
		Data struct {
			Result []struct {
				AccountNumber string `json:"id"`
			} `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intermediate); err != nil {
		return "", err
	}

	if intermediate.Status != "ok" {
		return "", ErrDataFetchFailed
	}

	if len(intermediate.Data.Result) != 1 {
		return "", ErrDataNotFound
	}

	return intermediate.Data.Result[0].AccountNumber, nil
}

func (c *Client) GetLoanAccountDetails(ctx context.Context, accountNumber string) (*fincloudapi.LoanInquiryResponse, error) {
	if c.api == nil {
		return nil, ErrMissingAPIClient
	}

	res, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodGet, "/pinjaman/inquiry/rekening/pinjaman", nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("id", accountNumber)
		req.URL.RawQuery = q.Encode()

		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, ErrDataFetchFailed
	}

	var intermediate struct {
		Status string `json:"status"`
		Data   struct {
			Result map[string]any `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&intermediate); err != nil {
		return nil, err
	}

	if intermediate.Status != "ok" {
		return nil, ErrDataFetchFailed
	}

	result := intermediate.Data.Result

	if _, ok := result["id"]; !ok {
		accNo, err2 := c.GetLoanAccountFromAltNumber(ctx, accountNumber)
		if err2 != nil {
			return nil, err
		}

		res2, err3 := c.GetLoanAccountDetails(ctx, accNo)
		if err3 != nil {
			return nil, err
		}

		return res2, nil
	}

	cifNo, ok := result["nocif"].(string)
	if !ok || cifNo == "" {
		return nil, ErrDataNotFound
	}

	portfolios, err := c.api.InquiryPortfolioLoanByCIFNo(ctx, cifNo)
	if err != nil {
		return nil, err
	}

	var loanDetails *fincloudapi.PortfolioLoan
	for _, portfolio := range portfolios {
		if portfolio.AccountNumber == accountNumber {
			loanDetails = &portfolio
			break
		}
	}
	if loanDetails == nil {
		return nil, ErrDataNotFound
	}

	nullableString := func(key string) string {
		if val, ok := result[key].(string); ok {
			return val
		}
		return ""
	}

	nullableDate := func(key string) string {
		if val, ok := result[key].(map[string]any); ok {
			if dateStr, ok := val["date"].(string); ok {
				return dateStr[0:10] // Extract the date part (YYYY-MM-DD)
			}
		}
		return ""
	}

	ret := &fincloudapi.LoanInquiryResponse{
		PrincipalArrears:        loanDetails.PrincipalArrears,
		SaForLoanRepayment:      nullableString("norektab_bayarangsuran2"),
		ProductName:             loanDetails.ProductName,
		StartPeriod:             loanDetails.StartPeriod,
		CurrencyCode:            loanDetails.CurrencyCode,
		InterestWriteOff:        nullableString("nilaihapusbuku_tunggakanbunga"),
		CustomerName:            loanDetails.CustomerName,
		ChargeOffDate:           nullableString("tglhapustagih"),
		TermPeriod:              loanDetails.TermPeriod,
		CreditAgreementInterest: fmt.Sprintf("%.2f", result["bungaflat"].(float64)),
		ProductID:               loanDetails.ProductCode,
		Term:                    loanDetails.Term,
		InstallmentAmount:       loanDetails.Installment,
		InterestArrears:         loanDetails.InterestArrears,
		PenaltyWriteOff:         nullableString("nilaihapusbuku_tunggakandenda"),
		Collectability:          loanDetails.Collectability,
		NextDueDate:             loanDetails.NextDueDate,
		CifNoAlt:                loanDetails.CifNoAlt,
		LoanOutStanding:         loanDetails.Outstanding,
		PenaltyArrears:          loanDetails.PenaltyArrears,
		WriteOffDate:            nullableDate("tglhapusbuku"),
		WriteOffAmount:          nullableString("totalhapusbuku"),
		Dpd:                     loanDetails.Dpd,
		CifNo:                   loanDetails.CifNo,
		CreditLimit:             loanDetails.CreditLimit,
		PrincipalWriteOff:       nullableString("nilaihapusbuku_saldopinjaman"),
		Status:                  loanDetails.Status,
		BranchCode:              loanDetails.BranchCode,
		AltNumber:               nullableString("noalt"),
		ChargeOffAmount:         nullableString("total_ht"),
		ClosedDate:              nullableString("tgltutup"),
		AccrueInterest:          loanDetails.AccrueInterest,
		EndPeriod:               loanDetails.EndPeriod,
		InterestRate:            loanDetails.InterestRate,
		LoanPrincipal:           loanDetails.LoanAmount,
		LastDueDate:             nullableDate("tgljtterakhir"),
		CompanyID:               loanDetails.CompanyID,
		AccountNumber:           loanDetails.AccountNumber,
	}

	return ret, nil
}
