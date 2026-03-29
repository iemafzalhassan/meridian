build:
	go build -o bin/meridian ./cmd/meridian

run: build
	./bin/meridian

ingest: build
	./bin/meridian ingest --limit 50

qdrant:
	docker compose up -d qdrant

test:
	go test ./...

lint:
	golangci-lint run

release:
	goreleaser release
