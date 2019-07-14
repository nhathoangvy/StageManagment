package main

import (
	"net/http"
	"github.com/labstack/echo"
	"encoding/json"
	"time"
	"strings"
	"io/ioutil"
	"log"
	"strconv"
	"fmt"
	"bytes"
)

func HealthCheck(c echo.Context) error{
	return c.String(http.StatusOK, "OK")
}

func SessionPlayList(c echo.Context) error {
	var auth map[string]interface{}
	if err := json.Unmarshal([]byte(passport), &auth); err != nil {
		log.Fatal(err.Error())
	}
	sessionId := auth["sid"].(string)
	data, err := sessionRepository.Get("auth_id", sessionId)
	if err != nil {
		log.Fatal(err.Error())
	}

	return c.JSON(http.StatusOK, data)
}

func SessionPlayInit(c echo.Context) error {
	modelRequest := new(SessionPlayInitResquest)
	var raw, auth map[string]interface{}
	var bodyBytes []byte
	if err := json.Unmarshal([]byte(passport), &auth); err != nil {
		panic(err)
	}
	if c.Request().Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	}
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		if err := c.Bind(modelRequest); err != nil{
			return errBadRequest
		}
	}
	request := modelRequest
	if raw != nil{
		request.MovieId = raw["movieId"].(string)
		request.Total = raw["total"].(string)
	}
	
	sessionId	:= auth["sid"].(string)
	userId		:= auth["uid"].(string)
	movieId		:= request.MovieId
	deviceId	:= auth["did"].(string)
	platform	:= auth["plt"].(string)
	total, _	:= strconv.ParseInt(request.Total, 10, 64)
	clientIp	:= auth["x-forwarded-for"].(string)
	rejected	:= auth["jti"].(string)

	data, err := sessionRepository.Create(SessionPlayCreate{
		sessionId, userId, movieId, deviceId, platform, total, clientIp, rejected,
	})
	if err != nil {
		log.Fatal(err.Error())
		return errBadRequest
	}

	return c.JSON(http.StatusOK, data)
}

func SessionPlayActived(c echo.Context) error {
	expired := time.Now().Add(time.Hour * 72)
	movieId := c.Param("movieId")
	userId := c.Param("userId")
	sessionPlayId := c.Param("playId")
	success, err := sessionRepository.Update(Where{Active: InActive, PlayId: sessionPlayId}, SessionPlayUpdate{Active: Actived})
	if err != nil || success.Code != 0 {
		return errBadRequest
	}
	result := map[string]interface{}{
		"accountingId": userId,
		"assetId": movieId,
	}
	return c.JSON(http.StatusOK, result)
}

func SessionPlayWatching(c echo.Context) error {
	h := Handler{}
	modelRequest := new(SessionPlayWatchingResquest)
	var raw map[string]interface{}
	var bodyBytes []byte

	if c.Request().Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	}
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		if err := c.Bind(modelRequest); err != nil{
			return errBadRequest
		}
	}
	request := modelRequest
	if raw != nil{
		request.PlayId	= c.Param("playId")
		request.Subtitle = raw["process"].(string)
		request.
	}
	inputData := fmt.Sprint(`{"pid":"%s","pcs":"%s"}`, request.PlayId, request.Process)
	h.Send(inputData)
	return c.JSON(http.StatusOK, "OK")
}

func SessionPlayKickout(c echo.Context) error {
	var auth map[string]interface{}
	if err := json.Unmarshal([]byte(passport), &auth); err != nil {
		log.Fatal(err.Error())
		return errBadRequest
	}

	sessionId := auth["sid"].(string)

	data, err := sessionRepository.Update(Where{SessionId: sessionId}, SessionPlayUpdate{Status: StatusKicked})
	if err != nil {
		log.Fatal(err.Error())
		return errBadRequest
	}
	return c.JSON(http.StatusOK, data)
}

func GetPlayList(c echo.Context) error {
	var auth, result, playlist_allow, playlistData map[string]interface{}
	var zips []map[string]interface{}
	var ErrAuth interface{}

	if err := json.Unmarshal([]byte(passport), &auth); err != nil {
		log.Fatal(err.Error())
	}
	userId		:= auth["uid"].(string)
	movieId		:= c.Param("movieId")
	playId		:= c.Param("playId")
	platform	:= auth["plt"].(string)
	deviceId	:= auth["did"].(string)
	sessionId	:= auth["sid"].(string)

	if _, ErrAuth := handler.CheckSubscriptions(userId, movieId, platform, deviceId); ErrAuth != ""{
		return errBadRequest
	}

	if _, ErrAuth = handler.UserPermissions(userId, movieId, deviceId); ErrAuth != ""{
		return errBadRequest
	}

	fcdn, err := local.GetCache("Streaming:"+movieId)
	if err != nil{
		return errBadRequest
	}
	tasks := []string{
		`SELECT * FROM movie WHERE id = '%s' AND status = 1`,
		`SELECT * FROM movie_subs WHERE id = '%s'`,
		`SELECT * FROM movie_source WHERE id = '%s'`,
		`SELECT * FROM movie_snapshot WHERE id = '%s'`,
		`SELECT * FROM movie_marker WHERE id = '%s'`,
	}
	task := make(chan []map[string]interface{}, len(tasks))
	go func() {
		for _, v := range tasks{
			data := otherRepository.Sequelize(CmDB, fmt.Sprintf(v, movieId))
			task <- data
		}
		close(task)
	}()

	movies, subtitle, source, snapshots, marker := <-task, <-task, <-task, <-task, <-task

	result := fmt.Sprintf(`{
		"movies": "%s",
		"subtitle": "%s",
		"snapshotImgs": "%s",
		"snapshotZips": "%s",
		"marker": "%s",
		"playlist":"%s"
	}`,movies, subtitle, source, snapshots, marker )
	if err := json.Unmarshal([]byte(keys), &result); err != nil{
		log.Fatal(err.Error())
	}
	return c.JSON(http.StatusOK, result)
}