package main

import (
	"os"
	"strings"
	"sync"
)

type Config struct {
	ZMQPublisher    string
	MqttBroker      string
	MqttClientID    string
	MqttUsername    string
	MqttPassword    string
	MqttTopicPrefix string
}

var _configInstance *Config
var _configOnce sync.Once

func GetConfig() *Config {
	_configOnce.Do(func() {
		_configInstance = &Config{}
		_configInstance.ReadConfig()
	})
	return _configInstance
}

func (c *Config) ReadConfig() {
	c.ZMQPublisher = c.getEnv("ZMQ_PUB", "tcp://localhost:5555")
	c.MqttBroker = c.getEnv("MQTT_BROKER", "tcp://localhost:1883")
	c.MqttClientID = c.getEnv("MQTT_CLIENT_ID", "fleet-telemetry")
	c.MqttUsername = c.getEnv("MQTT_USERNAME", "")
	c.MqttPassword = c.getEnv("MQTT_PASSWORD", "")
	c.MqttTopicPrefix = c.getEnv("MQTT_TOPIC_PREFIX", "tesla/telemetry")
	if !strings.HasSuffix(c.MqttTopicPrefix, "/") {
		c.MqttTopicPrefix += "/"
	}
}

func (c *Config) getEnv(key, defaultValue string) string {
	res := os.Getenv(key)
	if res == "" {
		return defaultValue
	}
	return res
}
