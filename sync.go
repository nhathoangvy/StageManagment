package main
import (
	"fmt"
	"time"
	"sync"
	"regexp"
	"encoding/json"
)
func mainss() {
	re := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")
    fmt.Println("Time parsing");
    dateString := "2017-12-08"
	fmt.Printf("Date: %v :%v\n", dateString, re.MatchString(dateString))
	if re.MatchString(dateString){
		fmt.Println("True");
	}else {
		fmt.Println("False");
	}
	str := "[1]"
	var strs []int
	_ = json.Unmarshal([]byte(str), &strs)
	for k, v := range strs {
		fmt.Println(k,v )
	}
	fmt.Println(strs)
}

func mains() {
	var v int
	var wg sync.WaitGroup
	wg.Add(2)
	m := sync.RWMutex{}

	go func() {
		// m.Lock()
		m.Lock()
		v = 1
		fmt.Println("FIRST::",v)
		time.Sleep(time.Second*1)
		m.Unlock()
		// wg.Done()
	}()
	go func() {
		// m.RLock()
		// m.Unlock()
		m.RLock()
		// v = 2
		fmt.Println("SECOND::",v)
		time.Sleep(time.Second*2)
		m.RUnlock()
		// wg.Done()
	}()
	go func() {
		// m.RLock()
		// m.Unlock()
		m.RLock()
		// v = 2
		fmt.Println("THIRD::",v)
		time.Sleep(time.Second*2)
		m.RUnlock()
		// wg.Done()
	}()
	// m.Unlock()
	// wg.Wait()
	<-time.After(time.Second*2)
}