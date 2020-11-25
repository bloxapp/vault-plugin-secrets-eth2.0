package httpex

import (
	"net/http"
	"time"

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

	client := retryClient.StandardClient()
	client.Timeout = clientTimeout

	return client
}
