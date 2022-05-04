package main

import (
	"os"
	"path"
	"strings"
	"time"
)

type QueryResponse struct {
	Results []struct {
		Hits []struct {
			Name        string   `json:"name"`
			OsList      []string `json:"oslist"`
			LastUpdated int      `json:"lastUpdated"`
			AppID       string   `json:"objectID"`
		} `json:"hits"`
		HitCount  int `json:"nbHits"`
		PageCount int `json:"nbPages"`
	} `json:"results"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type Entry struct {
	Name               string
	Status             string
	PreviousStatus     string
	FirstSeen          time.Time
	LastUpdatedSteamDB time.Time
	LastUpdatedHere    time.Time
	AppID              string
}

var cron bool
var debug bool
var quiet bool
var store Store

var home, _ = os.UserHomeDir()
var storeDir = path.Join(home, ".cache", "deck-verified")
var storePath = path.Join(storeDir, "store.json")

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func getVerificationStatus(oslist []string) string {
	x := "Steam Deck "

	for _, s := range oslist {
		if strings.HasPrefix(s, x) {
			return strings.Replace(s, x, "", 1)
		}
	}

	return ""
}

func init() {
	cron = os.Getenv("CRON") != ""
	debug = os.Getenv("DEBUG") != ""
	quiet = os.Getenv("QUIET") != ""

	store = getStore(false)
}

func main() {
	if len(os.Args) == 1 || len(os.Args) > 1 && os.Args[1] == "update" {
		cmdUpdate()
	}

	if len(os.Args) > 2 && os.Args[1] == "search" {
		term := strings.Join(os.Args[2:], " ")
		cmdSearch(term)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "list" {
			status := ""
			if len(os.Args) > 2 {
				status = os.Args[2]
			}

			cmdList(status)
		}

		if os.Args[1] == "feed" {
			if len(os.Args) > 2 && os.Args[2] == "serve" {
				cmdFeedServe(os.Args[3:]...)
			} else {
				cmdFeed()
			}
		}

		if os.Args[1] == "stats" {
			cmdStats()
		}
	}
}
