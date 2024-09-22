package main

import (
	"acmeproxy/dns"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libdns/libdns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

type DNSProvider struct {
	Zone     string         `yaml:"zone" validate:"required"`
	Provider string         `yaml:"provider" validate:"required"`
	Config   map[string]any `yaml:"config" validate:"required"`
}

type Provider struct {
	zone     string
	name     string
	provider dns.Provider
}

func NewProviderFromSpec(spec DNSProvider) (*Provider, error) {
	// setup env for config
	logrus.Infof("creating provider [%s]", &spec)

	cfgJson, err := json.Marshal(spec.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal config for [%s]", &spec)
	}

	dnsProvider := dns.NewProviderByNameWithConfig(spec.Provider, cfgJson)
	if dnsProvider == nil {
		return nil, fmt.Errorf("unable to obtain config for [%s]", &spec)
	}

	return &Provider{
		zone:     spec.Zone,
		name:     spec.Provider,
		provider: dnsProvider,
	}, nil
}

func (p *Provider) Present(ctx context.Context, record libdns.Record) ([]libdns.Record, error) {
	records, err := p.provider.AppendRecords(ctx, p.zone, []libdns.Record{record})
	if err != nil {
		return nil, err
	}
	return records, nil
}
func (p *Provider) CleanUp(ctx context.Context, record libdns.Record) ([]libdns.Record, error) {
	records, err := p.provider.GetRecords(ctx, p.zone)
	if err != nil {
		return nil, errors.Wrapf(err, "[%s] could not get records", p)
	}

	// some provider not returning fqdn, but a name, e.g. cloudflare
	possibleNames := []string{
		record.Name,
		// simply trim is enough, because we already verified its zone from last step
		strings.TrimSuffix(record.Name, p.zone),
	}
	var recordToDelete *libdns.Record
	for _, r := range records {
		// according to cert-manager, we should verify it's content
		// thus, it's convenient to check if content equals first
		if r.Value == record.Value {
			// check if in possible names
			for _, name := range possibleNames {
				if name == record.Name {
					recordToDelete = &r
				}
			}
		}
	}
	if recordToDelete == nil {
		return nil, fmt.Errorf("[%s] could not find record to delete", p)
	}
	records, err = p.provider.DeleteRecords(ctx, p.zone, []libdns.Record{*recordToDelete})
	if err != nil {
		return nil, errors.Wrapf(err, "[%s] could not delete record", p)
	}
	return records, nil

}
func (p *Provider) String() string {
	return fmt.Sprintf("%s/%s", p.name, p.zone)
}

func (d *DNSProvider) String() string {
	return fmt.Sprintf("%s/%s", d.Provider, d.Zone)
}
