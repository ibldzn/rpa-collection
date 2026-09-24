package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ibldzn/kolek-rpa/internal/api"
	"github.com/ibldzn/kolek-rpa/internal/fincloud"
	"github.com/ibldzn/kolek-rpa/internal/kolek"
	"github.com/ibldzn/kolek-rpa/internal/repayment"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	for _, name := range []string{
		"RPA_API_KEY", "FINCLOUD_BASE_URL", "FINCLOUD_API_BASE_URL", "FINCLOUD_LOOKUP_LOCATION_ID", "FINCLOUD_USERNAME",
		"FINCLOUD_PASSWORD", "FINCLOUD_ROLE_ID", "FINCLOUD_API_SECRET_KEY",
	} {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			log.Fatalf("%s is required", name)
		}
	}
	listen := strings.TrimSpace(os.Getenv("RPA_LISTEN_ADDR"))
	if listen == "" {
		listen = ":8080"
	}
	kolekConfig := kolek.Config{
		BaseURL: os.Getenv("FINCLOUD_BASE_URL"),
		Credentials: fincloud.Credentials{
			Username: os.Getenv("FINCLOUD_USERNAME"),
			Password: os.Getenv("FINCLOUD_PASSWORD"),
			RoleID:   os.Getenv("FINCLOUD_ROLE_ID"),
		},
		LookupLocationID: strings.TrimSpace(os.Getenv("FINCLOUD_LOOKUP_LOCATION_ID")),
	}
	repaymentConfig := repayment.Config{
		Kolek:        kolekConfig,
		APIBaseURL:   os.Getenv("FINCLOUD_API_BASE_URL"),
		APISecretKey: os.Getenv("FINCLOUD_API_SECRET_KEY"),
	}
	handler := api.Server{
		Key: os.Getenv("RPA_API_KEY"),
		UpdateKolek: func(ctx context.Context, target int, changeType string, accounts []string) []kolek.ProcessResult {
			return kolek.Run(ctx, target, changeType, accounts, kolekConfig)
		},
		SetRepaymentAccount: func(ctx context.Context, loan, saving string) (repayment.Result, error) {
			return repayment.SetLoanRepaymentAccount(ctx, loan, saving, repaymentConfig)
		},
	}.Handler()
	server := &http.Server{
		Addr: listen, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 10 * time.Minute, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	log.Printf("RPA API listening on %s", listen)
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
		if err := <-done; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server: %v", err)
		}
	}
}
