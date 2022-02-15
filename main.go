package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
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

type Store map[string]*Entry

//go:embed data/url
var urlData []byte

//go:embed data/request
var requestData []byte

//go:embed data/feeditem.tpl
var feedItemTemplate string

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

func readFileXZ(path string) []byte {
	f, err := os.Open(path)
	panicOnError(err)

	return readXZ(f)
}

func readFilePlain(path string) []byte {
	f, err := os.Open(path)
	panicOnError(err)

	return readPlain(f)
}

func readXZ(in io.Reader) []byte {
	r, err := xz.NewReader(in)
	panicOnError(err)

	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)

	_, err = io.Copy(w, r)
	panicOnError(err)

	return w.Bytes()
}

func readPlain(in io.Reader) []byte {
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)

	_, err := io.Copy(w, in)
	panicOnError(err)

	return w.Bytes()
}

func writeXZ(path string, out []byte) {
	f, err := os.Create(path)
	panicOnError(err)
	defer f.Close()

	w, err := xz.NewWriter(f)
	panicOnError(err)
	defer w.Close()

	io.Copy(w, bytes.NewBuffer(out))
	panicOnError(err)
}

func writePlain(path string, out []byte) {
	f, err := os.Create(path)
	panicOnError(err)
	defer f.Close()

	io.Copy(f, bytes.NewBuffer(out))
	panicOnError(err)
}

func searchSteamDB() []QueryResponse {
	u := readXZ(bytes.NewReader(urlData))
	b0 := readXZ(bytes.NewReader(requestData))

	responses := make([]QueryResponse, 0)
	pages := 1

	for page := 0; page < pages; page++ {
		r := requestPage(string(u), page, b0)

		if pages == 1 {
			if r.Status > 0 {
				panicOnError(errors.New(r.Message + " (" + strconv.FormatInt(int64(r.Status), 10) + ")"))
			}

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

func getStore(cborXZ bool) Store {
	store := make(Store)

	_, err := os.Stat(storePath)
	if errors.Is(err, os.ErrNotExist) {
		return store
	}

	if cborXZ {
		x := readFileXZ(storePath)
		err = cbor.Unmarshal(x, &store)
	} else {
		x := readFilePlain(storePath)
		err = json.Unmarshal(x, &store)
	}

	panicOnError(err)

	return store
}

func writeStore(cborXZ bool) {
	err := os.MkdirAll(storeDir, 0700)
	panicOnError(err)

	if cborXZ {
		out, err := cbor.Marshal(store, cbor.CanonicalEncOptions())
		panicOnError(err)

		writeXZ(storePath, out)
	} else {
		out, err := json.MarshalIndent(store, "", "  ")
		panicOnError(err)

		writePlain(storePath, out)
	}
}

func cmdUpdate() {
	t := tabby.New()

	responses := searchSteamDB()

	newCount := 0
	updatedCount := 0

	for _, response := range responses {
		for _, hit := range response.Results[0].Hits {
			lastUpdated := time.Unix(int64(hit.LastUpdated), 0)
			status := getVerificationStatus(hit.OsList)

			if hit.Name == "" {
				continue
			}

			if store[hit.Name] != nil {
				if !lastUpdated.After(store[hit.Name].LastUpdatedSteamDB) && store[hit.Name].Status == status {
					continue
				}

				if store[hit.Name].Status != status {
					store[hit.Name].LastUpdatedHere = time.Now()
				}

				store[hit.Name].PreviousStatus = store[hit.Name].Status
				store[hit.Name].Status = status
				store[hit.Name].LastUpdatedSteamDB = lastUpdated

				prev := ""
				if status != store[hit.Name].PreviousStatus && store[hit.Name].PreviousStatus != "" {
					prev = "(previously: " + store[hit.Name].PreviousStatus + ")"
				}

				t.AddLine("Updated", hit.Name, status, prev)
				updatedCount = updatedCount + 1

				continue
			}

			now := time.Now()

			store[hit.Name] = &Entry{
				Name:               hit.Name,
				Status:             status,
				PreviousStatus:     status,
				FirstSeen:          now,
				LastUpdatedSteamDB: lastUpdated,
				LastUpdatedHere:    now,
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

	writeStore(!debug)

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
			t.AddLine(name, data.Status)
		}
	}

	t.Print()
}

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

	entries := make([]*Entry, 0, l)

	for _, entry := range store {
		if entry.Name != "" {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUpdatedHere.After(entries[j].LastUpdatedHere)
	})

	feed := &feeds.Feed{
		Title:   "Steam Deck Verified",
		Created: time.Now(),
		Link:    &feeds.Link{Href: "http://github.com/nning/deck-verified"},
	}

	tpl, err := template.New("feedItem").Parse(feedItemTemplate)
	panicOnError(err)

	for _, entry := range entries[0:n] {
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
		store = getStore(!debug)
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

func init() {
	cron = os.Getenv("CRON") != ""
	debug = os.Getenv("DEBUG") != ""
	quiet = os.Getenv("QUIET") != ""

	if debug {
		storePath = storePath + ".json"
	}

	store = getStore(!debug)
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
