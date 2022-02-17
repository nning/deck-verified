package main

import (
	"strings"

	"github.com/cheynewallace/tabby"
)

func cmdList(status string) {
	t := tabby.New()

	for name, data := range store {
		if status == "" || status == strings.ToLower(data.Status) {
			t.AddLine(name, data.Status)
		}
	}

	t.Print()
}
