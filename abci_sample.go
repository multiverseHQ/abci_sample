package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MultiverseHQ/abci_sample/counter"
	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

var logger tmlog.Logger

type options struct {
	Address  string
	ABCIType string
	Serial   bool
	Verbose  bool
}

var opts options

func ParseOptions() options {
	var opts options
	flag.StringVar(&opts.Address, "addr", "tcp://0.0.0.0:46659", "Listen address")
	flag.StringVar(&opts.ABCIType, "abci", "socket", "ABCI server: socket | grpc")
	flag.BoolVar(&opts.Serial, "serial", false, "Enforce incrementing (serial) txs")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Set verbose output")
	flag.BoolVar(&opts.Verbose, "v", false, "Set verbose output")

	flag.Parse()
	return opts
}

func init() {
	opts = ParseOptions()

	baselogger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stderr))
	if opts.Verbose == true {
		logger = tmlog.NewFilter(baselogger, tmlog.AllowAll())
		logger.Info("debug output")
	} else {
		logger = tmlog.NewFilter(baselogger, tmlog.AllowInfo())
	}
}

func Execute() error {
	fmt.Printf("\n")
	fmt.Printf("Welcome to Multiverse\n")
	fmt.Printf("\n")
	fmt.Printf("This is the ABCi Developper Application Example.\n")
	fmt.Printf("\n")
	fmt.Printf("<3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3\n")
	fmt.Printf("\n")

	app := counter.NewCounterApplication(opts.Serial, logger)

	// Start the listener
	srv, err := server.NewServer(opts.Address, opts.ABCIType, app)
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if _, err := srv.Start(); err != nil {
		return err
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

	return nil
}

func main() {
	if err := Execute(); err != nil {
		logger.Error("unhandled error", "error", err)
		os.Exit(1)
	}
}
