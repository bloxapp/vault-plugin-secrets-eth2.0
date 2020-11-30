package httpex

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

var _ retryablehttp.Logger = &logrus.Entry{}

// Default configuration of HTTP client.
const (
	attempts        = 3
	attemptsWaitMin = time.Second
	attemptsWaitMax = time.Second * 3
	clientTimeout   = time.Minute
)

// CreateClient creates a new HTTP client.
func CreateClient(logger *logrus.Entry, errorHandler retryablehttp.ErrorHandler) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = attempts
	retryClient.RetryWaitMin = attemptsWaitMin
	retryClient.RetryWaitMax = attemptsWaitMax
	retryClient.Logger = logger
	retryClient.ErrorHandler = errorHandler

	// Override transport to support non-authorized HTTPS connections
	transport := cleanhttp.DefaultPooledTransport()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	retryClient.HTTPClient = &http.Client{
		Transport: transport,
	}

	client := &http.Client{
		Transport: &retryablehttp.RoundTripper{
			Client: retryClient,
		},
	}
	client.Timeout = clientTimeout

	return client
}
