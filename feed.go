package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
)

//go:embed data/feeditem.tpl
var feedItemTemplate string

func generateFeedItemContent(tpl *template.Template, entry *Entry) string {
	buf := new(bytes.Buffer)
	err := tpl.Execute(buf, entry)
	panicOnError(err)

	return buf.String()
}

func generateFeed(n int) *feeds.Feed {
	l := len(store)
	if n == -1 {
		n = l
	}

	entries := getEntriesFromStore(&store)

	feed := &feeds.Feed{
		Title:   "Steam Deck Verified",
		Created: time.Now(),
		Link:    &feeds.Link{Href: "http://github.com/nning/deck-verified"},
	}

	tpl, err := template.New("feedItem").Parse(feedItemTemplate)
	panicOnError(err)

	for _, entry := range entries[0:n] {
		if entry == nil {
			continue
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       "[" + entry.Status + "] " + entry.Name,
			Link:        &feeds.Link{Href: "https://steamdb.info/app/" + entry.AppID + "/info/"},
			Updated:     entry.LastUpdatedHere,
			Description: generateFeedItemContent(tpl, entry),
		})
	}

	return feed
}

func cmdFeed() {
	feed := generateFeed(-1)
	content, err := feed.ToAtom()
	panicOnError(err)

	fmt.Println(content)
}

func cmdFeedServe(args ...string) {
	http.HandleFunc("/feed.xml", func(w http.ResponseWriter, r *http.Request) {
		store = getStore(false)
		feed := generateFeed(-1)
		content, err := feed.ToAtom()
		panicOnError(err)

		fmt.Fprintf(w, content)
	})

	port := "8080"
	if len(args) > 0 {
		port = args[0]
	}

	fmt.Println("http://localhost:" + port + "/feed.xml")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
