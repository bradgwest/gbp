package main

import (
	"log"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
)

var db *mgo.Session

func dialdb() (err error) {
	log.Println("dialing mongodb: localhost")
	db, err = mgo.Dial("localhost")
	return
}

func closedb() {
	db.Close()
	log.Println("closed database connection")
}

type poll struct {
	Options []string
}

func loadOptions() ([]string, error) {
	var options []string
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	var p poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

type tweet struct {
	Text string
}

func generateTweets() <-chan string {
	out := make(chan string)
	statuses := map[string]int{
		"jimmycarter":          1,
		"roygoode":             2,
		"richardnixon":         9,
		"arnoldschwarzenegger": 25,
		"berniesanders":        4,
	}

	go func() {
		for person, n := range statuses {
			for n > 0 {
				out <- person
				n--
			}
		}
		close(out)
	}()

	return out
}

func readFromTwitter(votes chan<- string) {
	options, err := loadOptions()
	if err != nil {
		log.Println("failed to load options", err)
		return
	}

	for p := range generateTweets() {
		for _, option := range options {
			if strings.Contains(
				strings.ToLower(p),
				strings.ToLower(option),
			) {
				log.Println("vote:", option)
				votes <- option
			}
		}
	}
}

func startTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()

		for {
			select {
			case <-stopchan:
				log.Println("stopping twitter")
				return
			default:
				log.Println("querying twitter")
				readFromTwitter(votes)
				log.Println("waiting")
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return stoppedchan
}

func main() {}
