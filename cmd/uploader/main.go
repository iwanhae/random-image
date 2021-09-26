package main

import (
	"context"

	"github.com/iwanhae/random-image/pkg/config"
	"github.com/iwanhae/random-image/pkg/storage"
)

func main() {
	c := config.Init()

	mc, err := storage.GetS3Client(c)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	mc.ListBuckets(ctx)
}
