package main

import (
	_ "github.com/aaronland/go-sqlite-mattn"
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"log"

	"github.com/whosonfirst/go-whosonfirst-sqlite-features-index/v2/app/index"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := index.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to index, %v", err)
	}
}