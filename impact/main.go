package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/GoSVI/impact-go/analyze"
	"github.com/GoSVI/impact-go/config"
)

const defaultStorePath = "./report"

func init() {
	f, err := os.OpenFile("impact.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
}

func main() {
	log.Println("-----Start")
	parallel(getTaskChan(), 15)
	log.Println("-----End")
}
func parallel(c chan config.Item, workers int) {
	wg := &sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(wg, c)
	}
	wg.Wait()
}

func worker(wg *sync.WaitGroup, c chan config.Item) {
	defer wg.Done()
	for len(c) != 0 {
		select {
		case item := <-c:
			err := analyze.DetectImpact(item, defaultStorePath)
			if err != nil {
				log.Printf(": Fail - 0 %s - Detection Error: %s", item.Url, strings.ReplaceAll(err.Error(), "\n", ""))
			} else {
				log.Printf(": Success - 0 %s - Detection Sucuess ", item.Url)
			}
		default:
			return
		}
	}
}

func getTaskChan() chan config.Item {
	list := config.TaskList{}
	json.Unmarshal([]byte(readFile("/home/GoSVI/data/impact/report/client.json")), &list)

	c := make(chan config.Item, list.Count)
	for _, item := range list.Items {
		c <- item
	}
	return c
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}
