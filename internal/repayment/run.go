package repayment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/fincloudapi"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrConfig       = errors.New("unusable server configuration")
)

type Config struct {
	Kolek         kolek.Config
	APIBaseURL    string
	APISecretKey  string
	APIHTTPClient *http.Client
}

type Result struct {
	InputLoanAccount   string `json:"input_loan_account"`
	PrimaryLoanAccount string `json:"primary_loan_account"`
	Branch             string `json:"branch"`
	SavingAccount      string `json:"saving_account"`
	SavingAccountName  string `json:"saving_account_name"`
	SavingStatus       string `json:"saving_account_status"`
	Currency           string `json:"currency"`
}

func SetLoanRepaymentAccount(ctx context.Context, loanAccount, savingAccount string, config Config) (Result, error) {
	var result Result
	loanAccount, savingAccount = strings.TrimSpace(loanAccount), strings.TrimSpace(savingAccount)
	if loanAccount == "" || savingAccount == "" {
		return result, fmt.Errorf("loan_account and saving_account are required: %w", ErrInvalidInput)
	}
	if strings.TrimSpace(config.Kolek.Credentials.Username) == "" ||
		strings.TrimSpace(config.Kolek.Credentials.Password) == "" || strings.TrimSpace(config.Kolek.Credentials.RoleID) == "" ||
		strings.TrimSpace(config.APISecretKey) == "" {
		return result, ErrConfig
	}
	loan, err := kolek.ResolveLoanAccount(ctx, loanAccount, config.Kolek)
	if err != nil {
		return result, err
	}
	apiHTTPClient := config.APIHTTPClient
	if apiHTTPClient == nil {
		apiHTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	apiOptions := []fincloudapi.ClientOption{
		fincloudapi.WithBaseURL(config.APIBaseURL),
		fincloudapi.WithSecretKey(config.APISecretKey),
		fincloudapi.WithHTTPClient(apiHTTPClient),
	}
	api, err := fincloudapi.NewClient(apiOptions...)
	if err != nil {
		return result, fmt.Errorf("Fincloud API configuration: %w: %v", ErrConfig, err)
	}
	credentials := config.Kolek.Credentials
	credentials.LocationID = loan.BranchCode
	client, err := fincloud.NewClient(credentials,
		fincloud.WithBaseURL(config.Kolek.BaseURL),
		fincloud.WithHTTPClient(config.Kolek.HTTPClient),
		fincloud.WithAPIClient(api),
	)
	if err != nil {
		return result, fmt.Errorf("Fincloud Web configuration: %w: %v", ErrConfig, err)
	}
	if err := client.Login(ctx); err != nil {
		return result, fmt.Errorf("loan branch %s login: %w", loan.BranchCode, err)
	}
	saving, err := client.SetLoanRepaymentAccountDetails(ctx, loan.PrimaryAccount, savingAccount)
	if err != nil {
		return result, err
	}
	return Result{
		InputLoanAccount: loanAccount, PrimaryLoanAccount: loan.PrimaryAccount, Branch: loan.BranchCode,
		SavingAccount: saving.AccountNumber, SavingAccountName: saving.CustomerName,
		SavingStatus: saving.DocumentStatus, Currency: saving.Currency,
	}, nil
}
