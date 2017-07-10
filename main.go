package main

import (
	"github.com/ArthurHlt/gominlog"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/urfave/cli"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var logger *gominlog.MinLog
var loggerBosh boshlog.Logger
var logWriter io.Writer

func init() {
	logWriter = os.Stderr
	logger = gominlog.NewMinLog("bosh-commander", gominlog.Linfo, true, log.Ldate|log.Ltime)
	logger.SetWriter(logWriter)
	loggerBosh = boshlog.NewWriterLogger(boshlog.LevelInfo, logWriter, logWriter)
}

func main() {
	boshCommander := &BoshCommander{}
	app := cli.NewApp()
	usr, err := user.Current()
	if err != nil {
		logger.Severe(err.Error())
		return
	}
	app.Name = "bosh-commander"
	app.Version = "1.0.0"
	app.Usage = "Run commands on targetted job name on multiple bosh"
	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Run commands on targetted job name",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "director, d",
					Usage: "Set which bosh director should be use from config (this is the name of the director)",
				},
				cli.StringFlag{
					Name:  "file-script, f",
					Value: "bosh_script.yml",
					Usage: "Set where the output should be write set to - to see in stdout and load through stdin",
				},
			},
			Action: boshCommander.runCommandRunner,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: filepath.Join(usr.HomeDir, ".bosh_commander.yml"),
			Usage: "Path to the config file",
		},
		cli.BoolFlag{
			Name:  "no-color",
			Usage: "Logger will not display colors",
		},
	}

	app.Run(os.Args)
}
