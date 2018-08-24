build:
	@go build -o $(PWD)/om github.com/checkr/openmock/cmd/om

test:
	@go test -race -covermode=atomic .
