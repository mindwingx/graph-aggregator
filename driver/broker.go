package driver

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
	"github.com/mindwingx/graph-aggregator/helper"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type (
	KafkaBroker struct {
		producer *kafka.Producer
		consumer *kafka.Consumer
		conf     brokerConf
		logger   abstractions.LoggerAbstraction
	}

	brokerConf struct {
		Host             string `mapstructure:"KAFKA_HOST"`
		Port             string `mapstructure:"KAFKA_PORT"`
		Group            string `mapstructure:"KAFKA_GROUP"`
		Topic            string `mapstructure:"KAFKA_TOPIC"`
		AutoOffsetReset  string `mapstructure:"AUTO_OFFSET_RESET"`
		EnableAutoCommit string `mapstructure:"ENABLE_AUTO_COMMIT"`
	}

	wsItem struct {
		State   string `json:"state"`
		Message string `json:"message"`
	}
)

func InitMessageBroker(registry abstractions.RegAbstraction, logger abstractions.LoggerAbstraction) abstractions.MessageBrokerAbstraction {
	k := new(KafkaBroker)
	registry.Parse(&k.conf)
	k.logger = logger

	broker := fmt.Sprintf("%s:%s", k.conf.Host, k.conf.Port)

	// init producer instance

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})

	if err != nil {
		fmt.Printf("[socket-aggregator][kafka-producer] failed to create producer: %s\n", err)
		os.Exit(1)
	}

	k.producer = producer

	// init consumer instance

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        broker,
		"broker.address.family":    "v4",
		"group.id":                 k.conf.Group,
		"session.timeout.ms":       6000,
		"auto.offset.reset":        k.conf.AutoOffsetReset,
		"enable.auto.offset.store": false,
	})

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[socket-aggregator][kafka-consumer] failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("[socket-aggregator][kafka-consumer] consumer initialized...")

	err = c.SubscribeTopics([]string{k.conf.Topic}, nil)

	run := true

	go func() {
		for run {
			select {
			case sig := <-signalChan:
				fmt.Printf("Caught signal %v: terminating\n", sig)
				run = false
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					msg := string(e.Value)
					k.logger.Logger().Info("[ws][msg]", zap.String(helper.NewUuid(), msg))

					// start ws
					var wsMsg, wsRes wsItem

					wsMsg = wsItem{
						State:   "aggregator",
						Message: msg,
					}

					ws := wsConn()

					err = ws.WriteJSON(wsMsg)
					if err != nil {
						fmt.Println("wrk", err)
					}

					err = ws.ReadJSON(&wsRes)
					if err != nil {
						fmt.Println("rk", err)
					}
					_ = ws.Close()
					//end ws

					if e.Headers != nil {
						fmt.Printf("%% Headers: %v\n", e.Headers)
					}

					_, err = c.StoreMessage(e)
					if err != nil {
						_, _ = fmt.Fprintf(
							os.Stderr,
							"[socket-aggregator][kafka-consumer] error storing offset after message %s:\n",
							e.TopicPartition,
						)
					}
				case kafka.Error:
					_, _ = fmt.Fprintf(os.Stderr, "[socket-aggregator][kafka-consumer] error: %v: %v\n", e.Code(), e)
					if e.Code() == kafka.ErrAllBrokersDown {
						run = false
					}
				}
			}
		}

		fmt.Printf("[socket-aggregator][kafka-consumer] closing consumer\n")
		_ = c.Close()
	}()

	return k
}

func (k *KafkaBroker) Broker() abstractions.MessageBrokerAbstraction {
	return k
}

func (k *KafkaBroker) Produce(key, value string) error {
	defer func() {
		if rec := recover(); rec != nil {
			// Handle the panic and log the recovered value.
			res := fmt.Sprintf("[socket-aggregator][kafka-produce][recovered] %s", rec)
			fmt.Println(res)
		}
	}()

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.conf.Topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}

	deliveryChan := make(chan kafka.Event)
	err := k.producer.Produce(message, deliveryChan)
	if err != nil {
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}

	close(deliveryChan)
	return nil
}

func wsConn() *websocket.Conn {
	conn, err := helper.ConnectToWebSocket(helper.ProcessorSocketUrl)
	if err != nil {
		log.Println("[api-coordinator]", err)
		return nil
	}

	return conn
}
