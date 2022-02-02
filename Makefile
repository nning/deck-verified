.phony: clean

BIN_NAME = deck-verified
SOURCES = $(shell find . -name \*.go)

$(BIN_NAME): $(SOURCES)
	go build .

run: $(BIN_NAME)
	./$(BIN_NAME)

clean:
	rm -f $(BIN_NAME) data/response*.json