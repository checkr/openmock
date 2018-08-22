build:
	go build

test:
	go test -race -covermode=atomic .
