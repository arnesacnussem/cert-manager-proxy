package acmeproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type DNSProviderConfig struct {
	User           string                     `json:"user"`
	Token          string                     `json:"token"`
	Server         string                     `json:"server"`
	UserSecretRef  cmmetav1.SecretKeySelector `json:"userSecretRef"`
	TokenSecretRef cmmetav1.SecretKeySelector `json:"tokenSecretRef"`
}

type request struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

type DNSClient struct {
	client *http.Client
	cfg    *DNSProviderConfig
}

// Present a challenge to the DNS provider. This will add a TXT record that the Let's Encrypt
func (p *DNSClient) present(fqdn, value string) error {
	err := p._request("present", &request{
		fqdn, value,
	})
	if err != nil {
		return errors.Wrap(err, "could not present challenge")
	}
	return nil
}

func (p *DNSClient) cleanup(fqdn, value string) error {
	err := p._request("cleanup", &request{
		fqdn, value,
	})
	if err != nil {
		return errors.Wrap(err, "could not cleanup challenge")
	}
	return nil
}

func (p *DNSClient) _request(action string, request *request) error {
	log.Printf("[Provider]: action=%s fqdn=%q value=%q", action, request.FQDN, request.Value)
	body, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "could not marshal request")
	}

	url := fmt.Sprintf("%s/%s", p.cfg.Server, action)
	log.Printf("[Provider]: POST %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}

	req.SetBasicAuth(p.cfg.User, p.cfg.Token)
	resp, err := p.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not send request")
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	log.Printf("[Provider]: %q %s", resp.Status, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could finish request: api returned %s %s", resp.Status, resp.Body)
	}

	return nil
}
