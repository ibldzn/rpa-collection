package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/fincloudapi"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <loan-account> <oper-account>", os.Args[0])
	}
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	for _, name := range []string{
		"FINCLOUD_USERNAME",
		"FINCLOUD_PASSWORD",
		"FINCLOUD_LOCATION_ID",
		"FINCLOUD_ROLE_ID",
		"FINCLOUD_API_SECRET_KEY",
	} {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			log.Fatalf("%s is required", name)
		}
	}

	api, err := fincloudapi.NewClient(
		fincloudapi.WithBaseURL(os.Getenv("FINCLOUD_API_BASE_URL")),
		fincloudapi.WithSecretKey(os.Getenv("FINCLOUD_API_SECRET_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}
	client, err := fincloud.NewClient(
		fincloud.Credentials{
			Username:   os.Getenv("FINCLOUD_USERNAME"),
			Password:   os.Getenv("FINCLOUD_PASSWORD"),
			LocationID: os.Getenv("FINCLOUD_LOCATION_ID"),
			RoleID:     os.Getenv("FINCLOUD_ROLE_ID"),
		},
		fincloud.WithBaseURL(os.Getenv("FINCLOUD_BASE_URL")),
		fincloud.WithAPIClient(api),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := client.Login(ctx); err != nil {
		log.Fatal(err)
	}
	if err := client.RegisterAutodebitRemoval(ctx, os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("registered autodebit removal for %s using %s\n", os.Args[1], os.Args[2])
}
