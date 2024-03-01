package abstractions

type RouterAbstraction interface {
	InitWsConnWorkerPool()
	Routes()
	Serve() error
}
