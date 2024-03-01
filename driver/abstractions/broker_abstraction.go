package abstractions

type MessageBrokerAbstraction interface {
	Broker() MessageBrokerAbstraction
	Produce(key, value string) error
}
