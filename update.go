package main

import (
	"fmt"
	"time"

	"github.com/cheynewallace/tabby"
)

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

			entry := store[hit.AppID]

			if entry != nil {
				if !lastUpdated.After(entry.LastUpdatedSteamDB) && entry.Status == status {
					continue
				}

				if entry.Status != status {
					entry.LastUpdatedHere = time.Now()
				}

				entry.PreviousStatus = entry.Status
				entry.Status = status
				entry.LastUpdatedSteamDB = lastUpdated

				prev := ""
				if status != entry.PreviousStatus && entry.PreviousStatus != "" {
					prev = "(previously: " + entry.PreviousStatus + ")"
				}

				t.AddLine("Updated", hit.Name, status, prev)
				updatedCount = updatedCount + 1

				continue
			}

			now := time.Now()

			store[hit.AppID] = &Entry{
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
		fmt.Printf("Total: %v, New: %v, Updated: %v\n", len(store), newCount, updatedCount)
	}
}
