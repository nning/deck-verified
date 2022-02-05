package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cheynewallace/tabby"
	"github.com/fxamacker/cbor"
	"github.com/gorilla/feeds"
	"github.com/ulikunitz/xz"
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
}

type Entry struct {
	Name               string
	Status             string
	FirstSeen          time.Time
	LastUpdatedSteamDB time.Time
	LastUpdatedHere    time.Time
	AppID              string
}

type Store map[string]*Entry

//go:embed data/url
var urlData []byte

//go:embed data/request
var requestData []byte

var cron bool
var debug bool
var quiet bool
var store Store

var home, _ = os.UserHomeDir()
var storeDir = path.Join(home, ".cache", "deck-verified")
var storePath = path.Join(storeDir, "store")

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func requestPage(url string, page int, requestTemplate []byte) QueryResponse {
	strPage := strconv.FormatInt(int64(page), 10)
	b1 := bytes.Replace(requestTemplate, []byte("${page}"), []byte(strPage), 1)
	b2 := bytes.NewReader(b1)

	req, err := http.NewRequest("POST", url, b2)
	panicOnError(err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", "https://steamdb.info/")

	client := &http.Client{}
	res, err := client.Do(req)
	panicOnError(err)

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	panicOnError(err)

	if debug {
		err = os.WriteFile("data/response"+strconv.FormatInt(int64(page), 10)+".json", body, 0600)
		panicOnError(err)
	}

	var r QueryResponse
	err = json.Unmarshal(body, &r)
	panicOnError(err)

	return r
}

func readFile(path string) []byte {
	f, err := os.Open(path)
	panicOnError(err)

	return read(f)
}

func read(in io.Reader) []byte {
	r, err := xz.NewReader(in)
	panicOnError(err)

	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)

	_, err = io.Copy(w, r)
	panicOnError(err)

	return w.Bytes()
}

func write(path string, out []byte) {
	f, err := os.Create(path)
	panicOnError(err)
	defer f.Close()

	w, err := xz.NewWriter(f)
	panicOnError(err)
	defer w.Close()

	io.Copy(w, bytes.NewBuffer(out))
	panicOnError(err)
}

func searchSteamDB() []QueryResponse {
	u := read(bytes.NewReader(urlData))
	b0 := read(bytes.NewReader(requestData))

	responses := make([]QueryResponse, 0)
	pages := 1

	for page := 0; page < pages; page++ {
		r := requestPage(string(u), page, b0)

		if pages == 1 {
			pages = r.Results[0].PageCount
		}

		responses = append(responses, r)
	}

	return responses
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

func getStore() Store {
	store := make(Store)

	if _, err := os.Stat(storePath); !errors.Is(err, os.ErrNotExist) {
		x := readFile(storePath)
		err = cbor.Unmarshal(x, &store)
		panicOnError(err)
	}

	return store
}

func cmdUpdate() {
	t := tabby.New()

	responses := searchSteamDB()

	newCount := 0
	updatedCount := 0

	for _, response := range responses {
		for _, hit := range response.Results[0].Hits {
			lastUpdated := time.Unix(int64(hit.LastUpdated), 0)

			if store[hit.Name] != nil {
				status := getVerificationStatus(hit.OsList)
				if lastUpdated.After(store[hit.Name].LastUpdatedSteamDB) || store[hit.Name].Status != status {
					store[hit.Name].Status = status
					store[hit.Name].LastUpdatedSteamDB = lastUpdated
					store[hit.Name].LastUpdatedHere = time.Now()

					t.AddLine("Updated", hit.Name, status)
					updatedCount = updatedCount + 1
				}

				continue
			}

			status := getVerificationStatus(hit.OsList)
			store[hit.Name] = &Entry{
				Name:               hit.Name,
				Status:             status,
				FirstSeen:          time.Now(),
				LastUpdatedSteamDB: lastUpdated,
				LastUpdatedHere:    time.Now(),
				AppID:              hit.AppID,
			}
			t.AddLine("New", hit.Name, status)
			newCount = newCount + 1
		}
	}

	if !quiet {
		t.Print()
		if newCount+updatedCount > 0 {
			fmt.Println()
		}
	}

	out, err := cbor.Marshal(store, cbor.CanonicalEncOptions())
	panicOnError(err)

	err = os.MkdirAll(storeDir, 0700)
	panicOnError(err)

	write(storePath, out)

	if !quiet && (!cron || cron && (newCount > 0 || updatedCount > 0)) {
		fmt.Printf("Total: %v, New: %v, Updated: %v\n", responses[0].Results[0].HitCount, newCount, updatedCount)
	}
}

func cmdSearch(term string) {
	t := tabby.New()

	for name, data := range store {
		if strings.Contains(strings.ToLower(name), term) {
			t.AddLine(name, data.Status, data.LastUpdatedSteamDB)
		}
	}

	t.Print()
}

func cmdList(status string) {
	t := tabby.New()

	for name, data := range store {
		if status == "" || status == strings.ToLower(data.Status) {
			t.AddLine(name, data.Status, data.LastUpdatedSteamDB)
		}
	}

	t.Print()
}

func generateFeed(n int) *feeds.Feed {
	l := len(store)
	if n == -1 {
		n = l
	}

	entries := make([]*Entry, 0, l)

	for _, entry := range store {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUpdatedHere.After(entries[j].LastUpdatedHere)
	})

	feed := &feeds.Feed{
		Title:   "Steam Deck Verified",
		Created: time.Now(),
		Link:    &feeds.Link{Href: "http://nning.io/deck-verified"},
	}

	for _, entry := range entries[0:n] {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:   "[" + entry.Status + "] " + entry.Name,
			Link:    &feeds.Link{Href: "https://steamdb.info/app/" + entry.AppID + "/info/"},
			Created: entry.LastUpdatedHere,
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
		store = getStore()
		feed := generateFeed(15)
		content, err := feed.ToAtom()
		panicOnError(err)

		fmt.Fprintf(w, content)
	})

	port := "8080"
	if len(args) > 0 {
		port = args[0]
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func init() {
	cron = os.Getenv("CRON") != ""
	debug = os.Getenv("DEBUG") != ""
	quiet = os.Getenv("QUIET") != ""

	store = getStore()
}

func main() {
	if len(os.Args) == 1 || len(os.Args) > 1 && os.Args[1] == "update" {
		cmdUpdate()
	}

	if len(os.Args) > 2 && os.Args[1] == "search" {
		term := strings.Join(os.Args[2:], " ")
		cmdSearch(term)
	}

	if len(os.Args) > 1 && os.Args[1] == "list" {
		status := ""
		if len(os.Args) > 2 {
			status = os.Args[2]
		}

		cmdList(status)
	}

	if len(os.Args) > 1 && os.Args[1] == "feed" {
		if len(os.Args) > 2 && os.Args[2] == "serve" {
			cmdFeedServe(os.Args[3:]...)
		} else {
			cmdFeed()
		}
	}
}
