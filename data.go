package main

const (
	MQTTClientID = "GopherVision3000"

	discoveryTopic          = "vision"
	orientationTopic        = "vision/orientation"
	ledsTopic               = "vision/leds"
	circlesTopic            = "vision/circles"
	circlesArcTopic         = "vision/circleArc"
	circlesOrientationTopic = "vision/circleOrientation"
	circlesRadiusTopic      = "vision/circleRadius"
	mazeTopic               = "vision/maze"
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
