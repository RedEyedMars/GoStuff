package Networking

import (
	"Events"
	"Logger"
	"context"
	"crypto/tls"
	"flag"
	"io/ioutil"
	"net/http"
	"time"
)

func StartWebClientTLS(toClose chan bool) {
	Shutdown = toClose
	const GET = "GET"
	SetupAdminCommands()
	setupNetworkingRegex()
	homeRaw, err := ioutil.ReadFile("src/Networking/home.html")
	if err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartWebClient"}
	}
	homeHtml = string(homeRaw)

	flag.Parse()
	registry := newRegistry()
	go registry.run()

	mux := http.NewServeMux()
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	imgHandler := func(imgName string) {
		mux.HandleFunc(imgName, func(w http.ResponseWriter, r *http.Request) {
			Logger.Verbose <- Logger.Msg{"Get image:" + r.URL.String()}
			if r.Method != GET {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			http.ServeFile(w, r, "assets/imgs"+r.URL.String())
		})
	}

	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(registry, w, r)
	})
	mux.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		Logger.Verbose <- Logger.Msg{"Get stylesheet:" + r.URL.String()}
		if r.Method != GET {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "src/Networking"+r.URL.String())
	})
	imgHandler("/Pending.jpg")
	imgHandler("/Fail.jpg")
	imgHandler("/Success.jpg")

	srv := &http.Server{
		Addr:         ":443",
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)}
	Events.GoFuncEvent("Networking.ListenAndServe", func() {
		err := http.ListenAndServeTLS(":443", "https-server.crt", "https-server.key", nil)
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "Networking.ListenAndServe"}
	})
	onClose = func() {
		err := srv.Shutdown(context.Background())
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "Networking.Shutdown"}
	}
	go func() {
		time.Sleep(1 * time.Hour)
		close(Shutdown)
	}()
}
