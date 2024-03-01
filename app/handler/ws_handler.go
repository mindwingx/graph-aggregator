package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mindwingx/graph-aggregator/database/models"
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
	"github.com/mindwingx/graph-aggregator/helper"
	"log"
	"net/http"
	"sync"
)

type (
	WsConn struct {
		*websocket.Conn
	}

	WsItem struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}
)

var (
	mx       = sync.RWMutex{}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request, workerPool chan struct{}, db abstractions.SqlAbstraction, broker abstractions.MessageBrokerAbstraction) {
	<-workerPool

	defer func() {
		workerPool <- struct{}{}
	}()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("[socket-aggregator] ws. connection upgrade error:", err)
		return
	}

	response := WsItem{
		State:   "aggregator",
		Message: "message received",
	}

	mx.Lock()
	conn := WsConn{Conn: ws}

	//used the goroutine to handle too many request error
	go func() {
		if err = ws.WriteJSON(response); err != nil {
			fmt.Println("[socket-aggregator] write-json error:", err)
		}
	}()
	mx.Unlock()

	go wsListener(&conn, db)

	go func() {
		for {
			select {
			case msg := <-db.DbChan():
				tx := db.Db()
				er := tx.Omit("deleted_at").Create(&msg).Error
				if er != nil {
					fmt.Println("er:", er)
				}

				_ = broker.Produce(msg.Uuid, msg.Message)
			}
		}
	}()

}

// HELPER METHODS

func wsListener(
	conn *WsConn,
	db abstractions.SqlAbstraction,
) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[socket-aggregator][err] ", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsItem

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// Log the error and break out of the loop if there's an issue reading from the WebSocket
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				_ = conn.Close()
				fmt.Println("[socket-aggregator] unexpected close error:", err)
			}
			break
		}

		message := models.TransactionalOutboxMessages{
			Uuid:    helper.NewUuid(),
			Message: payload.Message,
		}

		db.DbChan() <- message
	}
}

func txHandler(
	msg string,
	db abstractions.SqlAbstraction,
	broker abstractions.MessageBrokerAbstraction,
) {
	fmt.Println("cap", cap(db.DbChan()))
	tx := db.Sql().Db()
	tx.Begin()

	defer func() {
		log.Println("rc")
		//db.DbChan() <- true
		recErr := recover()

		if recErr != nil {
			tx.Rollback()
			log.Printf("[socket-aggregator][recovered] db/produce error: %v", recErr)
		} else {
			tx.Commit()
			fmt.Println("done")
		}
	}()

	uuid := helper.NewUuid()
	message := models.TransactionalOutboxMessages{
		Uuid:    uuid,
		Message: msg,
	}

	err := tx.Omit("deleted_at").Create(&message).Error
	log.Println("tx", err)
	if err != nil {
		log.Printf("[socket-aggregator] failed to insert message: %v", err)
		return
	}

	log.Println("br")
	if err := broker.Produce(message.Uuid, message.Message); err != nil {
		log.Printf("[socket-aggregator] failed to produce Kafka message: %v", err)
		return
	}
}
