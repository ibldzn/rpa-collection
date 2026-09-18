package fincloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) IsLoggedIn() bool {
	return c.sessionIDValue() != ""
}

func (c *Client) Login(ctx context.Context) error {
	c.reauthMu.Lock()
	defer c.reauthMu.Unlock()
	return c.login(ctx)
}

func (c *Client) login(ctx context.Context) error {
	if strings.TrimSpace(c.creds.Username) == "" || strings.TrimSpace(c.creds.Password) == "" {
		return ErrMissingCredentials
	}

	form := url.Values{}
	form.Set("locationid", c.creds.LocationID)
	form.Set("roleid", c.creds.RoleID)
	form.Set("username", c.creds.Username)
	form.Set("pwd", c.creds.Password)

	req, err := c.newRequest(ctx, http.MethodPost, "/admin/access/login", strings.NewReader(form.Encode()), false)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidCredentials
	}

	var intermediate struct {
		Data struct {
			Result struct {
				SessionID string `json:"sessionid"`
			} `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
		Error  *struct {
			System string `json:"system"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intermediate); err != nil {
		return err
	}

	if intermediate.Status != "ok" {
		if intermediate.Error != nil {
			return fmt.Errorf("%w: %s", ErrInvalidCredentials, intermediate.Error.System)
		}
		return ErrInvalidCredentials
	}

	c.setSessionID(intermediate.Data.Result.SessionID)

	return nil
}

func (c *Client) Logout(ctx context.Context) error {
	if !c.IsLoggedIn() {
		return ErrNotLoggedIn
	}

	req, err := c.NewRequest(ctx, http.MethodPost, "/admin/access/logout", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.clearSessionID()

	return nil
}

func (c *Client) GetAuthLabels(ctx context.Context) (*AuthorizationModel, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		return c.NewRequest(ctx, http.MethodGet, "/admin/access/listvalues", nil)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var intermediate struct {
		Data struct {
			Result AuthorizationModel `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intermediate); err != nil {
		return nil, err
	}

	if intermediate.Status != "ok" {
		return nil, ErrDataFetchFailed
	}

	return &intermediate.Data.Result, nil
}
