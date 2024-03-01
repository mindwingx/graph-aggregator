package helper

import "time"

const (
	EnvFile                        = "env"
	TimestampLayout  string        = "2006-01-02 03:04:05"
	SlowSqlThreshold time.Duration = 5
	//ProcessorSocketUrl               = "ws://localhost:9991/ws/processor"
	ProcessorSocketUrl = "ws://processor:9991/ws/processor" // docker network
)
