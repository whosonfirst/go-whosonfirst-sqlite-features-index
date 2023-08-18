GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-sqlite-index-features cmd/wof-sqlite-index-features/main.go
