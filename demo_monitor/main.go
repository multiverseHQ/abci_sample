package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	ui "github.com/gizak/termui"
	"github.com/jessevdk/go-flags"
)

type App struct {
	nm      *nodeManager
	logfile io.WriteCloser

	echo  *EchoArea
	table *ui.Table
	mx    sync.Mutex
}

type Options struct {
	Verbose bool   `short:"v" long:"verbose" description:"Make verbose output"`
	Address string `short:"a" long:"address" description:"Address of the tendermint first node" default:"localhost"`
	Port    int    `short:"p" long:"port" description:"Port of the tendermint RPC" default:"46657"`
}

func (app *App) updateTable(infos []*nodeInfo) {
	app.table.BorderLabel = " Network Status | " + time.Now().Format("Mon Jan 2 15:04:05") + " "
	changed := false
	if len(infos)+1 != len(app.table.Rows) {
		changed = true
	}
	app.table.Rows = app.table.Rows[0:1]
	for _, i := range infos {
		app.table.Rows = append(app.table.Rows,
			[]string{
				i.Key[0:6],
				fmt.Sprintf("%04d", i.Reward),
				fmt.Sprintf("%04d", i.VotingPower),
				fmt.Sprintf("%04d", i.TxCount),
			})
	}
	if changed == true {
		log.Printf("Got new nodes!!!")
		app.table.Analysis()
		app.table.SetSize()
		ui.Body.Align()
		ui.Render(ui.Body)
		return
	}
	ui.Render(app.table)
}

func (app *App) SetUpUi() {
	log.Printf("Creating ui")

	info := ui.NewPar(`Press Q to quit
Press T to send a Tx to a random node`)
	info.Height = 4
	info.TextFgColor = ui.ColorWhite
	info.BorderLabel = "Info"
	app.echo = NewEchoArea(10)
	app.echo.Label = " Logs "

	app.table = ui.NewTable()
	app.table.Rows = [][]string{
		[]string{"Key", "Reward", "Power", "Tx"},
	}
	infos, err := app.nm.fetchStatus()
	if err == nil {
		app.updateTable(infos)
	}
	app.table.Height = len(app.nm.Keys) + 2
	app.table.BgColor = ui.ColorDefault
	app.table.FgColor = ui.ColorDefault

	app.table.Separator = true
	app.table.Border = true
	app.table.Analysis()
	app.table.FgColors[0] = ui.ColorGreen
	app.table.SetSize()
	app.table.BorderLabel = " Network Status "

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, info),
		),
		ui.NewRow(
			ui.NewCol(12, 0, app.table),
		),
		ui.NewRow(
			ui.NewCol(12, 0, app.echo),
		),
	)

	ui.Body.Align()

	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		log.Printf("Exiting")
		ui.StopLoop()
	})

	ui.Handle("/sys/kbd/t", func(ui.Event) {
		go func() {
			app.mx.Lock()
			defer app.mx.Unlock()
			log.Printf("Commiting transaction...")
			//ugly way to do it
			idx := app.nm.curNodeIdx + 1
			for i, _ := range app.table.BgColors {
				if i == 0 {
					continue
				}
				app.table.FgColors[i] = ui.ColorDefault
			}
			app.table.FgColors[idx] = ui.ColorCyan
			ui.Render(app.table)
			err := app.nm.commitTx(false)
			if err != nil {
				log.Printf("could not send transaction: %s", err)
			}

		}()
	})

	ui.DefaultEvtStream.Merge("/timer/500ms", ui.NewTimerCh(500*time.Millisecond))
	ui.Handle("/timer/500ms", func(e ui.Event) {
		go func() {
			app.mx.Lock()
			defer app.mx.Unlock()
			infos, err := app.nm.fetchStatus()
			if err != nil {
				log.Printf("Cannot fetch status: %s", err)
				return
			}
			app.updateTable(infos)
		}()
	})

}

func (app *App) SetUpLog() error {
	filename := fmt.Sprintf("%s/multiverse_demo.%d.log", os.TempDir(), os.Getpid())
	var err error
	app.logfile, err = os.Create(filename)
	if err != nil {
		return err
	}
	log.SetOutput(app.logfile)
	return nil
}

func Execute() error {

	var opts Options
	if _, err := flags.Parse(&opts); err != nil {
		if ferr, ok := err.(*flags.Error); ok == true && ferr.Type == flags.ErrHelp {
			return nil
		}
		return err
	}

	var app App

	err := app.SetUpLog()
	if err != nil {
		return err
	}
	defer func() {
		log.SetOutput(os.Stderr)
		app.logfile.Close()
	}()

	app.nm, err = newNodeManager(fmt.Sprintf("http://%s:%d", opts.Address, opts.Port))
	if err != nil {
		return err
	}

	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	app.SetUpUi()
	if opts.Verbose == true {
		log.SetOutput(io.MultiWriter(app.logfile, app.echo))
	} else {
		log.SetOutput(app.logfile)
	}

	ui.Loop()

	return nil
}

func main() {
	if err := Execute(); err != nil {
		log.Printf("Got unhandled error: %s", err)
		os.Exit(1)
	}
}
