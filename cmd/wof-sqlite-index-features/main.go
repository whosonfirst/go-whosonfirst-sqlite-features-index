package main

// To do: Move all of this in to app/query

import (
	_ "github.com/aaronland/go-sqlite-modernc"
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features-index/v2/app/index"
	"log"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := index.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to index, %v", err)
	}
}
