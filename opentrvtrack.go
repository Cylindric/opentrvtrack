package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cylindric/opentrvgo"
	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/spf13/viper"
	"github.com/tarm/serial"
	"github.com/twinj/uuid"
)

// Config represents all the user-configurable options.
type Config struct {
	SerialPort                 string
	SerialBaud                 int
	ThingspeakAPIKey           string
	ThingspeakTemperatureField string
	ThingspeakHumidityField    string
	LibratoAPIKey              string
	LibratoUsername            string
	ElasticIndex               string
}

// config holds the current user-configurable options.
var config Config

var host string

var es = elastigo.NewConn()

func main() {
	var config = ReadConfig()
	log.Printf("Connecting to %s at %d\n", config.SerialPort, config.SerialBaud)

	log.SetFlags(log.LstdFlags)
	flag.Parse()

	// Trace all requests
	es.RequestTracer = func(method, url, body string) {
		log.Printf("Requesting %s %s", method, url)
		log.Printf("Request body: %s", body)
	}

	c := &serial.Config{Name: config.SerialPort, Baud: config.SerialBaud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	// Give the board time to reset after Serial wakes it up.
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
	config.LibratoAPIKey = viper.GetString("librato.api_key")
	config.LibratoUsername = viper.GetString("librato.username")
	config.ElasticIndex = viper.GetString("elasticsearch.index")

	es.Domain = viper.GetString("elasticsearch.server")

	return config
}

// ProcessLine takes a byte array and attempts to extract sensor data from it.
func ProcessLine(input []byte) {

	if strings.HasPrefix(string(input), "{\"@") {

		sample, err := opentrvgo.ParseSensorReport(input)
		if err != nil {
			log.Print(string(input))
			log.Printf("Error parsing response: %s", err)
			return
		}

		err = SendDataToES(sample)
	}
}

// SendDataToES sends the supplied data packet to ElasticSearch
func SendDataToES(sample map[string]interface{}) (err error) {
	id := uuid.NewV4()
	response, err := es.Index(fmt.Sprintf("%s-%s", config.ElasticIndex, time.Now().Format("2006-01-02")), "sample", id.String(), nil, sample)

	if err != nil {
		log.Print(response)
		log.Print(err)
	}

	return err
}

// SendDataToLibrato sends the supplied temperature reading to SendTempDataToLibrato
func SendDataToLibrato(gauge string, sensor string, temp float64) {
	posturl := "https://metrics-api.librato.com/v1/metrics"

	var jsonStr = []byte(`{"gauges":[{"name":"` + gauge + `","value":"` + strconv.FormatFloat(temp, 'f', 2, 32) + `","source":"` + sensor + `"}]}`)

	client := &http.Client{}
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(jsonStr))
	req.SetBasicAuth(config.LibratoUsername, config.LibratoAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Print(resp)
		log.Print(err)
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
}
