package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gocolly/colly"
)

const url = "https://www.ebay-kleinanzeigen.de/seite:%v/s-%s/k0l%vr%v"

const cityURL = "https://www.ebay-kleinanzeigen.de/s-ort-empfehlungen.json?query=%s"

// Ad is a representation of the kleinanzeigen ads
type Ad struct {
	Title    string
	Link     string
	Price    string
	Location string
	ID       string
}

// GetAds gets the ads for the specified page serachterm citycode and radius
func GetAds(page int, term string, cityCode int, radius int, maxPrice *int, minPrice *int) []Ad {
	log.Debug().Msg("scraping for ads")
	query := fmt.Sprintf(url, page, strings.ReplaceAll(term, " ", "-"), cityCode, radius)
	ads := make([]Ad, 0, 0)
	c := colly.NewCollector()

	c.OnHTML("#srchrslt-adtable", func(adListEl *colly.HTMLElement) {
		adListEl.ForEach(".ad-listitem", func(_ int, e *colly.HTMLElement) {
			if !strings.Contains(e.DOM.Nodes[0].Attr[0].Val, "is-topad") {
				link := e.DOM.Find("a[class=ellipsis]")
				linkURL, _ := link.Attr("href")
				price := strings.TrimSpace(e.DOM.Find("p[class=aditem-main--middle--price-shipping--price]").Text())

				space := regexp.MustCompile(`\s+`)
				location := strings.TrimSpace(e.DOM.Find("div [class=aditem-main--top--left]").Last().Text())

				location = space.ReplaceAllString(location, " ")

				if maxPrice != nil && strings.ToLower(price) != "zu verschenken" {
					replacted := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.Trim(price, " "), "VB", ""), "€", ""), ".", ""), " ")

					if len(replacted) == 0 {
						return
					}

					priceValue, err := strconv.Atoi(replacted)

					if err != nil {
						log.Warn().Str("price-string", replacted).Msg("could not parse price from ad")
						return
					}

					if priceValue >= *maxPrice {
						log.Debug().Msg("price is bigger than requested")
						return
					}

					if minPrice != nil && priceValue < *minPrice {
						log.Debug().Msg("price is lower than requested")
						return
					}
				}

				id, idExsits := e.DOM.Find("article[class=aditem]").Attr("data-adid")
				//details := e.DOM.Find("div[class=aditem-details]")
				title := link.Text()
				if idExsits {
					ads = append(ads, Ad{Title: title, Link: "https://www.ebay-kleinanzeigen.de" + linkURL, ID: id, Price: price, Location: location})
				}
			}
		})
	})
	c.OnError(func(r *colly.Response, e error) {
		log.Error().Err(e).Str("term", term).Int("radius", radius).Msg("error while scraping for ads")
	})

	c.Visit(query)

	c.Wait()

	log.Debug().Str("query", term).Int("number_of_queries", len(ads)).Msg("scraped ads for query")

	return ads
}

// FindCityID finds the city by the name/postal code
func FindCityID(untrimmed string) (int, string, error) {
	log.Debug().Str("city_search_term", untrimmed).Msg("finding city id")

	city := strings.Trim(untrimmed, " ")

	spaceClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(cityURL, city), nil)

	if err != nil {
		log.Error().Err(err).Msg("could not create the request")
		return 0, "", errors.New("could not make request")
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:74.0) Gecko/20100101 Firefox/74.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	res, getErr := spaceClient.Do(req)

	if getErr != nil {
		return 0, "", errors.New("could not send request")
	}

	if res.StatusCode != 200 {
		log.Error().Str("status_code", res.Status).Msg("received a wrong status code.")
		if res.StatusCode == 403 {
			log.Error().Msg("ip address might be blocked by kleinanzeigen.")
		}
		return 0, "", errors.New("request for city not successful")
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		return 0, "", errors.New("could not read response")
	}

	var cities map[string]string

	unmarshalErr := json.Unmarshal(body, &cities)

	if unmarshalErr != nil {
		return 0, "", errors.New("could not parse json")
	}

	if len(cities) == 0 {
		return 0, "", errors.New("could not find city")
	}

	for key, value := range cities {
		cityIDString := []rune(key)

		cityID, err := strconv.Atoi(strings.Trim(string(cityIDString[1:]), " "))

		if err != nil {
			return 0, "", errors.New("could not get cityId")
		}

		log.Debug().Int("city_id", cityID).Str("city_name", value).Msg("found city")

		return cityID, value, nil
	}

	return 0, "", errors.New("no city id found")
}
