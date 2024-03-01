package driver

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mindwingx/graph-aggregator/app/handler"
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
	"net/http"
	"time"
)

type (
	Mux struct {
		router     *mux.Router
		handler    *http.Server
		conf       routerConfig
		workerPool chan struct{}
		db         abstractions.SqlAbstraction
		broker     abstractions.MessageBrokerAbstraction
		//logger     LoggerAbstraction
	}

	routerConfig struct {
		Host         string `mapstructure:"HOST"`
		Port         string `mapstructure:"PORT"`
		WriteTimeout int    `mapstructure:"WRITE_TIMEOUT"`
		ReadTimeout  int    `mapstructure:"READ_TIMEOUT"`
		MaxWorkers   int    `mapstructure:"WS_HANDLER_MAX_WORKERS"`
	}
)

func InitRouter(
	registry abstractions.RegAbstraction,
	db abstractions.SqlAbstraction,
	broker abstractions.MessageBrokerAbstraction,
) abstractions.RouterAbstraction {
	m := new(Mux)
	registry.Parse(&m.conf)
	m.router = mux.NewRouter()
	m.workerPool = make(chan struct{}, m.conf.MaxWorkers)
	m.db = db
	m.broker = broker.Broker()
	//m.logger = logger

	return m
}

func (mux *Mux) InitWsConnWorkerPool() {
	fmt.Println("[socket-aggregator] worker pool started...")

	for i := 0; i < mux.conf.MaxWorkers; i++ {
		mux.workerPool <- struct{}{}
	}
}

func (mux *Mux) Routes() {
	// SOCKET routes
	socket := mux.router.PathPrefix("/ws").Subrouter()
	socket.HandleFunc("/aggregator", func(rw http.ResponseWriter, r *http.Request) {
		handler.WebSocketHandler(rw, r, mux.workerPool, mux.db, mux.broker.Broker())
	})

	mux.handler = &http.Server{
		Handler:      mux.router,
		Addr:         fmt.Sprintf("%s:%s", mux.conf.Host, mux.conf.Port),
		WriteTimeout: time.Duration(mux.conf.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(mux.conf.ReadTimeout) * time.Second,
	}
}

func (mux *Mux) Serve() error {
	return mux.handler.ListenAndServe()
}
