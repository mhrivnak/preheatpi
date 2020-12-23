package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

type Relay struct {
	Pin    int
	ID     string
	Driver *gpio.RelayDriver
}

type Response struct {
	Value   string
	Version int
}

func main() {
	url := os.Getenv("PREHEATBOTURL")
	if url == "" {
		log.Fatal("must set envvar PREHEATBOTURL")
	}
	username := os.Getenv("PREHEATBOTUSERNAME")
	if username == "" {
		log.Fatal("must set envvar PREHEATBOTUSERNAME")
	}
	relayString := os.Getenv("RELAYS")
	if relayString == "" {
		log.Fatal("must set envvar RELAYS")
	}
	relayParts := strings.Split(relayString, ",")
	if len(relayParts)%2 != 0 {
		log.Fatal("RELAYS must be specified as CSV alternating GPIO pin and identifier")
	}
	relays := make([]Relay, len(relayParts)/2)
	for i := 0; i*2 < len(relayParts); i++ {
		pin, err := strconv.Atoi(relayParts[i*2])
		if err != nil {
			log.Fatalf("Could not parse GPIO pin integer from \"%s\"", relayParts[i*2])
		}
		relays[i].Pin = pin
		relays[i].ID = relayParts[i*2+1]
	}

	r := raspi.NewAdaptor()
	for i := range relays {
		relays[i].Driver = gpio.NewRelayDriver(r, strconv.Itoa(relays[i].Pin))
		relays[i].Driver.Inverted = true
		relays[i].Driver.Start()
		go watch(url, username, relays[i])
	}
	for {
	}
	/*
			relay1 := gpio.NewRelayDriver(r, "8")
			relay1.Inverted = true
			relay2 := gpio.NewRelayDriver(r, "10")
			relay2.Inverted = true

		for {
			relay1.On()
			fmt.Println("relay 1 on")
			time.Sleep(1 * time.Second)
			relay1.Off()
			fmt.Println("relay 1 off")
			time.Sleep(1 * time.Second)
		}
	*/
}

func watch(apiURL, username string, relay Relay) {
	log.Info(apiURL)
	observedVersion := -1
	for {
		resourceURL, err := url.Parse(apiURL)
		if err != nil {
			log.WithError(err).Error("error parsing URL")
			time.Sleep(30 * time.Second)
			continue
		}
		resourceURL.Path = path.Join(resourceURL.Path, "users", username, "heaters", relay.ID)

		if observedVersion >= 0 {
			values := resourceURL.Query()
			values.Add("longpoll", "true")
			values.Add("version", strconv.Itoa(observedVersion))
			resourceURL.RawQuery = values.Encode()
		}
		log.Info(resourceURL.String())
		resp, err := http.Get(resourceURL.String())
		if err != nil {
			log.WithError(err).Error("error talking to API")
			time.Sleep(30 * time.Second)
			continue
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err).Error("error reading response body")
			time.Sleep(30 * time.Second)
			continue
		}
		response := Response{}
		err = json.Unmarshal(data, &response)
		if err != nil {
			log.WithError(err).Error("error decoding json")
			log.Info(string(data))
			time.Sleep(30 * time.Second)
			continue
		}
		log.Infof("setting relay %s to %s", relay.ID, response.Value)
		switch response.Value {
		case "on":
			relay.Driver.On()
			time.Sleep(1 * time.Second)
			relay.Driver.On()
		case "off":
			relay.Driver.Off()
		default:
			log.Errorf("got unknown value: ", response.Value)
		}
		observedVersion = response.Version
	}
}
