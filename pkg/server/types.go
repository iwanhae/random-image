package server

import (
	"encoding/base64"
	"path"

	"github.com/iwanhae/random-image/pkg/store"
	"github.com/labstack/echo"
)

type Message struct {
	Msg string `json:"msg"`
}

func ErrorResponse(c echo.Context, code int, err error) error {
	return c.JSON(code, Message{Msg: err.Error()})
}

type ObjectMeta struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

func CreateObjectMeta(v *store.ObjectMeta) ObjectMeta {

	return ObjectMeta{
		ID:   v.ID,
		Name: path.Base(v.Key),
		Group: base64.StdEncoding.EncodeToString(
			[]byte(path.Dir(v.Key)),
		),
	}
}
