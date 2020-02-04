package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type counters struct {
	sync.Mutex
	View  int `json:"view"`
	Click int `json:"click"`
}

var (
	content     = []string{"sports", "entertainment", "business", "education"}
	counterList map[string]*counters
	//RateNum is number of requests per time
	rateNum = 2
	//RateTime is time limit before refresh
	rateTime = 30 * time.Second
	//Rate Queue
	rateQueue = []time.Time{}
	//Output file for storage of counter
	fileName = "output.json"
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]
	//Add a view
	alterCounterList(data, 1, 0)

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR: Issue processing request")
		w.WriteHeader(400)
		return
	}
	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}
}

//Insert some view request here
func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	alterCounterList(data, 0, 1)
	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
	//Read JSON storage file and return to user
	w.Header().Set("Content-Type", "application/json")
	var objmap map[string]*counters
	objmap = make(map[string]*counters)
	jsonFile, err := ioutil.ReadFile(fileName)
	if !handleError(err) {
		fmt.Println("ERROR: Failed reading stats file")
		return
	}
	err = json.Unmarshal(jsonFile, &objmap)
	if !handleError(err) {
		fmt.Println("ERROR: Failed parsing stats json")
		return
	}
	fmt.Fprint(w, string(jsonFile))
}
func isAllowed() bool {
	maxIndex := 0
	for _, elem := range rateQueue {
		if elem.Add(rateTime).Before(time.Now().UTC()) {
			maxIndex++
		} else {
			break
		}
	}
	rateQueue = rateQueue[maxIndex:]
	if len(rateQueue) < rateNum {
		rateQueue = append(rateQueue, time.Now().UTC())
		return true
	}
	return false
}

func uploadCounters() error {
	bytes, err := json.Marshal(counterList)
	if !handleError(err) {
		return err
	}
	//Insert Mock Store Here
	err = ioutil.WriteFile(fileName, bytes, 0644)

	if !handleError(err) {
		return err
	}
	return nil
}

func uploadLoop() {
	for {
		<-time.After(5 * time.Second)
		go uploadCounters()
	}
}

func main() {
	counterList = make(map[string]*counters)
	rateQueue = make([]time.Time, 0)
	go uploadLoop()
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getDateTime() string {
	return time.Now().UTC().Format("2006-01-02 15:04")
}
func getKeyValue(data string) string {
	return data + " : " + getDateTime()
}
func alterCounterList(data string, views int, clicks int) {
	c, exists := counterList[getKeyValue(data)]
	if !exists {
		counterList[getKeyValue(data)] = &counters{}
		c = counterList[getKeyValue(data)]
	}
	c.Lock()
	c.View += views
	c.Click += clicks
	c.Unlock()
}
func handleError(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
