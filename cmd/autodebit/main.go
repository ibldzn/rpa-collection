package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
	"github.com/ibldzn/kolek-rpa/internal/repayment"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <loan-account> <saving-account>", os.Args[0])
	}
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	for _, name := range []string{
		"FINCLOUD_USERNAME", "FINCLOUD_PASSWORD", "FINCLOUD_ROLE_ID", "FINCLOUD_API_SECRET_KEY",
	} {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			log.Fatalf("%s is required", name)
		}
	}
	config := repayment.Config{
		Kolek: kolek.Config{
			BaseURL: os.Getenv("FINCLOUD_BASE_URL"),
			Credentials: fincloud.Credentials{
				Username: os.Getenv("FINCLOUD_USERNAME"),
				Password: os.Getenv("FINCLOUD_PASSWORD"),
				RoleID:   os.Getenv("FINCLOUD_ROLE_ID"),
			},
			LookupLocationID: strings.TrimSpace(os.Getenv("FINCLOUD_LOOKUP_LOCATION_ID")),
		},
		APIBaseURL:   os.Getenv("FINCLOUD_API_BASE_URL"),
		APISecretKey: os.Getenv("FINCLOUD_API_SECRET_KEY"),
	}
	result, err := repayment.SetLoanRepaymentAccount(context.Background(), os.Args[1], os.Args[2], config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("set repayment account for %s using %s\n", result.PrimaryLoanAccount, result.SavingAccount)
}
