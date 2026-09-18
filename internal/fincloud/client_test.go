package fincloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoRequestReauthsAfterUnauthorized(t *testing.T) {
	ctx := context.Background()
	unauthorizedBody := &trackingReadCloser{Reader: strings.NewReader("expired")}
	var dataRequests int
	var loginRequests int

	client := newTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/admin/access/login":
			loginRequests++
			if got := req.Header.Get("sessionid"); got != "" {
				return nil, fmt.Errorf("login request sessionid = %q, want empty", got)
			}
			return loginResponse(req, "session-new"), nil
		case "/data":
			dataRequests++
			if dataRequests == 1 {
				if got := req.Header.Get("sessionid"); got != "session-old" {
					return nil, fmt.Errorf("initial request sessionid = %q, want session-old", got)
				}
				return responseWithBody(req, http.StatusUnauthorized, unauthorizedBody), nil
			}
			if got := req.Header.Get("sessionid"); got != "session-new" {
				return nil, fmt.Errorf("retry request sessionid = %q, want session-new", got)
			}
			return textResponse(req, http.StatusOK, "ok"), nil
		default:
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		}
	}))
	client.setSessionID("session-old")

	resp, err := client.DoRequest(ctx, func() (*http.Request, error) {
		return client.NewRequest(ctx, http.MethodGet, "/data", nil)
	})
	if err != nil {
		t.Fatalf("DoRequest: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if dataRequests != 2 {
		t.Fatalf("data requests = %d, want 2", dataRequests)
	}
	if loginRequests != 1 {
		t.Fatalf("login requests = %d, want 1", loginRequests)
	}
	if !unauthorizedBody.closed.Load() {
		t.Fatal("unauthorized response body was not closed")
	}
}

func TestDoRequestReturnsClearErrorWhenReloginFails(t *testing.T) {
	ctx := context.Background()
	unauthorizedBody := &trackingReadCloser{Reader: strings.NewReader("expired")}

	client := newTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/admin/access/login":
			return textResponse(req, http.StatusForbidden, "invalid"), nil
		case "/data":
			return responseWithBody(req, http.StatusUnauthorized, unauthorizedBody), nil
		default:
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		}
	}))

	resp, err := client.DoRequest(ctx, func() (*http.Request, error) {
		return client.NewRequest(ctx, http.MethodGet, "/data", nil)
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil {
		t.Fatal("response = non-nil, want nil")
	}
	if !strings.Contains(err.Error(), "fincloud session expired and re-login failed") {
		t.Fatalf("error = %q, want re-login failure context", err)
	}
	if !unauthorizedBody.closed.Load() {
		t.Fatal("unauthorized response body was not closed")
	}
}

func TestDoRequestReturnsClearErrorWhenRetryStillUnauthorized(t *testing.T) {
	ctx := context.Background()
	firstUnauthorizedBody := &trackingReadCloser{Reader: strings.NewReader("expired")}
	secondUnauthorizedBody := &trackingReadCloser{Reader: strings.NewReader("still expired")}
	var dataRequests int

	client := newTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/admin/access/login":
			return loginResponse(req, "session-new"), nil
		case "/data":
			dataRequests++
			if dataRequests == 1 {
				return responseWithBody(req, http.StatusUnauthorized, firstUnauthorizedBody), nil
			}
			return responseWithBody(req, http.StatusUnauthorized, secondUnauthorizedBody), nil
		default:
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		}
	}))

	resp, err := client.DoRequest(ctx, func() (*http.Request, error) {
		return client.NewRequest(ctx, http.MethodGet, "/data", nil)
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil {
		t.Fatal("response = non-nil, want nil")
	}
	if err.Error() != "fincloud unauthorized after re-login" {
		t.Fatalf("error = %q, want unauthorized after re-login", err)
	}
	if dataRequests != 2 {
		t.Fatalf("data requests = %d, want 2", dataRequests)
	}
	if !firstUnauthorizedBody.closed.Load() {
		t.Fatal("first unauthorized response body was not closed")
	}
	if !secondUnauthorizedBody.closed.Load() {
		t.Fatal("second unauthorized response body was not closed")
	}
}

func TestDoRequestDoesNotLoginOnSuccess(t *testing.T) {
	ctx := context.Background()
	var loginRequests int

	client := newTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/admin/access/login":
			loginRequests++
			return loginResponse(req, "session-new"), nil
		case "/data":
			return textResponse(req, http.StatusOK, "ok"), nil
		default:
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		}
	}))

	resp, err := client.DoRequest(ctx, func() (*http.Request, error) {
		return client.NewRequest(ctx, http.MethodGet, "/data", nil)
	})
	if err != nil {
		t.Fatalf("DoRequest: %v", err)
	}
	defer resp.Body.Close()

	if loginRequests != 0 {
		t.Fatalf("login requests = %d, want 0", loginRequests)
	}
}

