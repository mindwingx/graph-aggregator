package main

import "github.com/mindwingx/graph-aggregator/bootstrap"

func main() {
	service := bootstrap.NewApp()
	service.Init()
	service.Start()
}
