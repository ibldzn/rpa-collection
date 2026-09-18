package fincloud

import "errors"

var (
	ErrMissingCredentials = errors.New("missing Fincloud credentials")
	ErrInvalidCredentials = errors.New("invalid Fincloud credentials")
	ErrNotLoggedIn        = errors.New("not logged in to Fincloud")
	ErrDataFetchFailed    = errors.New("failed to fetch data from Fincloud")
	ErrUnableToReauth     = errors.New("unable to re-authenticate with Fincloud")
	ErrDataNotFound       = errors.New("data not found")
	ErrMissingAPIClient   = errors.New("missing Fincloud API client")
)

type AuthorizationModel struct {
	Locations []AuthLabel `json:"locationid"`
	Roles     []AuthLabel `json:"roleid"`
}

type AuthLabel struct {
	ID          string `json:"id"`
	Description string `json:"descr"`
}
