package bootstrap

import (
	"fmt"
	"log"
)

func (app *App) Start() {
	fmt.Println("[socket-aggregator] service started...")
	err := app.router.Serve()
	if err != nil {
		app.database.Close()
		close(app.database.DbChan())

		log.Fatal(err)
	}
}
