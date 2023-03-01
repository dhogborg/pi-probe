package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const Topic = "pi-probe/temperature"

var crc = regexp.MustCompile("YES")
var temp = regexp.MustCompile("t=(.*)")

var path = flag.String("path", "", "path to device file")
var mqttAddr = flag.String("mqtt", "", "ip address to mqtt server")

func main() {
	flag.Parse()

	if path == nil || *path == "" {
		panic("required argument: path")
	}

	t, err := readTemperature(*path)
	if err != nil {
		log.Fatalf("unable to read temperature: %s", err.Error())
	}

	fmt.Printf("temperature: %0.2f°C\n", t)

	if mqttAddr != nil && *mqttAddr != "" {
		jsb, err := json.Marshal(map[string]any{"temperature": t})
		if err != nil {
			log.Fatalf("unable to marshal value to mqtt: %s", err.Error())
		}

		client, err := getMqttClient(*mqttAddr)
		if err != nil {
			log.Fatalf("unable to connect to mqtt: %s", err.Error())
		}

		token := client.Publish(Topic, 0, false, string(jsb))
		token.Wait()
		if token.Error() != nil {
			log.Fatalf("mqtt publish error: %s", err.Error())
		}

		fmt.Printf("published to mqtt: %s\n", Topic)
	}

}

func readTemperature(path string) (float64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	if !crc.Match(b) {
		return 0, errors.New("no valid CRC found")
	}

	m := temp.FindAllString(string(b), -1)
	if len(m) == 0 {
		return 0, errors.New("no temperature matches")
	}

	millic, err := strconv.ParseInt(m[0][2:], 10, 64)
	if err != nil {
		return 0, err
	}

	return float64(millic) / 1000, nil
}

func getMqttClient(addr string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", addr, 1883))
	opts.SetClientID("pi-probe_mqtt_client")
	opts.WriteTimeout = 20 * time.Second
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("connected to mqtt broker")
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Printf("lost mqtt broker: %s\n", err.Error())
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
