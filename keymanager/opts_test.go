package keymanager

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalConfigFile(t *testing.T) {
	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "invalid JSON string",
			args: args{
				r: ioutil.NopCloser(strings.NewReader(`invalid`)),
			},
			wantErr: true,
		},
		{
			name: "valid JSON string",
			args: args{
				r: ioutil.NopCloser(strings.NewReader(`{"location":"location","access_token":"access_token","public_key":"public_key","network":"network"}`)),
			},
			want: &Config{
				Location:    "location",
				AccessToken: "access_token",
				PubKey:      "public_key",
				Network:     "network",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalConfigFile(tt.args.r)
			require.Equal(t, tt.want, got, got)
			require.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}
