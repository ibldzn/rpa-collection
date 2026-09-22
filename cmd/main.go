package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
	"github.com/joho/godotenv"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, output, errors io.Writer) int {
	if len(args) < 3 || len(args[0]) != 1 || args[0][0] < '1' || args[0][0] > '5' {
		fmt.Fprintln(errors, "usage: app <kolek: 1-5> <Manual|Automatic> <loan-account> [loan-account...]")
		return 2
	}
	var changeType string
	switch {
	case strings.EqualFold(strings.TrimSpace(args[1]), "Manual"):
		changeType = "Manual"
	case strings.EqualFold(strings.TrimSpace(args[1]), "Automatic"):
		changeType = "Automatic"
	default:
		fmt.Fprintln(errors, "change type must be Manual or Automatic")
		return 2
	}
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(errors, "load .env: %v\n", err)
		return 2
	}
	config, err := configFromEnv()
	if err != nil {
		fmt.Fprintln(errors, err)
		return 2
	}
	target := int(args[0][0] - '0')
	results := kolek.Run(context.Background(), target, changeType, args[2:], config)
	var success, skipped, failed int
	for _, result := range results {
		branch := result.BranchCode
		if branch == "" {
			branch = "---"
		}
		account := result.InputAccount
		if result.PrimaryAccount != "" && strings.TrimSpace(result.InputAccount) != result.PrimaryAccount {
			account += " -> " + result.PrimaryAccount
		}
		fmt.Fprintf(output, "[%s] %s", branch, account)
		if result.HasOld {
			fmt.Fprintf(output, " | BI %d -> %d | BPR %d -> %d", result.OldKolekBI, target, result.OldKolekBPR, target)
		} else {
			fmt.Fprintf(output, " | target %d", target)
		}
		fmt.Fprintf(output, " | change %s | %s", changeType, result.Status)
		if result.Err != nil {
			fmt.Fprintf(output, ": %v", result.Err)
		} else if result.Reason != "" {
			fmt.Fprintf(output, ": %s", result.Reason)
		}
		fmt.Fprintln(output)
		switch result.Status {
		case kolek.ProcessSuccess:
			success++
		case kolek.ProcessSkipped:
			skipped++
		case kolek.ProcessFailed:
			failed++
		}
	}
	fmt.Fprintf(output, "Processed: %d\nSuccess:   %d\nSkipped:   %d\nFailed:    %d\n", len(results), success, skipped, failed)
	if failed > 0 {
		return 1
	}
	return 0
}

func configFromEnv() (kolek.Config, error) {
	config := kolek.Config{
		BaseURL: os.Getenv("FINCLOUD_BASE_URL"),
		Credentials: fincloud.Credentials{
			Username: os.Getenv("FINCLOUD_USERNAME"),
			Password: os.Getenv("FINCLOUD_PASSWORD"),
			RoleID:   os.Getenv("FINCLOUD_ROLE_ID"),
		},
		LookupLocationID: strings.TrimSpace(os.Getenv("FINCLOUD_LOOKUP_LOCATION_ID")),
	}
	for _, name := range []string{"FINCLOUD_BASE_URL", "FINCLOUD_USERNAME", "FINCLOUD_PASSWORD", "FINCLOUD_ROLE_ID"} {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			return config, fmt.Errorf("%s is required", name)
		}
	}
	return config, nil
}
