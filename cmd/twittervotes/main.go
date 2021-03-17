package main

import (
	"flag"
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

func dialdb(host string) (err error) {
	log.Println("dialing mongodb: ", host)
	db, err = mgo.Dial(host)
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

func publishVotes(votes <-chan string, nsqAddr string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer(nsqAddr, nsq.NewConfig())

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
	var (
		loadOnce sync.Once
		stoplock sync.Mutex
		mongo    = flag.String("mongoAddr", "localhost", "mongodb address")
		nsqAddr  = flag.String("nsqAddr", "localhost:4150", "nsq address")
	)
	flag.Parse()

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

	if err := dialdb(*mongo); err != nil {
		log.Fatalln("failed to dial MongoDB:", err)
	}
	defer closedb()

	// TODO would need to parse this from cmd line, or hard code
	// loadOnce.Do(func() {
	// 	if err := db.DB("ballots").C("polls").Insert(); err != nil {
	// 		log.Fatalln("failed to load mongodb:", err)
	// 	}
	// })

	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes, *nsqAddr)
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
