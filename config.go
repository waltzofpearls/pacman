package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Listen     string `default:":9000"`
	UseMTLS    bool   `envconfig:"USE_MTLS" default:"true"`
	RootCA     string `envconfig:"TLS_ROOT_CA"`
	ServerKey  string `envconfig:"TLS_SERVER_KEY"`
	ServerCert string `envconfig:"TLS_SERVER_CERT"`
}

func newConfig() (*config, error) {
	var conf config
	err := envconfig.Process("", &conf)
	return &conf, err
}

func (c *config) tls() (*tls.Config, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(c.RootCA)) {
		return nil, errors.New("cannot append root CA cert")
	}
	certificate, err := tls.X509KeyPair([]byte(c.ServerCert), []byte(c.ServerKey))
	if err != nil {
		return nil, fmt.Errorf("cannot load server TLS key and cert: %s", err)
	}
	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}, nil
}
