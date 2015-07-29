// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"syscall"
	"text/template"
	"time"

	bitrise "github.com/bitrise-io/bitrise-cli/bitrise"
	env "github.com/bitrise-io/stepman/models"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))
var testRunning = false
var cmd *exec.Cmd
var abort = make(chan string, 1)
var randomBytes = bytes.Buffer{}
var history [][]byte

//Workflows ...
type Workflows struct {
	FormatVersion string                 `json:"format_version" yaml:"format_version"`
	Workflows     map[string]interface{} `json:"workflows" yaml:"workflows"`
}

func mapStringInterfaceToKeyValue(e env.EnvironmentItemModel) {
	fmt.Printf("futyi")
	for key, value := range e {
		fmt.Println(" apadlova Key:", key, "Value:", value)
	}
}

func readYAMLToBytes() []byte {
	// var workflowData = &Workflows{}
	// source, err := ioutil.ReadFile("./bitrise.yml")
	// printError("File read error:", err)
	// err = yaml.Unmarshal(source, &workflowData)
	// printError("Json parse:", err)

	bitriseConfig, err := bitrise.ReadBitriseConfig("./bitrise.yml")
	if err != nil {
		fmt.Println(bitriseConfig)
	}
	bitriseConfig.Normalize()
	bitriseConfig.Validate()
	bitriseConfig.FillMissingDeafults()
	//fmt.Printf("%#v", bitriseConfig)
	// for k := range workflowData.Workflows {
	// 	wfs = append(wfs, k)
	// }
	// var message = initMessage{}
	// message.Msg = wfs
	// message.Type = "init"
	// m, err := json.Marshal(&message)
	// printError("Json encoding:", err)
	// return m
	for _, value := range bitriseConfig.App.Environments {
		mapStringInterfaceToKeyValue(value)
	}
	var message = initMessage{}
	var a []env.EnvironmentItemModel
	bitriseConfig.App.Environments = a
	message.Msg = bitriseConfig
	message.Type = "init"

	m, err := json.Marshal(&message)
	printError("Json encoding:", err)
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

func printError(from string, err error) {
	if err != nil {
		fmt.Println(from, err.Error())
	}
}

//KillGroupProcess ...
func killGroupProcess(cmd *exec.Cmd) {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, 15)
		sendMessage("info", "Run aborted\n")
	}
}

func runCommand(c *connection, workflowName string) {
	fmt.Println("Process started")
	cmd = exec.Command("bitrise-cli", "run", workflowName)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	prevBytes := bytes.Buffer{}
	randomBytes = bytes.Buffer{}
	cmd.Stdout = &randomBytes
	cmd.Stderr = &randomBytes
	testRunning = true

	err := cmd.Start()
	printError("Command start:", err)

	var dat = &Message{}
	dat.Type = "info"

	ticker := time.NewTicker(time.Millisecond * 500)
	go func(ticker *time.Ticker) {
		for _ = range ticker.C {
			r := bytes.TrimPrefix(randomBytes.Bytes(), prevBytes.Bytes())
			prevBytes = randomBytes
			history = append(history, r)
			dat.Msg = (string)(r)
			m, err := json.Marshal(&dat)
			printError("json encode:", err)
			h.broadcast <- (m)
		}
	}(ticker)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-abort:
		fmt.Println("abort")
		killGroupProcess(cmd)
		<-done
	case err := <-done:
		printError("Process error:", err)
	}
	testRunning = false
	time.Sleep(time.Second)
	ticker.Stop()
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
