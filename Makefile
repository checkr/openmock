build:
	go build -o $(PWD)/om github.com/checkr/openmock/cmd/om

test:
	go test -race -covermode=atomic .

run: build
	OPENMOCK_TEMPLATES_DIR=./demo_templates ./om
