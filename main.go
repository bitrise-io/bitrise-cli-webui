// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	//"bytes"
	"flag"
	"fmt"
	"net/http"
	//"os/exec"
	"log"
	"text/template"
)

var (
	addr      = flag.String("addr", ":8080", "http service address")
	homeTempl = template.Must(template.ParseFiles("home.html"))
)

// Handle http
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := homeTempl.Execute(w, r.Host)
	printError("Template Execute :", err)
}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump()
	c.sendHistory()
	c.readPump()
}

func main() {

	flag.Parse()
	go h.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	fmt.Println("Starting bitrise-cli-webui at http://localhost:8080")
	err := http.ListenAndServe(*addr, nil)
	printError("ListenAndServe:", err)
}
