package main

import (
	"bytes"
	"io"
	"os"

	"github.com/ulikunitz/xz"
)

func readFileXZ(path string) []byte {
	f, err := os.Open(path)
	panicOnError(err)

	return readXZ(f)
}

func readFilePlain(path string) []byte {
	f, err := os.Open(path)
	panicOnError(err)

	return readPlain(f)
}

func readXZ(in io.Reader) []byte {
	r, err := xz.NewReader(in)
	panicOnError(err)

	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)

	_, err = io.Copy(w, r)
	panicOnError(err)

	return w.Bytes()
}

func readPlain(in io.Reader) []byte {
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)

	_, err := io.Copy(w, in)
	panicOnError(err)

	return w.Bytes()
}

func writeXZ(path string, out []byte) {
	f, err := os.Create(path)
	panicOnError(err)
	defer f.Close()

	w, err := xz.NewWriter(f)
	panicOnError(err)
	defer w.Close()

	io.Copy(w, bytes.NewBuffer(out))
	panicOnError(err)
}

func writePlain(path string, out []byte) {
	f, err := os.Create(path)
	panicOnError(err)
	defer f.Close()

	io.Copy(f, bytes.NewBuffer(out))
	panicOnError(err)
}
