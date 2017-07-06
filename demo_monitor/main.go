package main

import (
	"log"
	"os"
)

func Execute() error {

	n, err := newNodeByAddress("http://127.0.0.1:46657")
	if err != nil {
		return err
	}

	err = n.makeTx(false)
	if err != nil {
		return err
	}

	rewards, err := n.getRewards()

	if err != nil {
		return err
	}
	log.Printf("%#v", rewards)

	return nil
}

func main() {
	if err := Execute(); err != nil {
		log.Printf("Got unhandled error: %s", err)
		os.Exit(1)
	}
}
