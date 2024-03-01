package helper

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func NewUuid() string {
	val, _ := uuid.NewRandom()
	return val.String()
}

func ConnectToWebSocket(url string) (*websocket.Conn, error) {
	var conn *websocket.Conn
	var err error
	st := time.Now()

	for {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			return conn, nil
		}

		log.Println("[socket-processor] connection error:", err)
		time.Sleep(250 * time.Millisecond) // Retry after 2 seconds

		if latency := time.Since(st).Milliseconds(); latency >= 500 {
			err = errors.New("socket connection failure")
			return nil, err
		}
	}
}
