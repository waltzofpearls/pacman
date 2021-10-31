package main

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigTLS(t *testing.T) {
	t.Parallel()

	rootCA, err := ioutil.ReadFile("testdata/Test_Root_CA.crt")
	require.NoError(t, err)
	serverKey, err := ioutil.ReadFile("testdata/unit_test.key")
	require.NoError(t, err)
	serverCert, err := ioutil.ReadFile("testdata/unit_test.crt")
	require.NoError(t, err)

	tests := []struct {
		name            string
		givenRootCA     string
		givenServerKey  string
		givenServerCert string
		wantError       error
	}{
		{
			name:        "failed appending root CA",
			givenRootCA: "not_a_root_ca",
			wantError:   errors.New("cannot append root CA cert"),
		},
		{
			name:            "failed appending TLS key and cert",
			givenRootCA:     string(rootCA),
			givenServerKey:  "not_a_tls_key",
			givenServerCert: "not_a_tls_cert",
			wantError:       errors.New("cannot load server TLS key and cert: tls: failed to find any PEM data in certificate input"),
		},
		{
			name:            "happy path",
			givenRootCA:     string(rootCA),
			givenServerKey:  string(serverKey),
			givenServerCert: string(serverCert),
			wantError:       nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := newConfig()
			require.NoError(t, err)

			c.RootCA = tc.givenRootCA
			c.ServerKey = tc.givenServerKey
			c.ServerCert = tc.givenServerCert

			tlsConfig, err := c.tls()

			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tlsConfig)
			}
		})
	}
}
