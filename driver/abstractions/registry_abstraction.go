package abstractions

type RegAbstraction interface {
	InitRegistry()
	Parse(interface{})
}
