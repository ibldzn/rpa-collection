package fincloudapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "http://172.22.80.18:17000"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:142.0) Gecko/20100101 Firefox/142.0"
)

var (
	ErrMissingSecretKey = errors.New("missing secret key")
	ErrAPIError         = errors.New("API error")
)

type Client struct {
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

type ClientOption func(*Client)

type APIError struct {
	ResponseCode string
	Description  string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s - %s", ErrAPIError, e.ResponseCode, e.Description)
}

func (e *APIError) Unwrap() error {
	return ErrAPIError
}

func newAPIError(resp baseResponse) error {
	return &APIError{
		ResponseCode: resp.ResponseCode,
		Description:  resp.Description,
	}
}

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

func WithSecretKey(secretKey string) ClientOption {
	return func(c *Client) {
		secretKey = strings.TrimSpace(secretKey)
		if secretKey != "" {
			c.secretKey = secretKey
		}
	}
}

func NewClient(options ...ClientOption) (*Client, error) {
	c := &Client{}
	for _, option := range options {
		option(c)
	}

	if c.secretKey == "" {
		return nil, ErrMissingSecretKey
	}

	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}

	return &Client{
		secretKey:  c.secretKey,
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var bodyReader io.Reader
	var bodyBytes []byte

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		bodyBytes = jsonBody
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Signature", generateSignature(bodyBytes, c.secretKey))
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func doAPI[T any](
	c *Client,
	ctx context.Context,
	method string,
	path string,
	body any,
	mutateReq func(*http.Request) error,
) (*T, error) {
	apiResp, err := doAPIInternal[genericResponse[T]](c, ctx, method, path, body, mutateReq)
	if err != nil || apiResp == nil {
		return nil, err
	}

	if apiResp.ResponseCode != "00" {
		return nil, newAPIError(apiResp.baseResponse)
	}

	return &apiResp.Data, nil
}

func doAPIList[T any](
	c *Client,
	ctx context.Context,
	method string,
	path string,
	body any,
	mutateReq func(*http.Request) error,
) ([]T, error) {
	apiResp, err := doAPIInternal[genericListResponse[T]](c, ctx, method, path, body, mutateReq)
	if err != nil || apiResp == nil {
		return nil, err
	}

	if apiResp.ResponseCode != "00" {
		return nil, newAPIError(apiResp.baseResponse)
	}

	return apiResp.List, nil
}

func doAPIInternal[T any](
	c *Client,
	ctx context.Context,
	method string,
	path string,
	body any,
	mutateReq func(*http.Request) error,
) (*T, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	if mutateReq != nil {
		if err := mutateReq(req); err != nil {
			return nil, err
		}
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp T
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func addNonEmptyQuery(q url.Values, key, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		q.Set(key, value)
	}
}
