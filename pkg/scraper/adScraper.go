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

const cityURL = "https://www.ebay-kleinanzeigen.de/s-ort-empfehlungen.json?query=%s"

// Ad is a representation of the kleinanzeigen ads
type Ad struct {
	Title string
	Link  string
	Price string
	ID    string
}

// GetAds gets the ads for the specified page serachterm citycode and radius
func GetAds(page int, term string, cityCode int, radius int) []Ad {
	query := fmt.Sprintf(url, page, strings.ReplaceAll(term, " ", "-"), cityCode, radius)
	ads := make([]Ad, 0, 0)

	noneFound := false

	c := colly.NewCollector()

	c.OnHTML(".ad-listitem", func(e *colly.HTMLElement) {
		if !strings.Contains(e.DOM.Nodes[0].Attr[0].Val, "is-topad") {
			link := e.DOM.Find("a[class=ellipsis]")
			linkURL, _ := link.Attr("href")
			price := e.DOM.Find("strong").Text()
			id, idExsits := e.DOM.Find("article[class=aditem]").Attr("data-adid")
			//details := e.DOM.Find("div[class=aditem-details]")
			title := link.Text()
			if idExsits {
				ads = append(ads, Ad{Title: title, Link: "https://www.ebay-kleinanzeigen.de" + linkURL, ID: id, Price: price})
			}
		}
	})

	// if there is a warning with for this search ignore fetched ads
	c.OnHTML(".outcomemessage-warning", func(e *colly.HTMLElement) {
		noneFound = true
	})

	c.Visit(query)

	c.Wait()

	if noneFound {
		return make([]Ad, 0, 0)
	}
	return ads
}

// FindCityID finds the city by the name/postal code
func FindCityID(untrimmed string) (int, string) {
	city := strings.Trim(untrimmed, " ")

	spaceClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(cityURL, city), nil)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:74.0) Gecko/20100101 Firefox/74.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

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
		cityIDString := []rune(key)

		cityID, err := strconv.Atoi(strings.Trim(string(cityIDString[1:]), " "))

		if err != nil {
			return 0, "could not get cityId"
		}

		return cityID, value
	}

	return 0, "no city id found"
}
