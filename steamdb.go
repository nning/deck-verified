package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//go:embed data/url
var urlData []byte

//go:embed data/request.json
var requestData []byte

func requestPage(url string, status string, page, priceMin, priceMax int, requestTemplate []byte) QueryResponse {
	strPage := strconv.FormatInt(int64(page), 10)
	strPriceMin := strconv.FormatInt(int64(priceMin), 10)
	strPriceMax := strconv.FormatInt(int64(priceMax), 10)

	b1 := bytes.Replace(requestTemplate, []byte("${page}"), []byte(strPage), 1)
	b1 = bytes.Replace(b1, []byte("${price_min}"), []byte(strPriceMin), 1)
	b1 = bytes.Replace(b1, []byte("${price_max}"), []byte(strPriceMax), 1)
	b1 = bytes.Replace(b1, []byte("${status}"), []byte(status), 1)

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
		name := strings.Join([]string{"response", status, strPriceMin, strPriceMax, strPage}, "-")
		err = os.WriteFile("data/"+name+".json", body, 0600)
		panicOnError(err)
	}

	var r QueryResponse
	err = json.Unmarshal(body, &r)
	panicOnError(err)

	return r
}

func searchSteamDB() []QueryResponse {
	u := readPlain(bytes.NewReader(urlData))
	b0 := readPlain(bytes.NewReader(requestData))

	responses := make([]QueryResponse, 0)

	// TODO: Improve price ranges, e.g. 0-9.99, 10-19.99, ...
	// TODO: Fire a number of requests concurrently
	for _, status := range []string{"Verified", "Playable", "Unsupported"} {
		for priceMin := 0; priceMin <= 120; priceMin += 10 {
			priceMax := priceMin + 10
			pages := 1

			// TODO: Request first page, then request other pages in parallel
			for page := 0; page < pages; page++ {
				r := requestPage(string(u), status, page, priceMin, priceMax, b0)

				if pages == 1 {
					if r.Status > 0 {
						panicOnError(errors.New(r.Message + " (" + strconv.FormatInt(int64(r.Status), 10) + ")"))
					}

					pages = r.Results[0].PageCount
				}

				responses = append(responses, r)
			}
		}
	}

	return responses
}
