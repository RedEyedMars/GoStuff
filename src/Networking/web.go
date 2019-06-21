package Networking

import (
	"Events"
	"Logger"
	"bufio"
	"context"
	"flag"
	"net/http"
	"os"
	"time"
)

var addr = flag.String("addr", ":8086", "http service address")

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

var Shutdown chan bool

func StartWebClient() <-chan bool {
	setupAdminCommands()
	setupNetworkingRegex()
	done := make(chan bool, 1)
	Shutdown = make(chan bool, 1)
	flag.Parse()
	srv := &http.Server{Addr: ":8086"}
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
	Events.GoFuncEvent("Networking.ListenForShutdown", func() {
		<-Shutdown
		err := srv.Shutdown(context.Background())
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "Networking.Shutdown"}
		done <- true
	})
	go func() {
		time.Sleep(300 * time.Second)
		Shutdown <- true
	}()

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if text[:4] == "exit" {
				Shutdown <- true
				break
			}
		}
	}()
	return done
}
