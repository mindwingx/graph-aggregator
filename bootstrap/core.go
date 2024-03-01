package bootstrap

import "github.com/mindwingx/graph-aggregator/driver"

func (app *App) initRegistry() {
	app.registry = driver.NewViper()
	app.registry.InitRegistry()
}

func (app *App) initDatabase() {
	app.database = driver.NewSql(app.registry)
	app.database.InitSql()
	app.database.Migrate()
}

func (app *App) initLogger() {
	app.logger = driver.InitLogger(app.registry)
	_ = app.logger.Logger().Sync()
}

func (app *App) initMessageBroker() {
	app.msgBroker = driver.InitMessageBroker(app.registry, app.logger)
}

func (app *App) initRouter() {
	app.router = driver.InitRouter(app.registry, app.database, app.msgBroker)
	app.router.InitWsConnWorkerPool()
	app.router.Routes()
}
