package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/rs/zerolog/log"
)

type ContextKey struct {
	Name string
}

var (
	RequestIDKey = &ContextKey{
		Name: "request_id",
	}
)

func GetRequestID(ctx context.Context) string {
	return fmt.Sprintf("%v", ctx.Value(RequestIDKey))
}
func ReqeuestIDWithContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func RequestIDGenerator(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		ctx = ReqeuestIDWithContext(ctx, uuid.New().String())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		rid := GetRequestID(ctx)
		logger := log.Logger.With().Str("request_id", rid).Logger()
		ctx = logger.WithContext(ctx)
		c.SetRequest(c.Request().WithContext(ctx))

		t := time.Now()
		if err := next(c); err != nil {
			c.Error(err)
		}
		log.Ctx(ctx).Info().
			Str("method", c.Request().Method).
			Str("ip", c.RealIP()).
			Str("path", c.Request().URL.Path).
			Int64("duration", time.Since(t).Milliseconds()).
			Send()
		return nil
	}
}
