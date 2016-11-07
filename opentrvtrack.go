package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
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

	viper.AddConfigPath("/etc/opentrvtrack/")
	viper.AddConfigPath("$HOME/.opentrvtrack")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s\n", err)
	}

	config.SerialPort = viper.GetString("serial.port")
	config.SerialBaud = viper.GetInt("serial.baud")
	config.ThingspeakAPIKey = viper.GetString("thingspeak.api_key")
	config.ThingspeakTemperatureField = viper.GetString("thingspeak.temperature_field")
	config.ThingspeakHumidityField = viper.GetString("thingspeak.humidity_field")
	return config
}

// ProcessLine takes a byte array and attempts to extract sensor data from it.
func ProcessLine(data []byte) {
	line := strings.TrimSpace(string(data))
	log.Println(line)

	if strings.HasPrefix(line, "OpenTRV") {
		// Welcome Banner
	}

	if strings.HasPrefix(line, "=F") {
		// Data from a stats hub
		// =F@22C3;T1 8 W255 0 F255 0 W255 0 F255 0;C5
	}

	if strings.HasPrefix(line, "{\"@") {
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
		var dat map[string]interface{}

		if err := json.Unmarshal(data, &dat); err != nil {
			log.Print(err)
		}
		// fmt.Println(line)

		var temp float64
		var humidity float64
		var serialnum string

		if rawSerial, ok := dat["@"]; ok {
			serialnum = rawSerial.(string)
			log.Print("Got Serial " + serialnum)
		}

		if rawTemp, ok := dat["T|C16"]; ok {
			temp = rawTemp.(float64) / 16
			log.Print("Got Temperature " + strconv.FormatFloat(float64(temp), 'f', 2, 32))
			SendTempDataToThingSpeak(temp)
		}

		if rawHumid, ok := dat["H|%"]; ok {
			humidity = rawHumid.(float64)
			log.Print("Got Humidity " + strconv.FormatFloat(float64(humidity), 'f', 2, 32))
			SendHumidityDataToThingSpeak(humidity)
		}

	}
}

func main() {
	// SendDataToSparkFun("test", 23, 11.23)
	// SendDataToThingSpeak("test", 23, 11.23)
	// log.Fatal("done")

	// ProcessLine([]byte(`{"@":"C1F8BED8A9AAB8C5","+":3,"L":105,"T|C16":290,"H|%":51}`)) // has temp and humid
	// log.Fatal("done")

	var config = ReadConfig()
	log.Printf("Connecting to %s at %d\n", config.SerialPort, config.SerialBaud)

	c := &serial.Config{Name: config.SerialPort, Baud: config.SerialBaud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	reader := bufio.NewReader(s)

	for {
		reply, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Process the data line
		ProcessLine(reply)

	}
}

// SendTempDataToThingSpeak sends the supplied temperature reading to ThingSpeak
func SendTempDataToThingSpeak(temp float64) {
	posturl := "https://api.thingspeak.com/update.json"

	postdata := "api_key=" + config.ThingspeakAPIKey
	postdata += "&" + config.ThingspeakTemperatureField + "=" + strconv.FormatFloat(temp, 'f', 2, 32)

	response, err := http.Post(posturl, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(postdata)))

	if err != nil {
		log.Print(response)
		log.Print(err)
	}

	// log.Print(response)
}

// SendHumidityDataToThingSpeak sends the supplied humidity reading to ThingSpeak
func SendHumidityDataToThingSpeak(humidity float64) {
	posturl := "https://api.thingspeak.com/update.json"

	postdata := "api_key=" + config.ThingspeakAPIKey
	postdata += "&" + config.ThingspeakHumidityField + "=" + strconv.FormatFloat(humidity, 'f', 2, 32)

	response, err := http.Post(posturl, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(postdata)))

	if err != nil {
		log.Print(response)
		log.Print(err)
	}

	// log.Print(response)
}
