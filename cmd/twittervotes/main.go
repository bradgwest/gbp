package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
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

func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())

	go func() {
		for vote := range votes {
			pub.Publish("votes", []byte(vote))
		}

		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
		stopchan <- struct{}{}
	}()

	return stopchan
}

func main() {
	var stoplock sync.Mutex
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)

	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		// closeConn()  // commented out because we don't actually open a twitter connection
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if err := dialdb(); err != nil {
		log.Fatalln("failed to dial MongoDB:", err)
	}
	defer closedb()

	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			// closeConn()  // commented out because we don't actually open a twitter connection
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()

	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}
