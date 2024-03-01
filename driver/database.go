package driver

import (
	"fmt"
	src "github.com/mindwingx/graph-aggregator"
	"github.com/mindwingx/graph-aggregator/database/models"
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
	"github.com/mindwingx/graph-aggregator/helper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sort"
	"time"
)

type (
	Sql struct {
		config    config
		DB        *gorm.DB
		MaxDbChan chan models.TransactionalOutboxMessages // to balance max new connections based on MAXIDLECONNECTIONS
	}

	config struct {
		Debug              bool   `mapstructure:"DEBUG"`
		Host               string `mapstructure:"DB_HOST"`
		Port               string `mapstructure:"DB_PORT"`
		Username           string `mapstructure:"USERNAME"`
		Password           string `mapstructure:"PASSWORD"`
		Database           string `mapstructure:"DATABASE"`
		Ssl                string `mapstructure:"SSL"`
		MaxIdleConnections int    `mapstructure:"MAXIDLECONNECTIONS"`
		MaxOpenConnections int    `mapstructure:"MAXOPENCONNECTIONS"`
		MaxLifetimeSeconds int    `mapstructure:"MAXLIFETIMESECONDS"`
	}
)

func NewSql(registry abstractions.RegAbstraction) abstractions.SqlAbstraction {
	database := new(Sql)
	registry.Parse(&database.config)
	database.MaxDbChan = make(chan models.TransactionalOutboxMessages, database.config.MaxOpenConnections)

	return database
}

func (g *Sql) InitSql() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		g.config.Username,
		g.config.Password,
		g.config.Host,
		g.config.Port,
		g.config.Database,
	)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 g.newGormLog(),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatal("DB conn. init error:", err)
	}

	sqlDatabase, err := database.DB()
	if err != nil {
		log.Fatal("DB conn. retrieve error:", err)
	}

	if g.config.MaxIdleConnections != 0 {
		sqlDatabase.SetMaxIdleConns(g.config.MaxIdleConnections)
	}

	if g.config.MaxOpenConnections != 0 {
		sqlDatabase.SetMaxOpenConns(g.config.MaxOpenConnections)
	}

	if g.config.MaxLifetimeSeconds != 0 {
		sqlDatabase.SetConnMaxLifetime(time.Second * time.Duration(g.config.MaxLifetimeSeconds))
	}

	if g.config.Debug {
		database = database.Debug()
		log.Println("DB debug mode enabled")
	}

	g.DB = database
}

func (g *Sql) Migrate() {
	path := fmt.Sprintf("%s/database/mysql", src.Root())

	// Open the directory
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal("DB sql file loading failure")
	}

	defer func() { _ = dir.Close() }()

	// Read the directory contents
	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal("DB sql dir scan failure")
	}

	// Sort the entries alphabetically by name - Sql file order by numeric(01, 02, etc)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Name() < fileInfos[j].Name()
	})

	// Iterate over the file info slice and print the file names
	for _, fileInfo := range fileInfos {
		if fileInfo.Mode().IsRegular() {
			if err = g.DB.Exec(g.parseSqlFile(path, fileInfo)).Error; err != nil {
				log.Fatal("DB migrate failure")
			}
		}
	}
}

func (g *Sql) Sql() abstractions.SqlAbstraction {
	return g
}

func (g *Sql) Db() *gorm.DB {
	return g.DB
}

func (g *Sql) DbChan() chan models.TransactionalOutboxMessages {
	return g.MaxDbChan
}

func (g *Sql) Close() {
	sql, err := g.DB.DB()
	if err != nil {
		log.Fatal("DB conn. close failure")
	}

	_ = sql.Close()
}

// HELPER METHOD

func (g *Sql) newGormLog() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             helper.SlowSqlThreshold * time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn,                           // Log level
			IgnoreRecordNotFoundError: false,                                 // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                                  // Disable color
		})
}

func (g *Sql) parseSqlFile(path string, fileInfo os.FileInfo) string {
	sqlFile := fmt.Sprintf("%s/%s", path, fileInfo.Name())
	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("DB sql file parse failure")
	}
	// Convert SQL file contents to string
	return string(sqlBytes)
}
