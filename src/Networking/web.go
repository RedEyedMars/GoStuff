package Networking

import (
	"Events"
	"Logger"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const GET = "GET"

var addr = flag.String("addr", ":8080", "http service address")
var Shutdown chan bool
var homeHtml string

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
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, homeHtml)
	//http.ServeFile(w, r, "src/Networking/home.html")
}

func HandleImg(imgName string) {
	http.HandleFunc(imgName, func(w http.ResponseWriter, r *http.Request) {
		Logger.Verbose <- Logger.Msg{"Get image:" + r.URL.String()}
		if r.Method != GET {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "assets/imgs"+r.URL.String())
	})
}
func HandleJs(libName string) {
	http.HandleFunc(libName, func(w http.ResponseWriter, r *http.Request) {
		Logger.Verbose <- Logger.Msg{"Get js:" + r.URL.String()}
		if r.Method != GET {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "src/www/js"+r.URL.String())
	})
}

var onClose func()

func End() {
	if onClose != nil {
		Events.FuncEvent("Networking.End", onClose)
	}
}

func StartWebClient(toClose chan bool) {
	Shutdown = toClose
	SetupAdminCommands()
	setupNetworkingRegex()
	homeRaw, err := ioutil.ReadFile("src/www/home.html")
	if err != nil {
		Logger.Error <- Logger.ErrMsg{Err: err, Status: "StartWebClient"}
	}
	homeHtml = string(homeRaw)

	flag.Parse()
	registry := newRegistry()
	go registry.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(registry, w, r)
	})
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		Logger.Verbose <- Logger.Msg{"Get stylesheet:" + r.URL.String()}
		if r.Method != GET {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "src/www"+r.URL.String())
	})
	http.HandleFunc("/forge-sha256.min.js", func(w http.ResponseWriter, r *http.Request) {
		Logger.Verbose <- Logger.Msg{"Get sha256:" + r.URL.String()}
		if r.Method != GET {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "www/lib/forge-sha256-master/build/forge-sha256.min.js")
	})
	HandleJs("/chat.js")
	HandleJs("/login.js")
	HandleImg("/Pending.jpg")
	HandleImg("/Fail.jpg")
	HandleImg("/Success.jpg")

	srv := &http.Server{Addr: ":8080"}
	Events.GoFuncEvent("Networking.ListenAndServe", func() {
		err := http.ListenAndServe(":8080", nil)
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
