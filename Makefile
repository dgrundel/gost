.PHONY: build test clean

build:
	go build .

test:
	go test ./...

clean:
	rm -f guts

integ: build
	./guts integ/**/*.guts
