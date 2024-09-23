package proxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libdns/libdns"
	"github.com/sirupsen/logrus"
	"strings"
)

type Server struct {
	users  map[string]*User
	config *Config
}

type Request struct {
	FQDN  string `json:"fqdn" binding:"required"`
	Value string `json:"value" binding:"required"`
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
	// cert-manager may add a . to the end
	request.FQDN = strings.TrimSuffix(request.FQDN, ".")
	zone := s.findTargetZone(user, &request)
	if zone == nil {
		ctx.AbortWithStatusJSON(403, gin.H{
			"message": "domain not allowed",
			"success": false,
		})
		return nil, fmt.Errorf("domain not allowed")
	}

	return &action{
		provider: zone.provider,
		request: &libdns.Record{
			Type:  "TXT",
			Name:  request.FQDN,
			Value: request.Value,
		},
	}, nil
}

// findTargetZone finds the target zone for the given user and request.
//
// It iterates over the allowed zones for the given user and checks if the
// request's FQDN matches any of them.
// The first matched zone is returned.
func (s *Server) findTargetZone(username string, request *Request) *SubZone {
	user := s.users[username]
	for _, zone := range user.AllowedZones {
		if zone.Match(request.FQDN) {
			return zone
		}
	}

	return nil
}

func NewServer() *Server {
	config := loadConfig()
	return config.CreateServer()
}

func (s *Server) Serve() {
	accounts := make(map[string]string)
	for _, user := range s.users {
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
