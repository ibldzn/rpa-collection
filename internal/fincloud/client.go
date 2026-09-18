package fincloud

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ibldzn/fincloud-base/internal/fincloudapi"
)

const (
	defaultBaseURL = "https://172.20.57.7/fincloud-taspen-web"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:142.0) Gecko/20100101 Firefox/142.0"
)

// Credentials bundles the username/password required to login.
type Credentials struct {
	Username   string
	Password   string
	LocationID string
	RoleID     string
}

type Client struct {
	creds      Credentials
	httpClient *http.Client
	reauthMu   sync.Mutex
	sessionMu  sync.RWMutex
	sessionID  string
	baseURL    string
	api        *fincloudapi.Client
}

type ClientOption func(*Client)

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithAPIClient(api *fincloudapi.Client) ClientOption {
	return func(c *Client) {
		if api != nil {
			c.api = api
		}
	}
}

func NewClient(creds Credentials, opts ...ClientOption) (*Client, error) {
	if creds.Username == "" || creds.Password == "" {
		return nil, ErrMissingCredentials
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client := &Client{
		creds:      creds,
		httpClient: httpClient,
		baseURL:    defaultBaseURL,
	}
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// NewRequest builds an http.Request anchored to the configured BaseURL.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	return c.newRequest(ctx, method, path, body, true)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, includeSession bool) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	if includeSession {
		if sessionID := c.sessionIDValue(); sessionID != "" {
			req.Header.Set("sessionid", sessionID)
		}
	}

	return req, nil
}

func (c *Client) DoRequest(ctx context.Context, buildReq func() (*http.Request, error)) (*http.Response, error) {
	return c.doWithReauth(ctx, buildReq)
}

func (c *Client) doWithReauth(ctx context.Context, buildReq func() (*http.Request, error)) (*http.Response, error) {
	req, err := buildReq()
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	resp.Body.Close()

	c.reauthMu.Lock()
	err = c.login(ctx)
	c.reauthMu.Unlock()
	if err != nil {
		return nil, ErrUnableToReauth
	}

	req, err = buildReq()
	if err != nil {
		return nil, err
	}

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, ErrUnableToReauth
	}

	return resp, nil
}

func (c *Client) sessionIDValue() string {
	c.sessionMu.RLock()
	defer c.sessionMu.RUnlock()
	return c.sessionID
}

func (c *Client) setSessionID(sessionID string) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	c.sessionID = sessionID
}

func (c *Client) clearSessionID() {
	c.setSessionID("")
}
