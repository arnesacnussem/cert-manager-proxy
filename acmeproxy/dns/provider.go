package dns

// see https://github.com/orgs/libdns/repositories

import (
	"encoding/json"
	"github.com/libdns/libdns"
)

type Provider interface {
	libdns.RecordGetter
	libdns.RecordDeleter
	libdns.RecordAppender
}

func NewProviderByNameWithConfig(name string, cfgJson []byte) Provider {
	p := NewProviderByName(name)
	err := json.Unmarshal(cfgJson, p)
	if err != nil {
		return nil
	}
	return p
}
