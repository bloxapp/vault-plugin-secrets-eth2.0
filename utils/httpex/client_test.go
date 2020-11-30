package httpex

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCreateClient(t *testing.T) {
	t.Run("accepts untrusted HTTPS connection", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		}))
		defer srv.Close()

		req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
		require.NoError(t, err)

		client := CreateClient(logrus.NewEntry(logrus.New()), nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("accepts HTTP connection", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		}))
		defer srv.Close()

		req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
		require.NoError(t, err)

		client := CreateClient(logrus.NewEntry(logrus.New()), nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
