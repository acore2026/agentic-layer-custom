package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type IntentRequest struct {
	Type string `json:"type"`
	Data struct {
		Intent     string `json:"intent"`
		ScenarioID string `json:"scenarioId"`
	} `json:"data"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/v1/intents/stream"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("recv: %s\n", message)
			
			// If we see workflow_complete, we can exit after a short delay
			if string(message) != "" {
				var event struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(message, &event); err == nil && event.Type == "workflow_complete" {
					log.Println("Workflow complete received, exiting...")
					time.Sleep(1 * time.Second)
					return
				}
			}
		}
	}()

	req := IntentRequest{
		Type: "execute_intent",
	}
	req.Data.Intent = "Connect my new embodied agent to a high-reliability subnet."
	req.Data.ScenarioID = "ACN"

	payload, _ := json.Marshal(req)
	err = c.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
