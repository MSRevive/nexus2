package app

import (
	"fmt"
	"net/http"
	"crypto/tls"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func (a *App) StartHTTPWithCert() (err error) {
	cm := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.Config.Cert.Domain),
		Cache:      autocert.DirCache("./runtime/certs"),
	}

	a.HTTPServer.TLSConfig = &tls.Config{
		GetCertificate: cm.GetCertificate,
		NextProtos:     append(a.HTTPServer.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
	}

	go func() {
		if errr := http.ListenAndServe(":http", cm.HTTPHandler(nil)); errr != nil {
			err = errr
		}
	}()
	
	go func() {
		a.Logger.Core.Info("Listening with TLS", "IP", a.HTTPServer.Addr)

		if errr := a.HTTPServer.ListenAndServeTLS("", ""); errr != nil && errr != http.ErrServerClosed {
			err = fmt.Errorf("failed to serve over HTTPS: %v", err)
		}
	}()

	return
}

func (a *App) StartHTTP() (err error) {
	go func() {
		a.Logger.Core.Info("Listening", "IP", a.HTTPServer.Addr)

		if errr := a.HTTPServer.ListenAndServe(); errr != nil && errr != http.ErrServerClosed {
			err = fmt.Errorf("failed to serve over HTTP: %v", errr)
		}
	}()

	return
}