package main

import "log"
import "code.google.com/p/go.net/websocket"
import "net/http"
import "minispace"
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
			log.Fatal("Receive err %v\n", err)
			break
		}
		if err = websocket.Message.Send(ws, &reply); err != nil {
			log.Fatal("Send err %v\n", err)
			break
		}
		log.Printf("Send and recv %s\n", reply);
	}
}

func clientProc(ws *websocket.Conn) {
	log.Printf("recv client %v\n", *ws);
	client := minispace.NewClient(ws)
	client.Proc()
}

func main() {
	log.Printf("Starting minispace...");
	err := minispace.Init()
	if err != nil {
		log.Printf("minispace init failed\n")
		return
	}

	http.Handle("/echo", websocket.Handler(echo));
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
