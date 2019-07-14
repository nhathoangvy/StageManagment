package main

import (
	"log"
	"fmt"
	"time"
	"strconv"
	"regexp"
	"strings"
	"encoding/json"
)

const (
	SESSION_STORE	= "ss_pl"
	ENTITY_SUB		= "m_subs"
	ENTITY_SOURCE	= "e_s"
)

type (
	Repository struct {
		SessionRepository
		OtherRepository
	}
	SessionRepository interface {
		Get(key string, value string) (session []SessionPlay, err error)
		GetById(playid string) (session SessionPlayResponse, err error)
		Create(init SessionPlayCreate) (result SessionPlayResponse, err error)
		Update(where Where, session SessionPlayUpdate) (success Success, err error)
	}
	OtherRepository interface {
		Sequelize(connection string, sql string) []map[string]interface{} 
	}
)

func (r *Repository) Get(key string, value string) (session []SessionPlay, err error) {
	DB, err := InitConnection(WsDB)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer DB.Close()
	query := "SELECT * FROM " + SESSION_STORE
	if key != "" && value != "" {
		query = fmt.Sprintf("%s WHERE %s = '%s'", query, key, value)
	}
	fmt.Println(query)
	DB.Raw(query).Scan(&session)
	return
}

func (r *Repository) GetById(playid string) (session SessionPlayResponse, err error) {
	DB, err := InitConnection(WsDB)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer DB.Close()
	query := "SELECT * FROM " + SESSION_STORE
	if playid != "" {
		query = fmt.Sprintf("%s WHERE id = %s LIMIT 1", query, playid)
	}
	DB.Raw(query).Scan(&session)
	return
}

func (r *Repository) Create(init SessionPlayCreate) (result SessionPlayResponse, err error) {
	DB, err := InitConnection(WsDB)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer DB.Close()

	h := Handler{}
	session := SessionPlay{}
	playId				:= h.Gen()
	timestamp			:= time.Now().Unix()
	now 				:= time.Now()

	session.Id		  	= playId
	session.AuthId		= init.SessionId
	session.UserId		= init.UserId
	session.MovieId		= init.MovieId
	session.DeviceId	= init.DeviceId
	session.Platform	= init.Platform
	session.Active		= InActive
	session.Status		= StatusInit
	session.Ip			= init.Ip
	session.Process		= ProcessInit
	session.Total		= init.Total
	session.Timestamp 	= timestamp
	session.Rejected	= init.Rejected
	session.Expired	  	= timestamp + movieexp
	session.CreatedAt	= now
	session.UpdatedAt	= now
	
	var data []SessionPlayResponse

	DB.Create(session)
	
	DB.Table(SESSION_STORE).Select("*").Where("id = ?", playId).Scan(&data)

	result = data[0]
	return
}

func (r *Repository) Update(where Where, session SessionPlayUpdate) (success Success, err error) {
	DB, err := InitConnection(WsDB)
	defer DB.Close()
	ss := SessionPlay{}
	sess := SessionPlay{}
	now := time.Now()
	if session.Active != 0{
		sess.Active = session.Active
	}
	if session.Status != 0{
		sess.Status = session.Status
	}
	if session.Process != 0{
		sess.Process = session.Process
	}
	sess.UpdatedAt = now
	var sql string
	if where.Active > 0 {
		active := strconv.Itoa(where.Active)
		if where.SessionId != ""{
			sql = fmt.Sprintf("auth_id = '%s' AND active = '%s'", where.SessionId, active)
		}
		if where.PlayId != "" {
			sql = fmt.Sprintf("id = '%s' AND active = '%s'", where.PlayId, active)
		}
		if sql != "" {
			fmt.Println(sql)
			DB.Model(&ss).Where(sql).Update(sess)
			success.Code = 0
			success.Message = "Updated Success"
		}else {
			success.Code = 1
			success.Message = "Missing params"
		}
	}else {
		if where.PlayId != "" {
			sql = fmt.Sprintf("id = '%s'", where.PlayId)
		}
		DB.Model(&ss).Where(sql).Update(sess)
		success.Code = 0
		success.Message = "Actived Success"
	}
	return
}

func (r *Repository) Sequelize(connection string, sql string) []map[string]interface{} {
	DB, err := InitConnection(connection)
	defer DB.Close()

	rows, err := DB.Raw(sql).Rows()
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}
	columns, _ := rows.Columns()
	count := len(columns)
	if count == 0{
		return nil
	}
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	result := make([]map[string]interface{}, 0, count)
    dd := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")
	for rows.Next() {
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}
        rows.Scan(valuePtrs...)
        
        t_struct := make(map[string]interface{})
        var k string
		for i, col := range columns {
			var v interface{}
            val := values[i]
			b, ok := val.([]byte)
			valEle := map[string]string{}
			var valEleArr []map[string]interface{}
			var strArr []string
			var intArr []int
			if (ok) {
				v = string(b)
				if v != nil{
					t_struct[col] = v
					err = json.Unmarshal(b, &valEle)
					if err != nil{
						err = json.Unmarshal(b, &valEleArr)
						if err != nil{
							err = json.Unmarshal(b, &strArr)
							if err != nil{
								err = json.Unmarshal(b, &intArr)
								if err == nil{
									t_struct[col] = intArr
								}
							}else { 
								t_struct[col] = strArr
							}
						}else {
							t_struct[col] = valEleArr
						}
					}else {
						t_struct[col] = valEle
					}
				}
			} else {
				v = val
				if v != nil{ 
					k = string(b)
					d := strings.Split(k, " ")
					if dd.MatchString(d[0]){
						t_struct[col] = k
					}else { 
						t_struct[col] = fmt.Sprintf("%s",v)
					}
				}
            }
		}
		result = append(result, t_struct)
	}
	 
	return result
}