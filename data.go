package main

const (
	MQTTClientID = "GopherVision3000"

	discoveryTopic   = "vision"
	orientationTopic = "vision/orientation"
	ledsTopic        = "vision/leds"
	circlesTopic     = "vision/circles"
	mazeTopic        = "vision/maze"
)

var (
	WifiSSID     = ""
	WifiPassword = ""

	MQTTServer   = ""
	MQTTProtocol = "tcp"
	MQTTPort     = "1883"
	MQTTUser     = ""
	MQTTPassword = ""
)
