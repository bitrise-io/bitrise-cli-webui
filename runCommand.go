package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/kokomo88/bitrise-cli-webui/models"
)

var testRunning = false
var cmd *exec.Cmd
var abort = make(chan string, 1)
var randomBytes = bytes.Buffer{}
var history [][]byte

//KillGroupProcess ...
func killGroupProcess(cmd *exec.Cmd) {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		err = syscall.Kill(-pgid, 15)
		printError("syscall Kill :", err)
		sendMessage("info", "Run aborted\n")
	}
}

//running bitrise-cli
func runCommand(c *connection, workflowName string) {
	fmt.Println("Process started")
	cmd = exec.Command("bitrise-cli", "run", workflowName)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	//only for comparing data
	prevBytes := bytes.Buffer{}
	//getting random data from stdout and stderr
	randomBytes = bytes.Buffer{}
	cmd.Stdout = &randomBytes
	cmd.Stderr = &randomBytes
	testRunning = true

	err := cmd.Start()
	printError("Command start:", err)

	var dat = &models.Message{}
	dat.Type = "info"

	//reading data from command stdout every 500 Millisecond
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

	//Let wait for bitrise-cli to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	//handle error and abort
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
