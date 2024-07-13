package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	// Timeouts
	defaultGlobalTimeout         = 30 * time.Second
	defaultDialerTimeout         = 5 * time.Second
	defaultTLSHandshakeTimeout   = 5 * time.Second
	defaultResponseHeaderTimeout = 5 * time.Second
)

// ------------------------------------------------------------------------- //
// HTTP Helpers
// ------------------------------------------------------------------------- //

type HTTPError struct {
	StatusCode int
	Body       string
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", h.StatusCode, http.StatusText(h.StatusCode))
}

func getDefaultHTTPClient() (*http.Client, error) {

	// Set up a cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: defaultGlobalTimeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: defaultDialerTimeout,
			}).DialContext,
			TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
			ResponseHeaderTimeout: defaultResponseHeaderTimeout,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
		Jar: jar,
	}, nil
}

func get(
	ctx context.Context,
	client *http.Client,
	url string,
	headers http.Header,
) ([]byte, http.Header, error) {

	// Prepare the HTTP request
	httpRequest, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, http.NoBody,
	)
	if err != nil {
		return nil, nil, err
	}
	httpRequest.Header = headers

	// Execute the HTTP Get
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = httpResponse.Body.Close()
	}()

	// Extract the response's body
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, nil, err
	}

	// Check for HTTP request failure
	if httpResponse.StatusCode != http.StatusOK {
		return nil, httpResponse.Header, &HTTPError{
			StatusCode: httpResponse.StatusCode,
			Body:       string(body),
		}
	}

	// On success, return to user
	return body, httpResponse.Header, nil
}
