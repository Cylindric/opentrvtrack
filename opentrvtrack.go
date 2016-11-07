package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/tarm/serial"
)

// Config represents all the user-configurable options.
type Config struct {
	SerialPort                 string
	SerialBaud                 int
	ThingspeakAPIKey           string
	ThingspeakTemperatureField string
	ThingspeakHumidityField    string
}

// config holds the current user-configurable options.
var config Config

// ReadConfig loads the user-configurable options into config.
func ReadConfig() Config {
	var configfile string

	flag.StringVar(&configfile, "config", "opentrvtrack.config", "Config file")
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	tomlConfig, err := toml.LoadFile(configfile)
	if err != nil {
		log.Fatal(err)
	}

	config.SerialPort = tomlConfig.Get("serial.port").(string)
	config.SerialBaud = tomlConfig.Get("serial.baud").(int)
	config.ThingspeakAPIKey = tomlConfig.Get("thingspeak.api_key").(string)
	config.ThingspeakTemperatureField = tomlConfig.Get("thingspeak.temperature_field").(string)
	config.ThingspeakHumidityField = tomlConfig.Get("thingspeak.humidity_field").(string)
	return config
}

func main() {
	// SendDataToSparkFun("test", 23, 11.23)
	// SendDataToThingSpeak("test", 23, 11.23)
	// log.Fatal("done")

	var config = ReadConfig()
	log.Printf("Connecting to %s at %d\n", config.SerialPort, config.SerialBaud)
	log.Fatal("Done")

	c := &serial.Config{Name: config.SerialPort, Baud: config.SerialBaud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	var dat map[string]interface{}

	reader := bufio.NewReader(s)

	for {
		reply, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Convert the byte array to a string
		str := strings.TrimSpace(string(reply))
		log.Println(str)

		if strings.HasPrefix(str, "OpenTRV") {
			// Welcome Banner
		}

		if strings.HasPrefix(str, "=F") {
			// Data from a stats hub
			// =F@22C3;T1 8 W255 0 F255 0 W255 0 F255 0;C5
		}

		if strings.HasPrefix(str, "{\"@") {
			// Data from a sensor
			// {"@":"C1F8BED8A9AAB8C5","+":14,"L":41,"v|%":0,"tT|C":6}

			// Possible fields:
			// @      Device serial number
			// T|C16  Current temperature, in 1/16th of a Celcius
			// H|%    Current humidty, as %
			// +
			// v|%
			// tT|C
			// cV
			// 0
			// vac
			// tS|C
			// gE
			// L

			if err := json.Unmarshal(reply, &dat); err != nil {
				log.Print(err)
			}
			fmt.Println(str)

			var temp float64
			var humidity int
			var serialnum string

			if val, ok := dat["@"]; ok {
				serialnum = val.(string)
				log.Print("Got Serial " + serialnum)
			}

			if val, ok := dat["T|C16"]; ok {
				temp = val.(float64) / 16
				log.Print("Got Temperature " + strconv.FormatFloat(float64(temp), 'f', 2, 32))
				SendTempDataToThingSpeak(temp)
			}

			if val, ok := dat["H|%"]; ok {
				humidity = val.(int)
				log.Print("Got Humidity " + strconv.Itoa(humidity))
				SendHumidityDataToThingSpeak(humidity)
			}

		}
	}
}

// SendTempDataToThingSpeak sends the supplied temperature reading to ThingSpeak
func SendTempDataToThingSpeak(temp float64) {
	posturl := "https://api.thingspeak.com/update.json"

	postdata := "api_key=" + config.ThingspeakAPIKey
	postdata += "&" + config.ThingspeakTemperatureField + "=" + strconv.FormatFloat(temp, 'f', 2, 32)

	response, err := http.Post(posturl, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(postdata)))

	if err != nil {
		log.Print(err)
	}

	log.Print(response)
}

// SendHumidityDataToThingSpeak sends the supplied humidity reading to ThingSpeak
func SendHumidityDataToThingSpeak(humidity int) {
	posturl := "https://api.thingspeak.com/update.json"

	postdata := "api_key=" + config.ThingspeakAPIKey
	postdata += "&" + config.ThingspeakHumidityField + "=" + strconv.Itoa(humidity)

	response, err := http.Post(posturl, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(postdata)))

	if err != nil {
		log.Print(err)
	}

	log.Print(response)
}
