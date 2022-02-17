package main

import (
	"strings"

	"github.com/cheynewallace/tabby"
)

func cmdList(status string) {
	t := tabby.New()

	entries := getEntriesFromStore(&store, lessByNameAsc)

	for _, entry := range entries {
		if status == "" || status == strings.ToLower(entry.Status) {
			t.AddLine(entry.Name, entry.Status)
		}
	}

	t.Print()
}
