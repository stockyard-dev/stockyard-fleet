build:
	CGO_ENABLED=0 go build -o fleet ./cmd/fleet/

run: build
	./fleet

test:
	go test ./...

clean:
	rm -f fleet

.PHONY: build run test clean
