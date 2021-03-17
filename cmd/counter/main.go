package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const updateDuration = 1 * time.Second

var fatalErr error
var counts map[string]int
var countsLock sync.Mutex

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func doCount(countsLock *sync.Mutex, counts *map[string]int, pollData *mgo.Collection) {
	countsLock.Lock()
	defer countsLock.Unlock()

	if len(*counts) == 0 {
		log.Println("no new votes, skipping database update")
		return
	}

	log.Println("updating database...")
	log.Println(*counts)

	ok := true
	for option, count := range *counts {
		sel := bson.M{"options": bson.M{"$in": []string{option}}}
		up := bson.M{"$inc": bson.M{"results." + option: count}}

		if _, err := pollData.UpdateAll(sel, up); err != nil {
			log.Println("failed to update:", err)
			ok = false
		}
	}

	if ok {
		log.Println("finished updating database...")
		*counts = nil
	}
}

func main() {
	var (
		mongoAddr  = flag.String("mongoAddr", "localhost", "mongodb address")
		lookupAddr = flag.String("lookupAddr", "localhost:4161", "Address and port for nsq lookup")
	)
	flag.Parse()

	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	log.Println("connecting to database...")
	db, err := mgo.Dial(*mongoAddr)
	if err != nil {
		fatal(err)
		return
	}

	defer func() {
		log.Println("closing database connection...")
		db.Close()
	}()

	pollData := db.DB("ballots").C("polls")

	log.Println("connecting to nsq...")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}
	log.Println("connected to nsq")

	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))

	if err := q.ConnectToNSQLookupd(*lookupAddr); err != nil {
		fatal(err)
		return
	}

	ticker := time.NewTicker(updateDuration)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		select {
		case <-ticker.C:
			doCount(&countsLock, &counts, pollData)
		case <-termChan:
			ticker.Stop()
			q.Stop()
		case <-q.StopChan:
			// finished
			return
		}
	}

}
