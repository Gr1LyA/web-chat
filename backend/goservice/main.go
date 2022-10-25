package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

const natsURL = "nats://nats:4222"

var upgrader = websocket.Upgrader{}

func main() {
	http.HandleFunc("/ws", ws)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ws(w http.ResponseWriter, r *http.Request) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}

	connection, _ := upgrader.Upgrade(w, r, nil)
	defer connection.Close()
	defer nc.Close()

	// Получаем message из nats и отправляем по вебсокету а фронт
	nc.Subscribe("nats-server", func(msg *nats.Msg) {
		connection.WriteMessage(websocket.TextMessage, msg.Data)
	})

	// Читаем из сокета и отправляем в nats
	for {
		mt, message, err := connection.ReadMessage()

		if err != nil || mt == websocket.CloseMessage {
			break // Выходим из цикла, если клиент пытается закрыть соединение или связь прервана
		}
		nc.Publish("nats-server", message)
		fmt.Println("[message]", string(message))
	}
}
