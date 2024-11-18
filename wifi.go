package main

import (
	"context"
	"io"
	"net"
	"time"

	"math/rand"

	mqtt "github.com/soypat/natiu-mqtt"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	cl *mqtt.Client

	connectedWifi bool
	connectedMQTT bool
	pubVar        mqtt.VariablesPublish

	// IP address of the MQTT broker to use. Replace with your own info.
	// const server = MQTTProtocol + "://" + MQTTServer + ":" + MQTTPort
	server = MQTTServer + ":" + MQTTPort
)

func connect() {
	if WifiSSID != "" && server != "" {
		if connectToAP() {
			connectToMQTT()
		}
	}
}

func connectToAP() bool {
	time.Sleep(2 * time.Second)
	for retry := 0; retry < 4; retry++ {
		println("Connecting to " + WifiSSID)
		link, _ := probe.Probe()

		err := link.NetConnect(&netlink.ConnectParams{
			Ssid:       WifiSSID,
			Passphrase: WifiPassword,
		})
		if err != nil {
			println("[CONNECT TO AP]", err)
			println("Waiting 15s before trying to reconnect")
			connectedWifi = false
			time.Sleep(15 * time.Second)
		} else {
			connectedWifi = true
			return true
		}
	}
	return false
}

func connectToMQTT() bool {
	println("Connecting to MQTT")

	clientId := MQTTClientID + randomString(10)
	println("ClientId:", clientId)

	// Get a transport for MQTT packets
	println("Connecting to MQTT broker at ", server)
	conn, err := net.Dial("tcp", server)
	if err != nil {
		println("Error connection to MQTT", err)
		return false
	}
	// close it somewhere else ¯\_(ツ)_/¯
	//defer conn.Close()
	println("Connected to MQTT")
	// Create new client
	cl = mqtt.NewClient(mqtt.ClientConfig{
		Decoder: mqtt.DecoderNoAlloc{make([]byte, 1500)},
		OnPub: func(_ mqtt.Header, _ mqtt.VariablesPublish, r io.Reader) error {
			message, _ := io.ReadAll(r)
			println("Message  received on topic", string(message), discoveryTopic)
			return nil
		},
	})

	println("Connecting to client")
	// Connect client
	var varconn mqtt.VariablesConnect
	varconn.SetDefaultMQTT([]byte(clientId))
	varconn.Username = []byte(MQTTUser)
	varconn.Password = []byte(MQTTPassword)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = cl.Connect(ctx, conn, &varconn)
	if err != nil {
		println("failed to connect: ", err)
	}
	println("Connected to client")

	// Subscribe to topic
	println("Subscribing to topic", discoveryTopic)
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = cl.Subscribe(ctx, mqtt.VariablesSubscribe{
		PacketIdentifier: 23,
		TopicFilters: []mqtt.SubscribeRequest{
			{TopicFilter: []byte(discoveryTopic), QoS: mqtt.QoS0},
		},
	})
	if err != nil {
		println("failed to subscribe to", discoveryTopic, err)
		connectedMQTT = false
		return false
	}
	println("Subscribed to topic ", discoveryTopic)

	// Publish on topic
	/*	pubFlags, _ := mqtt.NewPublishFlags(mqtt.QoS0, false, false)
		pubVar := mqtt.VariablesPublish{
			TopicName: []byte(topic),
		}*/

	connectedMQTT = true
	return true
}

func publishDiscovery() {
	if !connectedMQTT {
		return
	}
}

func publishData(topic string, data *[]byte) {
	if !connectedMQTT {
		return
	}

	pubFlags, _ := mqtt.NewPublishFlags(mqtt.QoS0, false, false)
	pubVar.TopicName = []byte(topic)
	pubVar.PacketIdentifier++

	//println("[PUBLISH DATA]", "#"+topic, "MSG TO SEND", string(*data))
	err := cl.PublishPayload(pubFlags, pubVar, *data)
	if err != nil {
		println("error transmitting message: ", err)
	}
}

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}
