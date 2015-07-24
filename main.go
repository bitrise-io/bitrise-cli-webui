// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	// "github.com/gorilla/websocket"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))
var testRunning = false
var cmd *exec.Cmd
var abort = make(chan string, 1)
var randomBytes = bytes.Buffer{}
var history [][]byte

var workflowData = &Workflows{}

//Workflows ...
type Workflows struct {
	FormatVersion string                 `json:"format_version" yaml:"format_version"`
	Workflows     map[string]interface{} `json:"workflows" yaml:"workflows"`
}

func readYAML() []byte {
	source, err := ioutil.ReadFile("./yml/bitrise.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &workflowData)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%#v", workflowData)
	var wfs []string

	for k := range workflowData.Workflows {
		fmt.Println(len(wfs))
		wfs = append(wfs, k)
	}
	var message = Message{}
	message.Msg = wfs
	message.Type = "init"
	m, err := json.Marshal(&message)
	return m
}
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("%s\n", string(outs))
	}
}

func runCommand(c *connection) {
	fmt.Println("Process started")
	cmd = exec.Command("go", "run", "./command/test.go")
	//cmd = exec.Command("sleep", "30")
	prevBytes := bytes.Buffer{}
	randomBytes = bytes.Buffer{}
	//allBytes := &bytes.Buffer{}
	cmd.Stdout = &randomBytes

	// Start 	command asynchronously
	err := cmd.Start()
	printError(err)
	fmt.Println(cmd.Process.Pid)
	// Create a ticker that outputs elapsed time
	ticker := time.NewTicker(time.Millisecond * 500)
	go func(ticker *time.Ticker) {
		for _ = range ticker.C {
			r := bytes.TrimPrefix(randomBytes.Bytes(), prevBytes.Bytes())
			prevBytes = randomBytes
			//printOutput(r)
			// if err := c.write(websocket.TextMessage, (r)); err != nil {
			// 	return
			// }
			history = append(history, r)
			h.broadcast <- (r)
		}
	}(ticker)

	// Only proceed once the process has finished
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-abort:
		fmt.Println("abort")
		if err := cmd.Process.Signal(os.Kill); err != nil {
			log.Fatal("failed to kill: ", err)
		}
		<-done // allow goroutine to exit
	case err := <-done:
		if err != nil {
			log.Printf("process done with error = %v", err)
		}
	}

	testRunning = false
	fmt.Println("Process finished")
	time.Sleep(time.Second)
	ticker.Stop()
}

func main() {

	flag.Parse()
	go h.run()

	fileServer := http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules")))
	http.Handle("/node_modules/", fileServer)
	fileServer = http.StripPrefix("/yml/", http.FileServer(http.Dir("yml")))
	http.Handle("/yml/", fileServer)

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
