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
		fmt.Printf("Total: %v, New: %v, Updated: %v\n", len(store), newCount, updatedCount)
	}
}
