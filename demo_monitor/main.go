package main

import (
	"log"
	"os"
	"time"
)

func Execute() error {

	nm, err := newNodeManager("http://127.0.0.1:46657")
	if err != nil {
		return err
	}

	for i := 0; i < 40; i++ {
		if i%2 == 0 {
			nm.commitTx(false)
		}
		infos, err := nm.fetchStatus()
		if err != nil {
			return err
		}
		for i, ni := range infos {
			log.Printf("%03d: %s", i, ni)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func main() {
	if err := Execute(); err != nil {
		log.Printf("Got unhandled error: %s", err)
		os.Exit(1)
	}
}
