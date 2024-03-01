package abstractions

import (
	"github.com/mindwingx/graph-aggregator/database/models"
	"gorm.io/gorm"
)

type SqlAbstraction interface {
	InitSql()
	Migrate()
	Sql() SqlAbstraction
	Db() *gorm.DB
	DbChan() chan models.TransactionalOutboxMessages
	Close()
	//Seed()
}
