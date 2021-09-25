package server

import (
	"encoding/base64"
	"io"

	"github.com/iwanhae/random-image/pkg/store"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/minio/minio-go/v7"
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
		_, err = io.Copy(c.Response(), obj)
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