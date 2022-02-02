package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type QueryResponse struct {
	Results []struct {
		Hits []struct {
			Name   string   `json:"name"`
			OsList []string `json:"oslist"`
		} `json:"hits"`
		HitCount  int `json:"nbHits"`
		PageCount int `json:"nbPages"`
	} `json:"results"`
}

type Entry struct {
	Status    string
	FirstSeen time.Time
}

type Store map[string]*Entry

const urlPath = "data/url"
const requestPath = "data/request.json"
const storePath = "data/store.json"

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

	var r QueryResponse
	err = json.Unmarshal(body, &r)
	panicOnError(err)

	return r
}

func searchSteamDB() []QueryResponse {
	u, err := os.ReadFile(urlPath)
	panicOnError(err)

	b0, err := os.ReadFile(requestPath)
	panicOnError(err)

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
		in, err := os.ReadFile(storePath)
		panicOnError(err)

		err = json.Unmarshal(in, &store)
		panicOnError(err)
	}

	return store
}

func main() {
	responses := searchSteamDB()
	store := getStore()
	i := 0

	for _, response := range responses {
		for _, hit := range response.Results[0].Hits {
			if store[hit.Name] != nil {
				continue
			}

			status := getVerificationStatus(hit.OsList)
			store[hit.Name] = &Entry{status, time.Now()}
			fmt.Printf("%v: %v\n", hit.Name, status)
			i = i + 1
		}
	}

	out, err := json.MarshalIndent(store, "", "   ")
	panicOnError(err)

	err = os.WriteFile(storePath, out, 0600)
	panicOnError(err)

	fmt.Printf("\nTotal Hits: %v, New Games: %v\n", responses[0].Results[0].HitCount, i)
}
