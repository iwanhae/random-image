package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"runtime"
	"strconv"

	"github.com/iwanhae/random-image/pkg/store"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
)

func NewServer(mc minio.Client, db *store.Database) *echo.Echo {
	e := echo.New()
	e.Use(RequestIDGenerator)
	e.Use(LoggerMiddleware)

	e.GET("/api/sample", func(c echo.Context) error {
		ctx := c.Request().Context()
		objs, err := db.SampleObject(ctx, 50)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed at object sampling")
			return ErrorResponse(c, 500, err)
		}

		response := []ObjectMeta{}
		for _, v := range objs {
			response = append(response, CreateObjectMeta(v))
		}
		return c.JSONPretty(200, response, "\t")
	})

	e.GET("/data/:id", func(c echo.Context) error {
		ctx := c.Request().Context()
		id := c.Param("id")
		width := QueryParamInt(c, "w", 0)
		height := QueryParamInt(c, "h", 0)
		quality := QueryParamInt(c, "q", 0)
		objMeta, err := db.GetObjectMeta(ctx, id)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("id not found")
			return ErrorResponse(c, 404, err)
		}
		obj, err := mc.GetObject(ctx, "images", objMeta.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("fail to get obj")
			return ErrorResponse(c, 500, err)
		}

		if quality == 0 {
			_, err = io.Copy(c.Response(), obj)
		} else {
			img, err := jpeg.Decode(obj)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("fail to decode jpeg")
				return ErrorResponse(c, 500, err)
			}
			err = WebpConverter(ctx, c.Response(), img, width, height, quality)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("fail to encode webp")
				return ErrorResponse(c, 500, err)
			}
		}

		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("fail to send obj")
			return ErrorResponse(c, 500, err)
		}
		return nil
	})

	e.GET("/api/group/:id", func(c echo.Context) error {
		ctx := c.Request().Context()
		decoded, err := base64.StdEncoding.DecodeString(
			c.Param("id"),
		)
		if err != nil {
			return ErrorResponse(c, 401, err)
		}
		group := string(decoded)
		objs, err := db.ListObjectMetaByGroup(ctx, group)
		if err != nil {
			return ErrorResponse(c, 404, err)
		}
		response := []ObjectMeta{}
		for _, obj := range objs {
			response = append(response,
				CreateObjectMeta(obj),
			)
		}

		return c.JSONPretty(200, response, "\t")
	})
	e.Use(
		middleware.StaticWithConfig(middleware.StaticConfig{
			Skipper: middleware.DefaultSkipper,
			Root:    "web/out",
			Index:   "index.html",
			HTML5:   true,
		}))
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
