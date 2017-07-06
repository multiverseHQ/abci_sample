package main

import (
	"log"
	"os"
)

func Execute() error {

	nm, err := newNodeManager("http://127.0.0.1:46657")
	if err != nil {
		return err
	}

	log.Printf("%s", nm)
	return nil
}

func main() {
	if err := Execute(); err != nil {
		log.Printf("Got unhandled error: %s", err)
		os.Exit(1)
	}
}
