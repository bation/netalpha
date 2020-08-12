package main

// import (
// 	"fmt"
// 	"html/template"
// 	"net/http"
// 	"os"
// 	"strings"

// 	"golang.org/x/net/websocket"
// )

// func liteServer() {
// 	server := http.Server{
// 		Addr: ":8769",
// 	}
// 	http.Handle("/upper", websocket.Handler(upper))
// 	http.HandleFunc("/", templateFunc)
// 	server.ListenAndServe()
// }
// func templateFunc(w http.ResponseWriter, r *http.Request) {
// 	t, err := template.ParseFiles("./template/tmpl.html")
// 	if err != nil {
// 		fmt.Fprintln(w, err)
// 		return
// 	}
// 	t.Execute(w, "hello world!")
// }
// func upper(ws *websocket.Conn) {
// 	var err error
// 	for {
// 		var reply string

// 		if err = websocket.Message.Receive(ws, &reply); err != nil {
// 			fmt.Println(err)
// 			continue
// 		}

// 		if err = websocket.Message.Send(ws, strings.ToUpper(reply)); err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 	}
// }
