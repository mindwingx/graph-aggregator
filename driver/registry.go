package driver

import (
	"fmt"
	src "github.com/mindwingx/graph-aggregator"
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
	constants "github.com/mindwingx/graph-aggregator/helper"
	registry "github.com/spf13/viper"
	"log"
)

type Viper struct {
	*registry.Viper
}

func NewViper() abstractions.RegAbstraction {
	return &Viper{registry.New()}
}

func (v *Viper) InitRegistry() {
	envPath := fmt.Sprintf("%s/config", src.Root())

	v.AddConfigPath(envPath)
	v.SetConfigName(".env")
	v.SetConfigType(constants.EnvFile)
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func (v *Viper) Parse(item interface{}) {
	err := v.Unmarshal(&item)
	if err != nil {
		log.Fatal(err)
	}
}
