package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var cityUrl = "https://www.kleinanzeigen.de/s-ort-empfehlungen.json?query=%s"

func main() {
	city := "Bochum"

	id, _ := findCityId(city)

	log.Printf("%v", id)
}

func findCityId(city string) (int, string) {
	spaceClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(cityUrl, city), nil)

	if err != nil {
		return 0, "could not make request"
	}

	res, getErr := spaceClient.Do(req)

	if getErr != nil {
		return 0, "could not send request"
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		return 0, "could not read response"
	}

	var cities map[string]string

	unmarshalErr := json.Unmarshal(body, &cities)

	if unmarshalErr != nil {
		return 0, "could not parse json"
	}

	if len(cities) == 0 {
		return 0, "could not find city"
	}

	for key, _ := range cities {
		cityIdString := []rune(key)

		cityId, err := strconv.Atoi(strings.Trim(string(cityIdString[1:]), " "))

		if err != nil {
			return 0, "could not get cityId"
		}

		return cityId, ""
	}

	return 0, "no city id found"
}
