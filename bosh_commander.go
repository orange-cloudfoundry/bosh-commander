package main

import (
	"fmt"
	"github.com/ArthurHlt/gominlog"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type BoshCommander struct {
	conf *Config
}

func (b *BoshCommander) load(c *cli.Context) (err error) {
	configPath := c.GlobalString("config")
	noColor := c.GlobalBool("no-color")
	logger.WithColor(!noColor)
	b.conf, err = b.loadConfig(configPath)
	if b.conf.LogLevel == "" {
		return
	}
	switch strings.ToUpper(b.conf.LogLevel) {
	case "ERROR":
		loggerBosh = boshlog.NewWriterLogger(boshlog.LevelError, logWriter, logWriter)
		logger.SetLevel(gominlog.Lerror)
		return
	case "WARN":
		loggerBosh = boshlog.NewWriterLogger(boshlog.LevelWarn, logWriter, logWriter)
		logger.SetLevel(gominlog.Lwarning)
		return
	case "DEBUG":
		loggerBosh = boshlog.NewWriterLogger(boshlog.LevelDebug, logWriter, logWriter)
		logger = gominlog.NewMinLog("bosh-commander", gominlog.Ldebug, true, log.Ldate|log.Ltime)
		logger.SetWriter(logWriter)
		return
	}
	return
}

func (b BoshCommander) runCommandRunner(c *cli.Context) error {
	err := b.load(c)
	if err != nil {
		return err
	}
	directorName := c.String("director")
	var directors []BoshDirector
	if directorName != "" {
		director := b.conf.BoshDirectors.FindDirector(directorName)
		if director.Name == "" {
			return fmt.Errorf("Director %s cannot be found in config", directorName)
		}
		directors = append(directors, director)
	} else {
		directors = []BoshDirector(b.conf.BoshDirectors)
	}
	scriptPath := c.String("file-script")
	if scriptPath == "" {
		return fmt.Errorf("You must a path to your script to run")
	}
	boshCommanderScript, err := b.loadBoshCommanderScript(scriptPath)
	if err != nil {
		return err
	}
	commandRunner := NewCommandRunner(boshCommanderScript)
	for _, director := range directors {
		logger.Info("Running bosh commander on director '%s' ...\n", director.Name)
		err = director.LoadCaCertFile()
		if err != nil {
			return err
		}
		err := commandRunner.Run(director)
		if err != nil {
			return err
		}
		logger.Info("Finished running commander on director '%s'.\n\n", director.Name)
	}
	return nil
}

func (p BoshCommander) loadConfig(path string) (*Config, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(dat, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
func (p BoshCommander) loadBoshCommanderScript(path string) (*BoshCommanderScript, error) {
	var b []byte
	var err error
	if path == "-" {
		b, err = ioutil.ReadAll(os.Stdin)

	} else {
		b, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	boshCommanderScript := &BoshCommanderScript{}
	err = yaml.Unmarshal(b, &boshCommanderScript)
	if err != nil {
		return nil, err
	}
	return boshCommanderScript, nil
}
