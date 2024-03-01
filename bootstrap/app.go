package bootstrap

import (
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
)

type App struct {
	registry  abstractions.RegAbstraction
	database  abstractions.SqlAbstraction
	msgBroker abstractions.MessageBrokerAbstraction
	logger    abstractions.LoggerAbstraction
	router    abstractions.RouterAbstraction
}

func NewApp() *App {
	return &App{}
}

func (app *App) Init() {
	app.initRegistry()
	app.initDatabase()
	app.initLogger()
	app.initMessageBroker()
	app.initRouter()
}
