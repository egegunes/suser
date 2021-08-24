package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		suser := r.URL.Query().Get("suser")
		if suser == "" {
			http.Error(w, "suser query param is required", 400)
			return
		}

		feed := &feeds.Feed{
			Title:       "Suser: " + suser,
			Link:        &feeds.Link{Href: "http://eksisozluk.com/biri/" + suser},
			Description: "Suser: " + suser,
			Created:     time.Now(),
		}

		c := colly.NewCollector()

		c.OnHTML("ul.topic-list", func(e *colly.HTMLElement) {
			e.ForEach("li > a", func(_ int, c *colly.HTMLElement) {
				href := c.Attr("href")
				if !strings.HasPrefix(href, "/entry/") {
					return
				}
                                re := regexp.MustCompile(`#[0-9]+`)
				text := strings.TrimSpace(re.ReplaceAllString(c.Text, ""))
				link := e.Request.AbsoluteURL(href)
				feed.Items = append(feed.Items, &feeds.Item{Title: text, Link: &feeds.Link{Href: link}, Description: text})
			})
		})

		c.Visit(fmt.Sprintf("https://eksisozluk.com/basliklar/istatistik/%s/son-entryleri", suser))

                feed.Items = feed.Items[:10]
		for _, f := range feed.Items {
                        fmt.Println("visiting", f.Link.Href)
			e := colly.NewCollector()
			e.OnHTML("div.content", func(c *colly.HTMLElement) {
				f.Description = c.Text
			})
			e.Visit(f.Link.Href)
		}

		rss, err := feed.ToRss()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintln(w, rss)
	})

	panic(http.ListenAndServe(":8080", nil))
}
