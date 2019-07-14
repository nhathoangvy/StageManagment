package main

import (
	"github.com/labstack/echo"
)

type (	

	Error struct {
		Code 	int 	`json:"errorCode"`
		Message string	`json:"message"`
	}

	Success struct {
		Code	int 	`json:"resultCode"`
		Message string	`json:"message"`
	}

)

var (
	errBadRequest       = echo.NewHTTPError(400, Error{2000, "Bad Request"})
	errSessionExpired   = echo.NewHTTPError(401, Error{2001, "Session Expired"})
	errPermission 		= echo.NewHTTPError(403, Error{2002, "Unauthorized"})
	errNotFound 		= echo.NewHTTPError(404, Error{2003, "Not Found"})
)