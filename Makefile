.PHONY: all build ci-bot clean

all:build

build:ci-bot

ci-bot:
	go build
clean:
	rm -rf ./ci-bot
