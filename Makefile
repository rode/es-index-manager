.PHONY: test fmtcheck vet fmt mocks coverage license
GOFMT_FILES?=$$(find . -name '*.go')
MAKEFLAGS+=--silent

fmtcheck:
	lineCount=$(shell gofmt -l -s $(GOFMT_FILES) | wc -l | tr -d ' ') && exit $$lineCount

fmt:
	gofmt -w -s $(GOFMT_FILES)

vet:
	go vet ./...

mocks:
	go generate ./...

test: fmtcheck vet
	go test -short ./... -coverprofile=coverage.txt -covermode atomic

coverage: test
	go tool cover -html=coverage.txt

license:
	addlicense -c 'The Rode Authors' $(GOFMT_FILES)