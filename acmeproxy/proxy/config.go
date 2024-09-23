package proxy

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"slices"
)

type Config struct {
	Server    string         `yaml:"server"`
	Users     []*User        `yaml:"users"`
	Providers []*DNSProvider `yaml:"providers"`

	userMap         map[string]*User
	providerZoneMap map[string]*Provider
}

func (c *Config) CreateServer() *Server {
	savedErrors := c.loadAllProvider()
	if checkSavedErrors(savedErrors) {
		panic("error creating providers")
	}

	savedErrors = c.loadAllUser()
	if checkSavedErrors(savedErrors) {
		panic("error creating users")
	}

	return &Server{
		users:  c.userMap,
		config: c,
	}
}

func loadConfig() *Config {
	// load config from file at CONFIG_PATH or ./config.yaml
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config.yaml"
	}
	path, err := filepath.Abs(path)
	if err != nil {
		panic(errors.Wrapf(err, "unable to locate configfile"))
	}

	logrus.Infof("using config file at %q", path)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(errors.Wrapf(err, "unable to open config file %q", path))
	}
	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		panic(errors.Wrap(err, "error unmarshalling config"))
	}

	logrus.Infof("found %d users", len(config.Users))
	logrus.Infof("found %d providers", len(config.Providers))
	return config
}

func (c *Config) loadAllProvider() (savedErrors []error) {
	c.providerZoneMap = make(map[string]*Provider)
	for _, spec := range c.Providers {
		provider, err := spec.ToProvider()
		if err != nil {
			savedErrors = append(savedErrors, errors.Wrapf(err, "error creating provider %q", spec))
			continue
		}
		c.providerZoneMap[provider.zone] = provider
	}
	return savedErrors
}

func (c *Config) loadAllUser() (savedErrors []error) {
	c.userMap = make(map[string]*User)
	for _, user := range c.Users {
		err := user.init(c.providerZoneMap)
		if err != nil {
			savedErrors = append(savedErrors, errors.Wrapf(err, "error creating user %q", user.Name))
			continue
		}
		c.userMap[user.Name] = user
	}
	return savedErrors
}

func (c *Config) checkUnusedProvider() {
	var providerInUse []string
	for _, user := range c.userMap {
		for _, zone := range user.AllowedZones {
			providerInUse = append(providerInUse, zone.provider.String())
		}
	}

	var providerNotInUse []string
	providerInUse = removeDuplicateStr(providerInUse)
	for _, provider := range c.providerZoneMap {
		find := slices.Index(providerInUse, provider.String())
		if find == -1 {
			providerNotInUse = append(providerNotInUse, provider.String())
		}
	}

	if len(providerNotInUse) > 0 {
		logrus.Warnf("Found following provider is not in use: %s", providerInUse)
	}
}

func checkSavedErrors(e []error) bool {
	if e != nil && len(e) > 0 {
		for _, err := range e {
			logrus.Errorf("\t- %s", err)
		}
		return true
	}
	return false
}

// removeDuplicateStr copy from https://stackoverflow.com/a/66751055
func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
