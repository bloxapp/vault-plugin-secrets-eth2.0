package endpoint

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// BasePath is the base path for all endpoints.
	BasePath = "/v1/ethereum"
)

// Build builds full path.
func Build(network, pattern string) (string, error) {
	if len(network) > 0 {
		return fmt.Sprintf("%s/%s/%s", BasePath, network, pattern), nil
	}

	return "", errors.New("netowrk is not defined")
}
