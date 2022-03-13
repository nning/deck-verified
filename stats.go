package main

import (
	"time"

	"github.com/cheynewallace/tabby"
)

func cmdStats() {
	total := len(store)
	a, b, c := 0, 0, 0

	var d time.Time

	for _, entry := range store {
		switch entry.Status {
		case "Verified":
			a += 1
		case "Playable":
			b += 1
		case "Unsupported":
			c += 1
		}
	}

	entries := getEntriesFromStore(&store)
	d = entries[0].LastUpdatedHere

	t := tabby.New()

	t.AddLine("Total", total)
	t.AddLine("Verified", a)
	t.AddLine("Playable", b)
	t.AddLine("Verified+Playable", b+c)
	t.AddLine("Unsupported", c)
	t.AddLine("Last Updated", d)

	t.Print()
}
