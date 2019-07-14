package main

import (
	"encoding/json"
	"github.com/labstack/echo"
)

var passport string

func Headers(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		passport = c.Request().Header.Get("Authorization")
		var auth map[string]interface{}
		if err := json.Unmarshal([]byte(passport), &auth); err != nil {
			panic(err)
		}
		sessionId := auth["sid"]
		userId := auth["uid"]
		if sessionId == nil || userId == nil {
			return errPermission
		}
		return next(c)
	}
}