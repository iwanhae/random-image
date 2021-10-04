package server

import (
	"strconv"

	"github.com/iwanhae/random-image/pkg/config"
	"github.com/iwanhae/random-image/pkg/store"
	"github.com/labstack/echo"
)

func NewServer(conf config.RandomImageConfig, store *store.ObjectStore) *echo.Echo {
	e := echo.New()
	e.Use(RequestIDGenerator)
	e.Use(LoggerMiddleware)

	e.GET("/api/sample", func(c echo.Context) error {
		return nil
	})

	e.GET("/data/:id", func(c echo.Context) error {
		return nil
	})

	e.GET("/api/group/:id", func(c echo.Context) error {
		return nil
	})
	return e
}

func QueryParamInt(c echo.Context, name string, def int) int {
	param := c.QueryParam(name)
	result, err := strconv.Atoi(param)
	if err != nil {
		return def
	}
	return result
}

/*
var sem = make(chan int, runtime.NumCPU())

func WebpConverter(ctx context.Context, w io.Writer, img image.Image, width, height, quality int) error {
	sem <- 1
	defer func() {
		<-sem
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("Context canceled")
	default:
	}
	if width != 0 || height != 0 {
		if width == 0 {
			width = height * 3
		}
		if height == 0 {
			height = width * 3
		}
		img = resize.Thumbnail(uint(width), uint(height), img, resize.Lanczos3)
	}
	opt, _ := encoder.NewLossyEncoderOptions(encoder.PresetPhoto, float32(quality))
	return webp.Encode(w, img, opt)
}
*/
