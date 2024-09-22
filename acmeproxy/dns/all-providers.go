package dns

import (
	"github.com/libdns/acmeproxy"
	"github.com/libdns/alidns"
	"github.com/libdns/azure"
	"github.com/libdns/bunny"
	"github.com/libdns/civo"
	"github.com/libdns/cloudflare"
	"github.com/libdns/ddnss"
	"github.com/libdns/desec"
	"github.com/libdns/digitalocean"
	"github.com/libdns/dinahosting"
	"github.com/libdns/directadmin"
	"github.com/libdns/dnsimple"
	"github.com/libdns/dnspod"
	"github.com/libdns/dnsupdate"
	"github.com/libdns/duckdns"
	"github.com/libdns/dynu"
	"github.com/libdns/dynv6"
	"github.com/libdns/easydns"
	"github.com/libdns/gandi"
	"github.com/libdns/glesys"
	"github.com/libdns/godaddy"
	"github.com/libdns/googleclouddns"
	"github.com/libdns/he"
	"github.com/libdns/hetzner"
	"github.com/libdns/hexonet"
	"github.com/libdns/hosttech"
	"github.com/libdns/infomaniak"
)

// NewProviderByName see https://github.com/orgs/libdns/repositories
func NewProviderByName(name string) Provider {
	switch name {
	case "acmeproxy":
		return &acmeproxy.Provider{}

	case "alidns":
		return &alidns.Provider{}

	case "azure":
		return &azure.Provider{}

	case "bunny":
		return &bunny.Provider{}

	case "civo":
		return &civo.Provider{}

	case "cloudflare":
		return &cloudflare.Provider{}

	case "ddnss":
		return &ddnss.Provider{}

	case "desec":
		return &desec.Provider{}

	case "digitalocean":
		return &digitalocean.Provider{}

	case "dinahosting":
		return &dinahosting.Provider{}

	case "directadmin":
		return &directadmin.Provider{}

	case "dnsimple":
		return &dnsimple.Provider{}

	case "dnspod":
		return &dnspod.Provider{}

	case "dnsupdate":
		return &dnsupdate.Provider{}

	case "duckdns":
		return &duckdns.Provider{}

	case "dynu":
		return &dynu.Provider{}

	case "dynv6":
		return &dynv6.Provider{}

	case "easydns":
		return &easydns.Provider{}

	case "gandi":
		return &gandi.Provider{}

	case "glesys":
		return &glesys.Provider{}

	case "godaddy":
		return &godaddy.Provider{}

	case "googleclouddns":
		return &googleclouddns.Provider{}

	case "he":
		return &he.Provider{}

	case "hetzner":
		return &hetzner.Provider{}

	case "hexonet":
		return &hexonet.Provider{}

	case "hosttech":
		return &hosttech.Provider{}

	case "infomaniak":
		return &infomaniak.Provider{}

	}
	return nil
}
