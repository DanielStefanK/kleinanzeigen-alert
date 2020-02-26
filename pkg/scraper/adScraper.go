package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

const url = "https://www.ebay-kleinanzeigen.de/seite:%v/s-%s/k0l%vr%v"

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
