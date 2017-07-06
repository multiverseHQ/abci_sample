package main

import (
	"fmt"
	"os"

	"github.com/MultiverseHQ/abci_sample"
	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

var logger tmlog.Logger

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

	app := abcicounter.NewCounterApplication(opts.Serial, logger)

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
