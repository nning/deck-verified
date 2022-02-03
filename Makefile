.phony: clean

BIN_NAME = deck-verified
SOURCES = $(shell find . -name \*.go)
PREFIX = ~/.local/bin

build: $(BIN_NAME)

$(BIN_NAME): $(SOURCES)
	CGO_ENABLED=0 go build .

run: $(BIN_NAME)
	./$(BIN_NAME)

install: build
	mkdir -p $(PREFIX)
	cp $(BIN_NAME) $(PREFIX)

clean:
	rm -f $(BIN_NAME) data/response*.json