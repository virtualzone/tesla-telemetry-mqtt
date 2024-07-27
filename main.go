package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	zmq "github.com/pebbe/zmq4"
	"github.com/virtualzone/tesla-telemetry-mqtt/protos"
	"google.golang.org/protobuf/proto"
)

var mqttClient mqtt.Client

func main() {
	log.Println("Starting Fleet Telemetry ZMQ to MQTT Forwarder...")
	GetConfig().ReadConfig()
	connectMqtt()
	serveZMQ()
}

func serveZMQ() {
	if GetConfig().ZMQPublisher == "" {
		log.Panicf("ZMQ Publisher not set")
		return
	}

	log.Println("Initializing ZMQ subscriber...")

	zctx, _ := zmq.NewContext()
	s, _ := zctx.NewSocket(zmq.SUB)
	defer s.Close()
	if err := s.Connect(GetConfig().ZMQPublisher); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	if err := s.SetSubscribe(""); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	go func() {
		for {
			zmqLoop(s)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	mqttClient.Disconnect(250)
}

func zmqLoop(s *zmq.Socket) {
	address, err := s.Recv(0)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("reading zmq message in channel %s\n", address)

	if msg, err := s.RecvBytes(0); err != nil {
		log.Println(err)
		return
	} else {
		data := &protos.Payload{}
		if err := proto.Unmarshal(msg, data); err != nil {
			log.Println(err)
			return
		}

		log.Println(data)

		topicPrefix := GetConfig().MqttTopicPrefix
		for _, e := range data.Data {
			keyName := protos.Field_name[int32(e.Key)]
			value := e.Value.GetStringValue()
			topic := fmt.Sprintf("%s%s/%s", topicPrefix, data.Vin, keyName)
			log.Printf("Publishing to MQTT Topic %s = %s\n", topic, value)
			mqttClient.Publish(topic, 0, true, value)
		}
	}
}

func connectMqtt() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(GetConfig().MqttBroker)
	opts.SetClientID(GetConfig().MqttClientID)
	opts.SetUsername(GetConfig().MqttUsername)
	opts.SetPassword(GetConfig().MqttPassword)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Panicf(token.Error().Error())
	}
	mqttClient = c
}
