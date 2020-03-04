package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const url = "https://www.ebay-kleinanzeigen.de/seite:%v/s-%s/k0l%vr%v"

const cityUrl = "https://www.ebay-kleinanzeigen.de/s-ort-empfehlungen.json?query=%s"

type Ad struct {
	Title string
	Link  string
	Price string
	Id    string
}

func GetAds(page int, term string, cityCode int, radius int) []Ad {
	query := fmt.Sprintf(url, page, strings.ReplaceAll(term, " ", "-"), cityCode, radius)
	ads := make([]Ad, 0, 0)

	c := colly.NewCollector()

	c.OnHTML(".ad-listitem", func(e *colly.HTMLElement) {
		if !strings.Contains(e.DOM.Nodes[0].Attr[0].Val, "is-topad") {
			link := e.DOM.Find("a[class=ellipsis]")
			linkUrl, _ := link.Attr("href")
			price := e.DOM.Find("strong").Text()
			id, idExsits := e.DOM.Find("article[class=aditem]").Attr("data-adid")
			//details := e.DOM.Find("div[class=aditem-details]")
			title := link.Text()
			if idExsits {
				ads = append(ads, Ad{Title: title, Link: "https://www.ebay-kleinanzeigen.de" + linkUrl, Id: id, Price: price})
			}
		}
	})

	c.Visit(query)

	c.Wait()
	return ads
}

func FindCityId(untrimmed string) (int, string) {
	city := strings.Trim(untrimmed, " ")

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

	for key, value := range cities {
		cityIdString := []rune(key)

		cityId, err := strconv.Atoi(strings.Trim(string(cityIdString[1:]), " "))

		if err != nil {
			return 0, "could not get cityId"
		}

		return cityId, value
	}

	return 0, "no city id found"
}
