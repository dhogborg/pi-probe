default: build

.PHONY: build
build:
	go build -o build/probe *.go

PHONY: release
release:
	GOOS=linux GOARCH=arm go build -o build/probe *.go

.PHONY: deploy
deploy: release
	scp build/probe probe:~/probe
