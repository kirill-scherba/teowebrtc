// Copyright 2021 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Webretc signal server client (for teonet network)
package teowebrtc_signal_client

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"nhooyr.io/websocket"
)

// New signal server client
func New() *SignalClient {
	return new(SignalClient)
}

type SignalClient struct {
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

type Login struct {
	Signal string `json:"signal"`
	Login  string `json:"login"`
}

type Signal struct {
	Signal string      `json:"signal"`
	Peer   string      `json:"peer"`
	Data   interface{} `json:"data"`
}

// Connect to signal server and send login signal
func (cli *SignalClient) Connect(signalServerAddr, peerLogin string) (err error) {
	u := url.URL{Scheme: "ws", Host: signalServerAddr, Path: "/signal"}
	log.Printf("Connecting to %s\n", u.String())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	c, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	cli.conn = c
	cli.ctx = ctx
	cli.cancel = cancel

	// Send login signal
	var login = Login{"login", peerLogin}
	d, err := json.Marshal(login)
	if err != nil {
		log.Println("Login marshal:", err)
		return
	}
	err = c.Write(ctx, websocket.MessageText, d)
	if err != nil {
		log.Println("Write message error:", err)
		return
	}

	cli.waitAnswer()

	return
}

// Close connection to signal server
func (cli *SignalClient) Close() {
	log.Println("Sinal client closed")
	cli.conn.Close(websocket.StatusNormalClosure, "done")
	cli.cancel()
}

// WaitOffer wait offer signal received
func (cli SignalClient) WaitOffer() (sig Signal, err error) {
	message, err := cli.waitAnswer()
	if err != nil {
		return
	}
	json.Unmarshal(message, &sig)
	return
}

// WaitCandidate wait candidate signal received
func (cli SignalClient) WaitCandidate() (sig Signal, err error) {
	message, err := cli.waitAnswer()
	if err != nil {
		return
	}
	json.Unmarshal(message, &sig)
	return
}

// waitAnswer waite message received
func (cli SignalClient) waitAnswer() (message []byte, err error) {
	_, message, err = cli.conn.Read(cli.ctx)
	if err != nil {
		log.Println("Read message error:", err)
		return
	}
	return
}

// WriteOffer send offer signal
func (cli SignalClient) WriteOffer(peer string, offer []byte) (answer []byte, err error) {
	log.Printf("Send offer to %s", peer)
	err = cli.writeSignal("offer", peer, offer)
	if err != nil {
		return
	}
	answer, err = cli.waitAnswer()
	return
}

// WriteAnswer send answer signal
func (cli SignalClient) WriteAnswer(peer string, answer []byte) (err error) {
	log.Printf("Send answer to %s", peer)
	err = cli.writeSignal("answer", peer, answer)
	if err != nil {
		return
	}
	return
}

// WriteCandidate send candidate signal
func (cli SignalClient) WriteCandidate(peer string, candidate []byte) (err error) {
	log.Printf("Send ICECandidate to %s", peer)
	cli.writeSignal("candidate", peer, candidate)
	if err != nil {
		return
	}
	return
}

// writeSignal send signal
func (cli SignalClient) writeSignal(signal, peer string, date []byte) (err error) {
	var i interface{}
	json.Unmarshal(date, &i)
	data, _ := json.Marshal(Signal{signal, peer, i})
	err = cli.conn.Write(cli.ctx, websocket.MessageText, data)
	if err != nil {
		log.Println("Write message error:", err)
		return
	}
	return
}
