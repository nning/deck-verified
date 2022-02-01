.phony: clean

BIN_NAME = deck-verified
SOURCES = $(shell find . -name \*.go)

$(BIN_NAME): $(SOURCES)
	go build .

clean:
	rm -f $(BIN_NAME)