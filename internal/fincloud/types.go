package fincloud

import "errors"

var (
	ErrMissingCredentials    = errors.New("missing Fincloud credentials")
	ErrInvalidCredentials    = errors.New("invalid Fincloud credentials")
	ErrNotLoggedIn           = errors.New("not logged in to Fincloud")
	ErrDataFetchFailed       = errors.New("failed to fetch data from Fincloud")
	ErrUnableToReauth        = errors.New("fincloud unauthorized after re-login")
	ErrDataNotFound          = errors.New("data not found")
	ErrMissingAPIClient      = errors.New("missing Fincloud API client")
	ErrInvalidSavingAccount  = errors.New("invalid saving account")
	ErrInactiveSavingAccount = errors.New("inactive saving account")
)

type AuthorizationModel struct {
	Locations []AuthLabel `json:"locationid"`
	Roles     []AuthLabel `json:"roleid"`
}

type AuthLabel struct {
	ID          string `json:"id"`
	Description string `json:"descr"`
}
