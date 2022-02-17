package main

import (
	"encoding/json"
	"errors"
	"os"
	"sort"

	"github.com/fxamacker/cbor"
)

type Store map[string]*Entry

func lessByNameAsc(entries []*Entry) func(i, j int) bool {
	return func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	}
}

func lessByLastUpdatedHereDesc(entries []*Entry) func(i, j int) bool {
	return func(i, j int) bool {
		return entries[i].LastUpdatedHere.After(entries[j].LastUpdatedHere)
	}
}

func getEntriesFromStore(store *Store, funcs ...func(entries []*Entry) func(i, j int) bool) []*Entry {
	entries := make([]*Entry, 0, len(*store))

	for _, entry := range *store {
		if entry.Name != "" {
			entries = append(entries, entry)
		}
	}

	var lessFunc func(i, j int) bool
	if len(funcs) > 0 {
		lessFunc = funcs[0](entries)
	} else {
		lessFunc = lessByLastUpdatedHereDesc(entries)
	}

	sort.Slice(entries, lessFunc)

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
