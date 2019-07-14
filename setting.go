package main

import (
	
	"time"
	"fmt"
	"os"
	"log"
	"sync"
	// "io/ioutil"
	// "encoding/json"
	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
	"github.com/fatih/color"
	"github.com/caarlos0/env"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

)

const (
	MovieSubsPrefix = "/check/%s/%s?pt=%s&did=%s"
	UserSubsPrefix = "/subs"
	API_KEY_BILL = "xxxxx"
	EXCHANGE = "ss"
	EXCHANGETYPE = "topic"
	KEYROUTING = "ss.e.w"
)

var (
	SERVER = echo.New()
	ENV string
	WsDB string
	CmDB string
	MQ string
	CacheCnf []string
	Billing string
	MrDB string
	Debug bool
	Port  string
	local = Cache{}
	repository Repository = Repository{}
	sessionRepository SessionRepository = &repository
	otherRepository OtherRepository = &repository
	await sync.WaitGroup
	affect = sync.RWMutex{}
	handler = Handler{}
)

type (

	Excute struct {
		Env	  string	`env:"ENV" 	 envSeparator:"=" envDefault:"local"`
		Port  string    `env:"PORT"  envSeparator:"=" envDefault:"3000"`
		Debug bool 		`env:"DEBUG" envSeparator:"=" envDefault:false`
	}

	Cache struct {
		Client 			*redis.Client
		ClusterCl		*redis.ClusterClient
	}

)

func HttpStart() {

	//Installize
	Installize()

	// Debug mode
	SERVER.Debug = Debug

	// Middleware
	SERVER.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize:  1 << 10, // 1 KB
	}))
	SERVER.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		log.Println(fmt.Sprintf("%s", resBody))
	}))
	SERVER.Use(middleware.Logger())
	SERVER.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	SERVER.Use(Headers)

	// Routes
	Routes()
	// Start server
	SERVER.Logger.Fatal(SERVER.Start(":"+Port))
}

func InitConnection(Schema string) (db *gorm.DB, err error) {
	db, err = gorm.Open("mysql", Schema+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		color.Red("ERRRRRR", err.Error())
		log.Fatal(err.Error())
		time.Sleep(5 * time.Minute)
	}
	// defer db.Close()
	return
}

func Installize() {
	casbinTest()
	now := time.Now()
	cfg := Excute{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(0)
	}
	//a,b: for FPT. c for VDC
	switch cfg.Env {
		case "production":
			WsDB = "root:@localhost:8000/ss_db"
			CmDB = "root:@localhost:8000/cm_db"
			MQ = "amqp://localhost:5672"
			CacheCnf = append(CacheCnf, "localhost:6379")
			Billing = "root:@localhost:8000/bl_db"
			MrDB = "root:@localhost:8000/mr_db"
		default:
			WsDB = "root:@localhost:8000/ss_db"
			CmDB = "root:@localhost:8000/cm_db"
			MQ = "amqp://localhost:5672"
			CacheCnf = append(CacheCnf, "localhost:6379")
			Billing = "root:@localhost:8000/bl_db"
			MrDB = "root:@localhost:8000/mr_db"
	}
	Debug = cfg.Debug
	Port = cfg.Port
	ENV = cfg.Env
	local.InitCache()
	handler.SourceStreaming()
	go Worker()
	showLog := fmt.Sprintf("\n%s Your application's running on %s environment",now.Format("2019-04-02 15:04:05"),cfg.Env)
	if ENV == "production" {
		color.Magenta(showLog)
	}else { 
		color.Green(showLog)
	}
	return
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to MQ", err)
	}
	return
}

func ConnectMQ() (*amqp.Connection, *amqp.Channel, error) {
	
	connect, err := amqp.Dial(MQ)
	failOnError(err, "Failed to connect to MQ")

    channel, err := connect.Channel()
	failOnError(err, "Failed to connect to MQ")

	err = channel.ExchangeDeclare(EXCHANGE, EXCHANGETYPE,false,false,false,false,nil,) //name, type, durable, auto-deleted, internal, no-wait, arguments
	failOnError(err, "Failed to declare a queue")
	
	if err != nil {
		time.Sleep(30 * time.Second)
	}

	return connect, channel, nil
	
}

func (c *Cache) InitCache() {
	err := c.ConnectCache()
	if err != nil {
		color.Red(err.Error())
	}
	// TESTING CACHE DATABASE
	keyTest := "TEST"
	valTest := "CONTENT TEST"
	c.SetCache(keyTest, valTest, 10* time.Second)
	dataTest, err := c.GetCache(keyTest)
	if err != nil ||  dataTest == ""{
		color.Red(err.Error())
	}else {
		now := time.Now().Format("2019-04-02 15:04:05")
		color.Yellow("\n%s CHECKING CACHE DB::: %s: %s", now, keyTest, dataTest)
		color.Yellow("%s OK", now)
	}
	return
}

func (r *Cache) ConnectCache() (err error) {
	if ENV == "production"{
		r.ClusterCl = redis.NewClusterClient(&redis.ClusterOptions{Addrs:CacheCnf,})
		_, err = r.ClusterCl.Ping().Result()
	}else {
		r.Client = redis.NewClient(&redis.Options{Addr:CacheCnf[0],Password:"",DB:0, })
		_, err = r.Client.Ping().Result()
	}
	return
}

func (r *Cache) GetCache(key string) (data string, err error) {
	key = ENV + ":" + key
	if ENV == "production"{
		data, err = r.ClusterCl.Get(key).Result()
	}else { 
		data, err = r.Client.Get(key).Result()
	}
    return
}

func (r *Cache) SetCache(key string, value string, exp time.Duration) {
	key = ENV + ":" + key
	if ENV == "production"{
		r.ClusterCl.Set(key, value, exp)
	}else{
		r.Client.Set(key, value, exp)
	}
    return
}