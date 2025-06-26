
.PHONY: build run clean

build:
	go build -o bin/monitor ./cmd/monitor

run: build
	./bin/monitor

clean:
	rm -rf bin
