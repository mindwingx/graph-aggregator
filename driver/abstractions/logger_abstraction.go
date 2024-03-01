package abstractions

import "go.uber.org/zap"

type LoggerAbstraction interface {
	Logger() *zap.Logger
}
