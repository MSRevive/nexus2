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
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.Config.Cert.Domain),
		Cache: autocert.DirCache("./runtime/certs"),
	}

	a.httpServer.TLSConfig = &tls.Config{
		GetCertificate: cm.GetCertificate,
		NextProtos: append(a.httpServer.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
	}

	go func() {
		if errr := http.ListenAndServe(":http", cm.HTTPHandler(nil)); errr != nil {
			err = errr
		}
	}()
	
	go func() {
		a.Logger.Info("Listening with TLS on IP", "IP", a.httpServer.Addr)

		if errr := a.httpServer.ListenAndServeTLS("", ""); errr != nil && errr != http.ErrServerClosed {
			err = fmt.Errorf("failed to serve over HTTPS: %v", err)
		}
	}()

	return
}

func (a *App) StartHTTP() (err error) {
	go func() {
		a.Logger.Info("Listening on IP", "IP", a.httpServer.Addr)

		if errr := a.httpServer.ListenAndServe(); errr != nil && errr != http.ErrServerClosed {
			err = fmt.Errorf("failed to serve over HTTP: %v", errr)
		}
	}()

	return
}