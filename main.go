package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type QueryResponse struct {
	Results []struct {
		Hits []struct {
			Name   string   `json:"name"`
			OsList []string `json:"oslist"`
		} `json:"hits"`
	} `json:"results"`
}

func main() {
	u, err := os.ReadFile("data/url")
	if err != nil {
		panic(err)
	}

	b0, err := os.ReadFile("data/body.json")
	if err != nil {
		panic(err)
	}

	b1 := bytes.NewReader(b0)
	client := &http.Client{}

	req, err := http.NewRequest("POST", string(u), b1)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", "https://steamdb.info/")
	// req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:96.0) Gecko/20100101 Firefox/96.0")
	// req.Header.Add("Accept", "*/*")
	// req.Header.Add("Origin", "https://steamdb.info/")
	// req.Header.Add("Sec-Fetch-Dest", "empty")
	// req.Header.Add("Sec-Fetch-Mode", "cors")
	// req.Header.Add("Sec-Fetch-Site", "cross-site")
	// req.Header.Add("Pragma", "no-cache")
	// req.Header.Add("Cache-Control", "no-cache")

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	os.WriteFile("data/response.json", body, 0600)

	var results QueryResponse
	json.Unmarshal(body, results)

	fmt.Println(results)
}
