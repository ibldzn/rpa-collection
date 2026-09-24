package kolek

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
)

type LoanTarget struct {
	InputAccount   string
	PrimaryAccount string
	BranchCode     string
	inputIndex     int
}

type ProcessStatus string

const (
	ProcessSuccess ProcessStatus = "SUCCESS"
	ProcessSkipped ProcessStatus = "SKIPPED"
	ProcessFailed  ProcessStatus = "FAILED"
)

type ProcessResult struct {
	LoanTarget
	Status      ProcessStatus
	OldKolekBI  int
	OldKolekBPR int
	HasOld      bool
	NewKolek    int
	Reason      string
	Err         error
}

type Config struct {
	BaseURL          string
	Credentials      fincloud.Credentials
	LookupLocationID string
	HTTPClient       *http.Client
}

func Run(ctx context.Context, target int, changeType string, inputs []string, config Config) []ProcessResult {
	var results []ProcessResult
	var targets []LoanTarget
	var alternates []LoanTarget

	// Check every input before the first network request.
	for index, input := range inputs {
		item, alternate, err := loanInput(input)
		item.inputIndex = index
		if err != nil {
			results = append(results, failed(item, target, err))
			continue
		}
		if alternate {
			alternates = append(alternates, item)
			continue
		}
		targets = append(targets, item)
	}

	var resolver *fincloud.Client
	if len(alternates) > 0 {
		var err error
		if config.LookupLocationID == "" {
			err = errors.New("FINCLOUD_LOOKUP_LOCATION_ID is required for alternate accounts")
		} else {
			resolver, err = newClient(config, config.LookupLocationID)
			if err == nil {
				err = resolver.Login(ctx)
				if err != nil {
					resolver = nil
				}
			}
		}
		for _, item := range alternates {
			if err != nil {
				results = append(results, failed(item, target, fmt.Errorf("alternate lookup login: %w", err)))
				continue
			}
			item, err = resolveAlternate(ctx, item, resolver)
			if err != nil {
				results = append(results, failed(item, target, err))
				continue
			}
			targets = append(targets, item)
		}
	}

	branches, groups := groupByBranch(targets)
	for _, branch := range branches {
		client := resolver
		var err error
		if client == nil || branch != config.LookupLocationID {
			client, err = newClient(config, branch)
			if err == nil {
				err = client.Login(ctx)
			}
		}
		if err != nil {
			for _, item := range groups[branch] {
				results = append(results, failed(item, target, fmt.Errorf("branch %s login: %w", branch, err)))
			}
			continue
		}
		for _, item := range groups[branch] {
			results = append(results, processLoan(ctx, client, item, target, changeType))
		}
	}
	return results
}

func newClient(config Config, location string) (*fincloud.Client, error) {
	credentials := config.Credentials
	credentials.LocationID = location
	return fincloud.NewClient(credentials, fincloud.WithBaseURL(config.BaseURL), fincloud.WithHTTPClient(config.HTTPClient))
}

func processLoan(ctx context.Context, client *fincloud.Client, item LoanTarget, target int, changeType string) ProcessResult {
	result := ProcessResult{LoanTarget: item, NewKolek: target}
	inquiry, err := client.InquiryManualKolek(ctx, item.PrimaryAccount)
	if err != nil {
		result.Status, result.Err = ProcessFailed, err
		return result
	}
	result.OldKolekBI, result.OldKolekBPR, result.HasOld = inquiry.KolekBI, inquiry.KolekBPR, true
	if inquiry.KolekBI == target && inquiry.KolekBPR == target &&
		strings.EqualFold(strings.TrimSpace(inquiry.UpdateKolekBI), changeType) &&
		strings.EqualFold(strings.TrimSpace(inquiry.UpdateKolekBPR), changeType) {
		result.Status, result.Reason = ProcessSkipped, fmt.Sprintf("collectability and change type already %d/%s", target, changeType)
		return result
	}
	if err := client.SubmitManualKolek(ctx, *inquiry, target, changeType); err != nil {
		result.Status, result.Err = ProcessFailed, err
		return result
	}
	result.Status = ProcessSuccess
	return result
}

func failed(item LoanTarget, target int, err error) ProcessResult {
	return ProcessResult{LoanTarget: item, NewKolek: target, Status: ProcessFailed, Err: err}
}

func digits(value string) bool {
	if value == "" {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func branchFromPrimary(account string) (string, error) {
	if len(account) < 6 || !digits(account) {
		return "", fmt.Errorf("invalid primary loan account %q", account)
	}
	return account[3:6], nil
}

// ResolveLoanAccount accepts a primary loan account or a ten-digit alternate.
func ResolveLoanAccount(ctx context.Context, input string, config Config) (LoanTarget, error) {
	item, alternate, err := loanInput(input)
	if err != nil || !alternate {
		return item, err
	}
	if config.LookupLocationID == "" {
		return item, errors.New("FINCLOUD_LOOKUP_LOCATION_ID is required for alternate accounts")
	}
	client, err := newClient(config, config.LookupLocationID)
	if err != nil {
		return item, err
	}
	if err := client.Login(ctx); err != nil {
		return item, fmt.Errorf("alternate lookup login: %w", err)
	}
	return resolveAlternate(ctx, item, client)
}

func loanInput(input string) (LoanTarget, bool, error) {
	account := strings.TrimSpace(input)
	item := LoanTarget{InputAccount: input}
	if !digits(account) {
		return item, false, errors.New("account must contain only digits")
	}
	if len(account) == 10 {
		return item, true, nil
	}
	branch, err := branchFromPrimary(account)
	if err != nil {
		return item, false, err
	}
	item.PrimaryAccount, item.BranchCode = account, branch
	return item, false, nil
}

func resolveAlternate(ctx context.Context, item LoanTarget, client *fincloud.Client) (LoanTarget, error) {
	primary, err := client.GetLoanAccountFromAltNumber(ctx, strings.TrimSpace(item.InputAccount))
	if err != nil {
		return item, fmt.Errorf("alternate lookup: %w", err)
	}
	item.PrimaryAccount = strings.TrimSpace(primary)
	item.BranchCode, err = branchFromPrimary(item.PrimaryAccount)
	if err != nil {
		return item, fmt.Errorf("resolved account: %w", err)
	}
	return item, nil
}

func groupByBranch(targets []LoanTarget) ([]string, map[string][]LoanTarget) {
	groups := make(map[string][]LoanTarget)
	for _, item := range targets {
		groups[item.BranchCode] = append(groups[item.BranchCode], item)
	}
	branches := make([]string, 0, len(groups))
	for branch := range groups {
		branches = append(branches, branch)
		sort.SliceStable(groups[branch], func(i, j int) bool {
			return groups[branch][i].inputIndex < groups[branch][j].inputIndex
		})
	}
	sort.Strings(branches)
	return branches, groups
}
