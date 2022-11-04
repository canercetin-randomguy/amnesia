package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

func Connect(w http.ResponseWriter, r *http.Request, wg *sync.WaitGroup) {
	// accept the websocket connection from the port :3169
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{
		"*",
	}})
	if err != nil {
		log.Fatal("accept:", err)
	}
	defer func(c *websocket.Conn, code websocket.StatusCode, reason string) {
		err := c.Close(code, reason)
		if err != nil {
			log.Fatal("close:", err)
		}
	}(conn, websocket.StatusInternalError, "the sky is falling")

	// create a context that will expire in 1 hour.
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*60)
	defer cancel()
	for {
		err := WriteTesting(ctx, conn)
		if err != nil {
			log.Fatal("write:", err)
		}
		time.Sleep(time.Second * 5)
	}
	wg.Wait()
}
func WriteTesting(ctx context.Context, conn *websocket.Conn) error {
	// create a json string that can be sent and converted to a string again
	var tempComm TestComm
	tempComm.Message = "başardık"
	fmt.Println("Sending message to websocket:", tempComm.Message)
	pleasework, err := json.Marshal(tempComm)
	newWriter, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		log.Fatal("writer opening:", err)
	}
	_, err = newWriter.Write(pleasework)
	if err != nil {
		log.Fatal("write:", err)
	}
	return newWriter.Close()
}
func Mount(w http.ResponseWriter, r *http.Request) {
	var wg = sync.WaitGroup{}
	wg.Add(1)
	go Connect(w, r, &wg)
	wg.Wait()
}
