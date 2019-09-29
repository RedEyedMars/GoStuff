package Networking

import (
	"Events"
	"Logger"
	"context"
	"flag"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8080", "http service address")
var Shutdown chan bool

func setupAdminCommands() {
	adminCommands = make(map[string]Events.Event)
	adminCommands["exit"] = &Events.Function{Name: "Admin!Exit", Function: func() { Shutdown <- true }}
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

	setupAdminCommands()
	setupNetworkingRegex()

	flag.Parse()
	srv := &http.Server{Addr: ":8080"}
	registry := newRegistry()
	go registry.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(registry, w, r)
	})
	Events.GoFuncEvent("Networking.ListenAndServe", func() {
		err := http.ListenAndServe(*addr, nil)
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
