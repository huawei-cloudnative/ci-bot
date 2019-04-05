.PHONY: all build ci-bot clean

all:build

build:ci-bot

build-image:ci-bot-image

ci-bot:
	go build

ci-bot-image:
	docker build -t cibot:latest ./	

clean:
	rm -rf ./ci-bot
