package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GoSVI/semver/tools"
)

const defaultStorePath = "./report"

func init() {
	f, err := os.OpenFile("detection.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
}

func main() {
	log.Printf("----------Start")
	start := time.Now()
	parallel(getTaskChan(), 15, getWhiteList())
	fmt.Println("---------End: ", time.Since(start))
}

func parallel(c chan Item, workers int, whiteList []string) {
	wg := &sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(wg, c, whiteList)
	}
	wg.Wait()
}

func worker(wg *sync.WaitGroup, c chan Item, whiteList []string) {
	defer wg.Done()
	for len(c) != 0 {
		select {
		case item := <-c:
			err := tools.DetectSemVer(url.URL{Scheme: item.Url}, defaultStorePath, whiteList)
			if err != nil {
				log.Printf(": Fail - 0 %s - DetectSemVer Err: %s", item.Url, strings.ReplaceAll(err.Error(), "\n", ""))
			} else {
				log.Printf(": Success - 0 %s - detection sucuess ", item.Url)
			}
		default:
			return
		}
	}
}

func getTaskChan() chan Item {
	list := TaskList{}
	json.Unmarshal([]byte(readFile("/home/GoSVI/data/SemanticVersionStudy/semver/report/goRepositoryInfo.json")), &list)

	c := make(chan Item, list.Count)
	for _, item := range list.Items {
		item.Url = "https://github.com/" + item.FullName + ".git"
		c <- item
	}
	return c
}

func getWhiteList() []string {
	return trimWhiteSpace(strings.Split(readFile("white_list.txt"), "\r\n"))
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func trimWhiteSpace(list []string) []string {
	ret := make([]string, 0)
	for _, elem := range list {
		if elem != "" {
			ret = append(ret, elem)
		}
	}
	return ret
}

type TaskList struct {
	Count int    `json:"total_count"`
	Items []Item `json:"items"`
}

type Item struct {
	Index    int    `json:"index"`
	FullName string `json:"full_name"`
	Url      string `json:"url"`
	Star     int    `json:"star"`
}
