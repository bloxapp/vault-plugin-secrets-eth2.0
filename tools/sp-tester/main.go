package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var config struct {
	BaseURL         string `json:"base_url"`
	AuthToken       string `json:"auth_token"`
	WalletName      string `json:"wallet_name"`
	AccountName     string `json:"account_name"`
	SignMethod      string `json:"sign_method"`
	FileFormat      string `json:"file_format"`
	ParallelRequest int    `json:"parallel_request"`
	Requests        int    `json:"requests"`
}

func main() {
	flag.StringVar(&config.BaseURL, "base-url", "https://localhost:8200", "Base URL of the vault plugin server, e.g. https://localhost:8200")
	flag.StringVar(&config.AuthToken, "auth-token", "", "Root authorization token")
	flag.StringVar(&config.WalletName, "wallet", "", "Name of a wallet")
	flag.StringVar(&config.AccountName, "account", "", "Name of an account under the specified wallet")
	flag.StringVar(&config.SignMethod, "sign-method", "sign-attestation", "Sign method name, e.g. sign-attestation")
	flag.StringVar(&config.FileFormat, "file-format", "fixtures/%d-request.json", "File path format that contains request body, prefix will be applied based on the tread number")
	flag.IntVar(&config.ParallelRequest, "parallel", 2, "Parallel requests count")
	flag.IntVar(&config.Requests, "requests", 100, "Requests count")

	flag.Parse()

	// Print received config.
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		logrus.Fatal(err)
	}

	// Create HTTP client
	httpClient := http.Client{}

	// Prepare URL.
	targetURL := fmt.Sprintf("%s/v1/ethereum/wallets/%s/accounts/%s/%s", config.BaseURL, config.WalletName, config.AccountName, config.SignMethod)

	// Prepare request bodies
	bodies := make([][]byte, config.ParallelRequest)
	for i := 1; i <= config.ParallelRequest; i++ {
		cont, err := ioutil.ReadFile(fmt.Sprintf(config.FileFormat, i))
		if err != nil {
			logrus.Fatal(err)
		}

		bodies[i-1] = cont
	}

	// Run goroutines
	var unsuccessful int
	var wg sync.WaitGroup
	for k := 0; k < config.Requests; k++ {
		for i := 0; i < config.ParallelRequest; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// Prepare request
				req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(bodies[i]))
				if err != nil {
					logrus.Fatal(err)
				}

				req.Header.Set("Authorization", "Bearer "+config.AuthToken)

				// Do request
				resp, err := httpClient.Do(req)
				if err != nil {
					logrus.Fatal(err)
				}

				// Read response body
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logrus.Println(err)
				}
				defer resp.Body.Close()

				logrus.Println(resp.Status, string(respBody))

				if resp.StatusCode != http.StatusOK {
					unsuccessful++
				}
			}(i)
		}
	}
	wg.Wait()

	logrus.Println("unsuccessful: ", unsuccessful)
}

func runRequest(httpClient *http.Client, targetURL string) {
	logrus.Println("run request", targetURL)
}
