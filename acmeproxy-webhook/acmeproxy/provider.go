package acmeproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type DNSProviderConfig struct {
	User   string `json:"user"`
	Token  string `json:"token"`
	Server string `json:"server"`
}

type request struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

type DNSClient struct {
	client *http.Client
	cfg    DNSProviderConfig
}

func NewACMEProxyDNSProvider(jsonConfig []byte) (*DNSClient, error) {
	config := DNSProviderConfig{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal config json")
	}
	return &DNSClient{
		client: &http.Client{},
		cfg:    config,
	}, nil
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
	body, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "could not marshal request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", p.cfg.Server, action), bytes.NewBuffer(body))
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could finish request: api returned %s %s", resp.Status, resp.Body)
	}

	return nil
}
