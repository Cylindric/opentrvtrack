package main

import (
  "bufio"
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "strconv"
  "strings"
  "time"
  "github.com/tarm/goserial"
)

func main() {
  // SendDataToSparkFun("test", 23, 11.23)
  // SendDataToThingSpeak("test", 23, 11.23)
  // log.Fatal("done")

  c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 4800}
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

    if strings.HasPrefix(str, "OpenTRV") {
      // Welcome Banner
      log.Println(str)
    }

    if strings.HasPrefix(str, "=F") {
      // Data from a stats hub
      // str := string(reply[:n])
      // =F@22C3;T1 8 W255 0 F255 0 W255 0 F255 0;C5
    }

    if strings.HasPrefix(str, "{\"@") {
      // Data from a sensor
      // {"@":"C1F8BED8A9AAB8C5","+":14,"L":41,"v|%":0,"tT|C":6}

      // Possible fields:
      // @
      // +
      // v|%
      // tT|C
      // cV
      // T|C16
      // H|%
      // 0
      // vac
      // tS|C
      // gE
      // L

      if err := json.Unmarshal(reply, &dat); err != nil {
        log.Print(err)
      }
      fmt.Println(str)

      var temp float64 = 0.0
      var humidity int = 0
      var serialnum string = ""

      // var ready bool = true

      if val, ok := dat["@"]; ok {
        serialnum = val.(string)
        log.Print("Got Serial " + serialnum)
      } else {
        // ready = false
      }

      if val, ok := dat["T|C16"]; ok {
        temp = val.(float64) / 16
        log.Print("Got Temperature " + strconv.FormatFloat(float64(temp), 'f', 2, 32))
        SendTempDataToThingSpeak(temp)
      } else {
        // ready = false
      }

      if val, ok := dat["H|%"]; ok {
        humidity = val.(int)
        log.Print("Got Humidity " + strconv.Itoa(humidity))
        SendHumidityDataToThingSpeak(humidity)
      } else {
        // ready = false
      }

      // if ready {
      //   SendDataToSparkFun(serialnum, humidity, temp)
      // }
    }
  }
}


// func SendDataToSparkFun(serialnum string, humidity int, temp float64) {
//   geturl := "http://data.sparkfun.com/input/[publicKey]?private_key=[privateKey]&humidity=[humidity]&serial=[serial]&temp=[temp]"
//   pubkey := "4J1GOzA0wgT6mv97NJ1g"
//   privkey := "b5108DZByrinBMjKJk47"
//
//   geturl = strings.Replace(geturl, "[publicKey]", pubkey, -1)
//   geturl = strings.Replace(geturl, "[privateKey]", privkey, -1)
//   geturl = strings.Replace(geturl, "[humidity]", strconv.Itoa(humidity), -1)
//   geturl = strings.Replace(geturl, "[temp]", strconv.FormatFloat(float64(temp), 'f', 2, 32), -1)
//   geturl = strings.Replace(geturl, "[serial]", serialnum, -1)
//
//   // log.Print(geturl)
//   response, err := http.Get(geturl)
//   if err != nil {
//     log.Print(err)
//   }
//   log.Print(response)
// }

func SendTempDataToThingSpeak(temp float64) {
  posturl := "https://api.thingspeak.com/update.json"
  writekey := ""

  postdata := "api_key=" + writekey
  postdata += "&field2=" + strconv.FormatFloat(temp, 'f', 2, 32)

  response, err := http.Post(posturl, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(postdata)))

  if err != nil {
    log.Print(err)
  }
  // log.Print(postdata)
  log.Print(response)
}


func SendHumidityDataToThingSpeak(humidity int) {
  posturl := "https://api.thingspeak.com/update.json"
  writekey := ""

  postdata := "api_key=" + writekey
  postdata += "&field3=" + strconv.Itoa(humidity)

  response, err := http.Post(posturl, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(postdata)))

  if err != nil {
    log.Print(err)
  }
  // log.Print(postdata)
  log.Print(response)
}
