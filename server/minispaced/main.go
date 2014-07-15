package main

import "log"
import "code.google.com/p/go.net/websocket"
import "net/http"
import "github.com/lijie/minispace/server/minispace"
import "runtime"
import "flag"
import "fmt"
//import "code.google.com/p/log4go"

//var mainlog log4go.Logger
//func init() {
//	mainlog = make(log4go.Logger)
//	// log.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter())
//	mainlog.AddFilter("log", log4go.DEBUG, log4go.NewFileLogWriter("../log/main.log", true).SetRotateDaily(true))
//	mainlog.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
//}

func echo(ws *websocket.Conn) {
	var err error
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Printf("Receive err %v\n", err)
			break
		}
		fmt.Printf("recv %s\n", reply);
		if err = websocket.Message.Send(ws, reply); err != nil {
			fmt.Printf("Send err %v\n", err)
			break
		}
	}
}

func clientProc(ws *websocket.Conn) {
	log.Printf("recv client %v\n", *ws);
	client := minispace.NewClient(ws)
	client.Proc()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Printf("Starting minispace...");

	nodb := flag.Bool("nodb", false, "running without db")
	noai := flag.Bool("noai", false, "running without AI")
	flag.Parse()

	config := minispace.NewMiniConfig()
	config.EnableDB = !*nodb
	config.EnableAI = !*noai
	fmt.Printf("enable db %v\n", config.EnableDB)
	fmt.Printf("enable ai %v\n", config.EnableAI)

	err := minispace.Init(config)
	if err != nil {
		log.Printf("minispace init failed\n")
		return
	}

	echoproto := websocket.Server{
		Handshake: nil,
		Handler: websocket.Handler(echo),
	}
	http.Handle("/echo", echoproto);

	myproto := websocket.Server{
		Handshake: nil,
		Handler: websocket.Handler(clientProc),
	}
	http.Handle("/minispace", myproto)
	if err := http.ListenAndServe(":12345", nil); err != nil {
		log.Fatal("ListenAndServe err %v", err)
	}
	log.Printf("Stoping minispace...");
}
