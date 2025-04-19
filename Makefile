PACKAGE = github.com/ramonmedeiros/rsvp
SHORT_HASH=$(shell git rev-parse --short HEAD)
build: bin/rsvp
bin/rsvp:
	GOOS=linux GOARCH=amd64 go build -o bin/rsvp cmd/main.go

docker:
	docker build . -t rsvp:$(SHORT_HASH)
