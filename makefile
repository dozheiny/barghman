.PHONY: build
build:
	go build -o barghman ./...

.PHONY: clean
clean:
	rm -rf ~/.cache/barghman/*