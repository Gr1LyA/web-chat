package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

const natsURL = "nats://nats:4222"

var upgrader = websocket.Upgrader{}

func main() {
	nc, err := nats.Connect(natsURL, nats.ReconnectWait(time.Second), nats.MaxReconnects(600))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer nc.Close()

	http.HandleFunc("/ws", middleware(nc))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func middleware(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer connection.Close()

		log.Println("client connected to ws")

		// Получаем message из nats и отправляем по вебсокету а фронт
		nSub, err := nc.Subscribe("nats-server", func(msg *nats.Msg) {
			connection.WriteMessage(websocket.TextMessage, msg.Data)
		})
		defer nSub.Unsubscribe()

		// Читаем из сокета и отправляем в nats
		for {
			mt, message, err := connection.ReadMessage()
			if err != nil || mt == websocket.CloseMessage {
				break // Выходим из цикла, если клиент пытается закрыть соединение или связь прервана
			}

			nc.Publish("nats-server", message)
			log.Println("[message]", string(message))
		}
	}
}
