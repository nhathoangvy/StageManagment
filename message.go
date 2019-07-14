package main

import (
	"log"
	"github.com/streadway/amqp"
	"time"
	"encoding/json"
)

func SendMessage(data string) {
	connect, channel, err := ConnectMQ()
	defer connect.Close()
	defer channel.Close()

	err = channel.Publish(EXCHANGE, KEYROUTING, false, false, 
		amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte(data),
		})
	failOnError(err, "Failed to publish a message")

	log.Println("SENT", data)

	return
}

func Receiver(msgs <-chan amqp.Delivery) {
	var data map[string]interface{}
	for d := range msgs {
		DB, _ := InitConnection(SsDB)
		defer DB.Close()
		if err := json.Unmarshal([]byte(d.Body), &data); err != nil {
			log.Fatal(err.Error())
		}
		ss := SessionPlay{}
		now := time.Now()

		if data["pid"] != nil{
			pid := data["pid"].(string)
			pcs	:= int64(data["pcs"].(float64))
			DB.Model(&ss).Where("id = ?", pid).Update(SessionPlay{Process: pcs, Status: StatusPlaying, UpdatedAt: now})
			log.Printf("Received watching: %s", pid, pcs)
		}
		if data["id"] != nil { 
			jti := data["id"].(string)
			DB.Model(&ss).Where("rejected = ?", jti).Update(SessionPlay{Status: StatusKicked, UpdatedAt: now})
			log.Printf("Received revoked: %s", jti)
		}
	}
}