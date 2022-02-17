package main

import (
	"encoding/json"
	"errors"
	"os"
	"sort"

	"github.com/fxamacker/cbor"
)

type Store map[string]*Entry

// TODO Take optional less function, use sort by LastUpdatedHere as default but
//      allow sorting by Name in cmdSearch
func getEntriesFromStore(store *Store) []*Entry {
	entries := make([]*Entry, 0, len(*store))

	for _, entry := range *store {
		if entry.Name != "" {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUpdatedHere.After(entries[j].LastUpdatedHere)
	})

	return entries
}

func getStore(cborXZ bool) Store {
	store := make(Store)

	_, err := os.Stat(storePath)
	if errors.Is(err, os.ErrNotExist) {
		return store
	}

	if cborXZ {
		x := readFileXZ(storePath)
		err = cbor.Unmarshal(x, &store)
	} else {
		x := readFilePlain(storePath)
		err = json.Unmarshal(x, &store)
	}

	panicOnError(err)

	return store
}

func writeStore(cborXZ bool) {
	err := os.MkdirAll(storeDir, 0700)
	panicOnError(err)

	if cborXZ {
		out, err := cbor.Marshal(store, cbor.CanonicalEncOptions())
		panicOnError(err)

		writeXZ(storePath, out)
	} else {
		out, err := json.MarshalIndent(store, "", "  ")
		panicOnError(err)

		writePlain(storePath, out)
	}
}