func TestDoRequestSerializesConcurrentRelogin(t *testing.T) {
	ctx := context.Background()
	const requests = 8
	initialSeen := make(chan struct{}, requests)
	releaseInitial := make(chan struct{})
	var mu sync.Mutex
	var activeLogins int
	var maxActiveLogins int
	var loginRequests int

	client := newTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/admin/access/login":
			mu.Lock()
			activeLogins++
			if activeLogins > maxActiveLogins {
				maxActiveLogins = activeLogins
			}
			loginRequests++
			sessionID := fmt.Sprintf("session-new-%d", loginRequests)
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			activeLogins--
			mu.Unlock()
			return loginResponse(req, sessionID), nil
		case "/data":
			if req.Header.Get("sessionid") == "session-old" {
				initialSeen <- struct{}{}
				<-releaseInitial
				return textResponse(req, http.StatusUnauthorized, "expired"), nil
			}
			return textResponse(req, http.StatusOK, "ok"), nil
		default:
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		}
	}))
	client.setSessionID("session-old")

	var wg sync.WaitGroup
	errs := make(chan error, requests)
	for range requests {
		wg.Go(func() {
			resp, err := client.DoRequest(ctx, func() (*http.Request, error) {
				return client.NewRequest(ctx, http.MethodGet, "/data", nil)
			})
			if err != nil {
				errs <- err
				return
			}
			resp.Body.Close()
		})
	}

	timeout := time.After(2 * time.Second)
	for range requests {
		select {
		case <-initialSeen:
		case <-timeout:
			close(releaseInitial)
			t.Fatal("timed out waiting for initial unauthorized requests")
		}
	}
	close(releaseInitial)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("DoRequest: %v", err)
		}
	}
	if maxActiveLogins != 1 {
		t.Fatalf("max active logins = %d, want 1", maxActiveLogins)
	}
	if loginRequests != requests {
		t.Fatalf("login requests = %d, want %d", loginRequests, requests)
	}
}

func TestDoRequestRebuildsPostBodyForRetry(t *testing.T) {
	ctx := context.Background()
	const body = "payload=1"
	var dataRequests int
	var bodies []string

	client := newTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/admin/access/login":
			if req.Body != nil {
				req.Body.Close()
			}
			return loginResponse(req, "session-new"), nil
		case "/submit":
			dataRequests++
			data, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body.Close()
			bodies = append(bodies, string(data))
			if dataRequests == 1 {
				return textResponse(req, http.StatusUnauthorized, "expired"), nil
			}
			return textResponse(req, http.StatusOK, "ok"), nil
		default:
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		}
	}))

	resp, err := client.DoRequest(ctx, func() (*http.Request, error) {
		req, err := client.NewRequest(ctx, http.MethodPost, "/submit", strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
	if err != nil {
		t.Fatalf("DoRequest: %v", err)
	}
	defer resp.Body.Close()

	if len(bodies) != 2 {
		t.Fatalf("request bodies = %d, want 2", len(bodies))
	}
	for i, got := range bodies {
		if got != body {
			t.Fatalf("body %d = %q, want %q", i, got, body)
		}
	}
}

func newTestClient(t *testing.T, transport http.RoundTripper) *Client {
	t.Helper()

	client, err := NewClient(
		Credentials{Username: "user", Password: "pass", LocationID: "loc", RoleID: "role"},
		WithBaseURL("http://fincloud.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func loginResponse(req *http.Request, sessionID string) *http.Response {
	return textResponse(req, http.StatusOK, fmt.Sprintf(`{"status":"ok","data":{"result":{"sessionid":%q}}}`, sessionID))
}

func textResponse(req *http.Request, status int, body string) *http.Response {
	return responseWithBody(req, status, io.NopCloser(strings.NewReader(body)))
}

func responseWithBody(req *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       body,
		Header:     make(http.Header),
		Request:    req,
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type trackingReadCloser struct {
	*strings.Reader
	closed atomic.Bool
}

func (b *trackingReadCloser) Close() error {
	b.closed.Store(true)
	return nil
}

var _ io.ReadCloser = (*trackingReadCloser)(nil)
