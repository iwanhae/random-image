package main

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/iwanhae/random-image/pkg/config"
	"github.com/iwanhae/random-image/pkg/store"
	"github.com/iwanhae/random-image/pkg/store/meta"
	"github.com/rs/zerolog/log"
)

const ()

func main() {
	ctx := context.Background()
	// config
	c := config.Init()
	log.Info().Interface("config", c).Msg("config loaded")

	mc, err := store.GetS3Client(c)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	moc, err := meta.InitMongoDB(ctx, c)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	st, err := store.NewObjectStore(moc, mc, c.S3BucketName)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	files, _ := ioutil.ReadDir(".vscode")
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		log.Info().Str("file", v.Name()).Msg("uploading")
		f, _ := os.Open(
			path.Join(".vscode", v.Name()),
		)
		uuid, err := st.PutObject(ctx, meta.ObjectMeta{
			Path: f.Name(),
		}, f)
		if err != nil {
			log.Error().Err(err).Send()
		}
		log.Info().Str("uuid", uuid).Str("file", v.Name()).Msg("uploaded")

		_, r, _ := st.GetReader(ctx, uuid)
		io.Copy(os.Stdout, r)
	}

}
