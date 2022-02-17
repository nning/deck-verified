package main

import (
	"strings"

	"github.com/cheynewallace/tabby"
)

func cmdSearch(term string) {
	t := tabby.New()

	// TODO Sort by Name (using getEntriesFromStore with custom less function)
	for name, data := range store {
		if strings.Contains(strings.ToLower(name), term) {
			t.AddLine(name, data.Status, data.LastUpdatedSteamDB)
		}
	}

	t.Print()
}
