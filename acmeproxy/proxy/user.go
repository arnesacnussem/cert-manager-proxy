package proxy

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

type User struct {
	Name         string     `yaml:"name"`
	Token        string     `yaml:"token"`
	AllowedZones []*SubZone `yaml:"allowedZones"`
}

type SubZone struct {
	Zone  string `yaml:"zone"`
	Regex string `yaml:"regex"`

	regex    *regexp.Regexp
	provider *Provider
}

func (s *SubZone) Match(domain string) bool {
	if s.regex != nil {
		return s.regex.MatchString(domain)
	} else {
		return strings.HasSuffix(domain, "."+s.Zone)
	}
}
func (s *SubZone) init() error {
	if s.Zone == "" {
		return fmt.Errorf("empty zone")

	}
	if s.Regex == "" {
		return nil
	}

	if compile, err := regexp.Compile(s.Regex); err != nil {
		return errors.Wrapf(err, "in sub-zone %q: unable to compile regex %q", s.Zone, s.Regex)
	} else {
		s.regex = compile
	}
	return nil
}

func (u *User) init(providerZoneMap map[string]*Provider) error {
	var subZones []*SubZone
	for _, zone := range u.AllowedZones {
		err := zone.init()
		if err != nil {
			return errors.Wrap(err, "failed to initialize sub-rawZone")
		}

		// find provider for this sub-zone
		if zone.regex != nil {
			if provider, ok := providerZoneMap[zone.Zone]; ok {
				zone.provider = provider
			} else {
				return fmt.Errorf("unable to find provider for regex sub-rawZone %q", zone.Zone)
			}
		} else {
			for z, provider := range providerZoneMap {
				if strings.HasSuffix(zone.Zone, "."+z) {
					if zone.provider != nil {
						return fmt.Errorf("zone have multiple provider match: %q and %q", zone.provider, provider)
					} else {
						zone.provider = provider
					}
				}
			}
			if zone.provider == nil {
				return fmt.Errorf("unable to find provider for zone %q", zone.Zone)
			}
		}

		subZones = append(subZones, zone)
	}
	u.AllowedZones = subZones
	return nil
}
