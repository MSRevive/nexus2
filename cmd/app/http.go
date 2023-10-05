package app

import (
	"fmt"
	"net/http"
	"crypto/tls"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func (a *App) StartHTTPWithCert() error {
	cm := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.Cert.Domain),
		Cache:      autocert.DirCache("./runtime/certs"),
	}

	a.HTTPServer.TLSConfig = &tls.Config{
		GetCertificate: cm.GetCertificate,
		NextProtos:     append(srv.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
	}

	go func() {
		if err := http.ListenAndServe(":http", cm.HTTPHandler(nil)); err != nil {
			return err
		}
	}()
	
	go func() {
		logCore.Printf("Listening on: %s TLS", srv.Addr)
		if err := a.HTTPServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to serve over HTTPS: %v", err)
		}
	}()

	return err
}

func (a *App) StartHTTP() error {
	go func() {
		logCore.Printf("Listening on: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to serve over HTTP: %v", err)
		}
	}()
}