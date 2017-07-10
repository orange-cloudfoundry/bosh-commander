package main

import (
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	"io/ioutil"
	"regexp"
	"strings"
)

func checkOverflow(m map[string]interface{}, ctx string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown fields in %s: %s", ctx, strings.Join(keys, ", "))
	}
	return nil
}

type Config struct {
	BoshDirectors BoshDirectors `yaml:"bosh_directors"`
	LogLevel      string        `yaml:"log_level"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	names := map[string]struct{}{}
	for _, director := range c.BoshDirectors {
		if _, ok := names[director.Name]; ok {
			return fmt.Errorf("director name %q is not unique", director.Name)
		}
		names[director.Name] = struct{}{}
	}
	return nil
}

type BoshDirector struct {
	Name        string                 `yaml:"name"`
	Username    string                 `yaml:"username"`
	Password    string                 `yaml:"password"`
	DirectorUrl string                 `yaml:"director_url"`
	CACert      string                 `yaml:"ca_cert"`
	CACertFile  string                 `yaml:"ca_cert_file"`
	UaaUrl      string                 `yaml:"uaa_url"`
	Gateway     Gateway                `yaml:"gateway"`
	XXX         map[string]interface{} `yaml:",inline" json:"-"`
}
type Gateway struct {
	Username       string `yaml:"username"`
	Host           string `yaml:"host"`
	PrivateKeyPath string `yaml:"private_key_path"`
}

func (c *BoshDirector) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain BoshDirector
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.Name == "" {
		return fmt.Errorf("You must set an name to your bosh director.")
	}
	if c.DirectorUrl == "" {
		return fmt.Errorf("You must set the url to your director.")
	}
	return checkOverflow(c.XXX, "config")
}

type BoshDirectors []BoshDirector

func (boshDirs BoshDirectors) FindDirector(name string) BoshDirector {
	for _, boshDir := range boshDirs {
		if boshDir.Name == name {
			return boshDir
		}
	}
	return BoshDirector{}
}

func (boshDir *BoshDirector) LoadCaCertFile() error {
	if boshDir.CACertFile == "" {
		return nil
	}
	b, err := ioutil.ReadFile(boshDir.CACertFile)
	if err != nil {
		return fmt.Errorf("Error when loading CACert file for bosh director '%s': %s", boshDir.Name, err.Error())
	}
	if boshDir.CACert != "" {
		boshDir.CACert += "\n" + string(b)
	} else {
		boshDir.CACert = string(b)
	}
	return nil
}

type BoshCommanderScript struct {
	JobMatch    Regexp                 `yaml:"job_match"`
	Deployments []Regexp               `yaml:"deployments"`
	Sudo        bool                   `yaml:"sudo"`
	Script      []string               `yaml:"script"`
	AfterAll    []string               `yaml:"after_all"`
	XXX         map[string]interface{} `yaml:",inline" json:"-"`
}

func (c *BoshCommanderScript) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain BoshCommanderScript
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.JobMatch.String() == "" {
		return fmt.Errorf("You must set a job matcher (this can be a regex).")
	}
	if len(c.Script) == 0 {
		return fmt.Errorf("You must set an user name to connect to the supervision")
	}

	return checkOverflow(c.XXX, "config")
}

type Regexps []Regexp

func (re Regexps) MatchString(match string) bool {
	for _, regex := range re {
		if regex.MatchString(match) {
			return true
		}
	}
	return false
}

type Regexp struct {
	*regexp.Regexp
}

func (re *Regexp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	regex, err := regexp.Compile("^(?:" + s + ")$")
	if err != nil {
		return err
	}
	re.Regexp = regex
	return nil
}

type BoshSshInstance struct {
	Deployment boshdir.Deployment
	Instance   boshdir.VMInfo
}

func (i BoshSshInstance) String() string {
	indexJob := *i.Instance.Index
	return fmt.Sprintf("instance %s/%d in deployment %s", i.Instance.JobName, indexJob, i.Deployment.Name())
}
