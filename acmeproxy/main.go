package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libdns/libdns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return f.Function, ""
		},
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
}

type Config struct {
	Server    string        `yaml:"server"`
	Users     []User        `yaml:"users"`
	Providers []DNSProvider `yaml:"providers"`
}

type User struct {
	Name         string   `yaml:"name"`
	Token        string   `yaml:"token"`
	AllowedZones []string `yaml:"allowedZones"`
}

type Request struct {
	FQDN  string `json:"fqdn" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type Server struct {
	config    *Config
	users     map[string]*User
	providers map[string]*Provider
}

type action struct {
	provider *Provider
	request  *libdns.Record
}

func (s *Server) Present(ctx *gin.Context) {
	act, err := s.common(ctx)
	if err != nil {
		return
	}

	records, err := act.provider.Present(ctx, *act.request)
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": fmt.Sprintf("error appending record %s: %s", act.request.Name, err),
			"success": false,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"records": records,
		"success": true,
	})
}

func (s *Server) CleanUp(ctx *gin.Context) {
	act, err := s.common(ctx)
	if err != nil {
		return
	}
	records, err := act.provider.CleanUp(ctx, *act.request)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"message": fmt.Sprintf("error cleaning up record %s: %s", act.request.Name, err),
			"success": false,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"records": records,
		"success": true,
	})
}

func (s *Server) common(ctx *gin.Context) (*action, error) {
	user := ctx.MustGet(gin.AuthUserKey).(string)
	var request Request
	err := ctx.BindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"message": "bad request, unable to bind json",
		})
		return nil, err
	}

	// check allowed zones
	// cert-manager will add a . to the end
	request.FQDN = strings.TrimSuffix(request.FQDN, ".")
	subZone, ok := isInAllowedZones(request.FQDN, s.users[user].AllowedZones)
	if !ok {
		ctx.AbortWithStatusJSON(403, gin.H{
			"message": "domain not allowed",
			"success": false,
		})
		return nil, fmt.Errorf("domain not allowed")
	}

	return &action{
		provider: s.providers[subZone],
		request: &libdns.Record{
			Type:  "TXT",
			Name:  request.FQDN,
			Value: request.Value,
		},
	}, nil
}

func (s *Server) Serve() {

	accounts := make(map[string]string)
	for _, user := range s.config.Users {
		accounts[user.Name] = user.Token
	}
	router := gin.Default()
	router.POST("/present", gin.BasicAuth(accounts), s.Present)
	router.POST("/cleanup", gin.BasicAuth(accounts), s.CleanUp)

	err := router.Run(s.config.Server)
	if err != nil {
		logrus.Errorf("Failed to start server at: %s", s.config.Server)
		panic(err)
	}
}

func loadConfig() (*Config, string) {
	// load config from file at CONFIG_PATH or ./config.yaml
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config.yaml"
	}
	path, err := filepath.Abs(path)
	if err != nil {
		panic(errors.Wrapf(err, "unable to locate configfile"))
	}

	logrus.Infof("using config file: %s", path)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(errors.Wrapf(err, "unable to open config file [%s]", path))
	}
	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		panic(errors.Wrap(err, "error unmarshalling config"))
	}
	return config, path
}

func NewServer() *Server {
	config, path := loadConfig()
	logrus.Infof("found %d users", len(config.Users))
	logrus.Infof("found %d providers", len(config.Providers))

	errs := make([]error, 0)
	userMap := make(map[string]*User)
	providerMap := make(map[string]*Provider)
	userSubZoneMap := make(map[string]map[string]string)
	providerConfigMap := make(map[string]DNSProvider)
	subZoneProviderMap := make(map[string]*Provider)

	var providersAvailable []string
	for _, provider := range config.Providers {
		if strings.HasSuffix(provider.Zone, ".") {
			logrus.Warnf("zone [%s] should not end with a \".\"", provider.Zone)
			provider.Zone = strings.TrimSuffix(provider.Zone, ".")
		}
		providerConfigMap[provider.Zone] = provider
		providersAvailable = append(providersAvailable, provider.Zone)
	}

	for _, user := range config.Users {
		var subZones []string
		providerNameMap := make(map[string]string)
		for _, sub := range user.AllowedZones {
			if strings.HasSuffix(sub, ".") {
				logrus.Warnf("sub-sub [%s] of user [%s] should not end with a \".\"", sub, user.Name)
				sub = strings.TrimSuffix(sub, ".")
			}

			providerName, found := isInAllowedZones(sub, providersAvailable)
			if !found {
				errs = append(errs, fmt.Errorf("provider for [%s] is missing, but user [%s] wants one", sub, user.Name))
				continue
			}

			providerNameMap[sub] = providerName
			subZones = append(subZones, sub)
		}
		user.AllowedZones = subZones
		userMap[user.Name] = &user

		if len(providerNameMap) == 0 {
			errs = append(errs, fmt.Errorf("no provider found for user [%s]", user.Name))
			continue
		}

		userSubZoneMap[user.Name] = providerNameMap
		for zoneName, providerName := range providerNameMap {
			spec := providerConfigMap[providerName]
			provider, err := findOrCreateProvider(&providerMap, spec)
			if err != nil {
				errs = append(errs, err)
			}

			subZoneProviderMap[zoneName] = provider
		}
	}

	if len(errs) > 0 {
		logrus.Errorf("failed to validate config, errs: %d", len(errs))
		for _, s := range errs {
			logrus.Errorf("  - %s", s)
		}
		logrus.Fatalf("config at [%s] is invalid, please fix it and retry.", path)
	}

	return &Server{
		config:    config,
		users:     userMap,
		providers: subZoneProviderMap,
	}
}

func findOrCreateProvider(providerMap *map[string]*Provider, spec DNSProvider) (*Provider, error) {
	name := spec.Zone
	if provider, ok := (*providerMap)[name]; ok {
		return provider, nil
	}
	provider, err := NewProviderFromSpec(spec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dns provider [%s]", &spec)
	}
	(*providerMap)[name] = provider
	return provider, nil
}

func isInAllowedZones(domain string, allowedZones []string) (string, bool) {
	for _, zone := range allowedZones {
		if strings.HasSuffix(domain, "."+zone) || domain == zone {
			return zone, true
		}
	}
	return "", false
}

func main() {
	NewServer().Serve()
}
