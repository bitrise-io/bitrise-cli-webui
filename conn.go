// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kokomo88/bitrise-cli-webui/models"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024000,
	WriteBufferSize: 1024000,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func sendMessage(Type string, msg string) {
	var m = &models.Message{}
	m.Type = Type
	m.Msg = msg
	byteArr, err := json.Marshal(m)
	printError("Json encoding:", err)
	h.broadcast <- byteArr
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		err := c.ws.Close()
		printError("Websocket Close:", err)
	}()
	c.ws.SetReadLimit(maxMessageSize)
	err := c.ws.SetReadDeadline(time.Now().Add(pongWait))
	printError("Websocket SetReadDeadline:", err)
	c.ws.SetPongHandler(func(string) error {
		err = c.ws.SetReadDeadline(time.Now().Add(pongWait))
		printError("Websocket SetReadDeadline:", err)
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		//Handle incomming messages
		var dat = &models.Message{}
		err = json.Unmarshal(message, &dat)
		printError("Json encoding :", err)
		if dat.Type == "init" {
			message = readYAMLToBytes()
			h.broadcast <- message
		} else if dat.Type == "build" && !testRunning {
			go runCommand(c, dat.Msg)
			sendMessage("info", "$bitrise-cli run "+dat.Msg+"\n")
		} else if dat.Type == "save" {
			var data = &models.SaveMessage{}
			err = json.Unmarshal(message, &data)
			printError("Json encoding :", err)
			saveConfig(data.Msg)
			// fmt.Printf("%#v", dat.Msg)
		} else if dat.Type == "abort" && testRunning {
			abort <- "Aborting build"
			sendMessage("info", "Aborting build\n")
		}
	}

}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	err := c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	printError("Websocket SetWriteDeadline :", err)
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.ws.Close()
		printError("Websocket close :", err)
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				err := c.write(websocket.CloseMessage, []byte{})
				printError("Websocket write :", err)
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *connection) sendHistory() {
	for _, val := range history {
		//c.send <- val
		sendMessage("info", (string)(val))
		time.Sleep(time.Millisecond * 1)
	}
}
