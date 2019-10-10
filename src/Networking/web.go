package Networking

import (
	"Events"
	"Logger"
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"strings"
	"time"
)

var addr = flag.String("addr", ":443", "http service address")
var Shutdown chan bool

func SetupAdminCommands() {
	if adminCommands == nil {
		adminCommands = make(map[string]Events.Event)
		adminCommands["exit"] = &Events.Function{Name: "Admin!Exit", Function: func() { Shutdown <- true }}
		/*adminCommands["addMember"] = &Events.Function{Name: "Admin!AddMember", Function: func() {
			if adminArgs != nil {
				memberIp := adminArgs[0]
				databasing.NewMember(memberIp)
			}
		}}
		*/
	}
}
func HandleAdminCommand(msg string) bool {

	splice := strings.Split(msg, " ")
	if len(splice) == 1 {
		if command := adminCommands[msg]; command == nil {
			return false
		} else {
			Events.HandleEvent(command)
			return true
		}
	} else {
		if command := adminCommands[splice[0]]; command == nil {
			return false
		} else {
			adminArgs = splice[1:]
			Events.HandleEvent(command)
			return true
		}
	}

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	Logger.Verbose <- Logger.Msg{r.URL.String()}
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "src/Networking/home.html")
}

var onClose func()

func End() {
	if onClose != nil {
		Events.FuncEvent("Networking.End", onClose)
	}
}

func StartWebClient(toClose chan bool) {
	Shutdown = toClose
	const GET = "GET"
	SetupAdminCommands()
	setupNetworkingRegex()

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
		err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
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
