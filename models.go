package main

import 	(
	"time"
)

const (
	ProcessInit		= 0
	InActive		= 0
	Actived			= 1
	StatusKicked	= -2
	StatusExpired	= -1
	StatusInit		= 0
	StatusPlaying	= 1
	StatusFinished  = 2
	exp = 24*60*60*1000
	movieexp = 2*60*60*1000
)

type (

	SessionPlay struct {
		Id          string		`json:"id"`
		Token 		string		`json:"token"`
		User     	string		`json:"user"`
		Mid		    string 		`json:"mid"`
		Did			string		`json:"did"`
		Plt			string		`json:"plt"`
		Active      int			`json:"active"`
		Status      int			`json:"status"`
		Ip			string		`json:"ip"`
		Process		int64		`json:"process"`
		Total		int64		`json:"total"`
		Timestamp	int64		`json:"timestamp"`
		Expired		int64		`json:"expired"`
		Rejected	string		`json:"rejected"`
		CreatedAt   time.Time	`json:"createdAt"`
		UpdatedAt   time.Time	`json:"updatedAt"`
	}

	Where struct {
		PlayId 		string
		SessionId	string
		Active		int
	}

	RequestForm struct {
		Method string
		Url string
		ApiKey string
		Body string
	}

)