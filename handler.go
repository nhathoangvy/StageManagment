package main
import (
    "bytes"
    "math/rand"
    "io/ioutil"
    "fmt"
    "net/http"
    "encoding/json"
    "log"
    "github.com/fatih/color"
    "time"
    "math"
    "github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
)

const fixed = "abcdefghijklmnopqrstuvwxyz0123456789"

type Handler struct {}
type Member string
type hasher struct{}

func (h *Handler) Contains(a []string, x string) bool {
    for _, n := range a {
        if x == n {
            return true
        }
    }
    return false
}

func (h *Handler) Gen() string {
    b := make([]byte, 15)
    for i := range b {
        b[i] = fixed[rand.Intn(len(fixed))]
    }
    return string(b)
}

func (h *Handler) CheckSubscriptions(uid string, mid string, plt string, did string) (result map[string]interface{}, erro string) {
    uri := fmt.Sprintf(Billing+MovieSubsPrefix,mid, uid, plt, did)
    data, err := handler.Request(RequestForm{
        Method: "GET",
        Url: uri,
        ApiKey: API_KEY_BILL,
    })
    err = json.Unmarshal([]byte(data), &result)
    if result["code"] != nil || err != nil{
        erro = "Failed"
    }
    return
}

func (h *Handler) UserPermissions(uid string, mid string, did string) (result map[string]interface{}, erro string) {
    uri := fmt.Sprintf(Billing+UserSubsPrefix)
    dataReq := &UserSubs{
        UserId: uid,
        DeviceId: did,
        ItemId: mid,
    }
    dataReqs, err := json.Marshal(dataReq)
    data, err := handler.Request(RequestForm{
        Method: "POST",
        Url: uri,
        ApiKey: API_KEY_BILL,
        Body: string(dataReqs),
    })
    err = json.Unmarshal([]byte(data), &result)
    if err != nil{
        log.Fatal(err.Error())
        erro = "failed"
    }
    return
}

func (h *Handler) Request(data RequestForm) (result []byte, err error) {
    client := &http.Client{}
    if data.Method == "GET"{
        req, err := http.NewRequest(data.Method, data.Url, nil)
        if err !=nil {
            log.Fatal(err.Error())
        }
        req.Header.Add("api-key", data.ApiKey)
        resp, err := client.Do(req)
        result, err = ioutil.ReadAll(resp.Body)
        resp.Body.Close()
    }else { 
        req, err := http.NewRequest(data.Method, data.Url, bytes.NewBufferString(data.Body))
        if err !=nil {
            log.Fatal(err.Error())
        }
        req.Header.Add("api-key", data.ApiKey)
        resp, err := client.Do(req)
        result, err = ioutil.ReadAll(resp.Body)
        resp.Body.Close()
    }
    return
}

func (m Member) String() string {
	return string(m)
}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func (h *Handler) SourceStreaming() {
	cron, err := local.GetCache("CRON-UPDATE-STREAMING")
	if err != nil{
		fmt.Println(err.Error())
    }
    now := time.Now().Format("2006-01-02 15:04:05")
    color.Yellow(fmt.Sprintf("%s STREAMING CACHE RESOURCE %s", now, cron))
	if cron == ""{
        color.Green(fmt.Sprintf("%s LOADING STREAMING....", now))
		members := []consistent.Member{}
		for i := 1; i < 4; i++ {
			member := Member(fmt.Sprintf("http://localhost:8080/cdn", i))
			members = append(members, member)
		}
		cfg := consistent.Config{
			PartitionCount:    1000,
			ReplicationFactor: 400,
			Load:              1.3,
			Hasher:            hasher{},
		}
		c := consistent.New(members, cfg)
		entity := otherRepository.Sequelize(CmDB,"SELECT * FROM movieLists WHERE status = 1")
		keyCount := len(entity)
		load := (c.AverageLoad() * float64(keyCount)) / float64(cfg.PartitionCount)
		fmt.Println("Maximum key count for a member should be around this: ", math.Ceil(load))
		distribution := make(map[string]int)
		for i := 0; i < keyCount; i++ {
			// rand.Read(key)
			key := []byte(entity[i]["id"].(string))
			member := c.LocateKey(key)
			local.SetCache(fmt.Sprintf("Streaming:%s",string(key)), fmt.Sprintf("%s",member), 24* time.Hour)
			fmt.Println("SET CACHE", fmt.Sprintf("Streaming:%s",string(key)), member)
			distribution[member.String()]++
		}
			local.SetCache("CRON-UPDATE-STREAMING","ALIVE",24* time.Hour)
			fmt.Println("SET CACHE CRON-UPDATE-STREAMING")
	}
	return
}