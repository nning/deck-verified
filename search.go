package main

import (
	"strings"

	"github.com/cheynewallace/tabby"
)

func cmdSearch(term string) {
	t := tabby.New()

	entries := getEntriesFromStore(&store, func(entries []*Entry) func(i, j int) bool {
		return func(i, j int) bool {
			return entries[i].Name < entries[j].Name
		}
	})

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), term) {
			t.AddLine(entry.Name, entry.Status, entry.LastUpdatedHere)
		}
	}

	t.Print()
}
