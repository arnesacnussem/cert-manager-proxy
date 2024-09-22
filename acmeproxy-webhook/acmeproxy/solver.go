package acmeproxy

import (
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
)

type Solver struct {
	provider *DNSClient
}

func (c *Solver) Name() string {
	return "acmeproxy"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *Solver) Present(ch *v1alpha1.ChallengeRequest) error {
	client, err := newClient(ch.Config)
	if err != nil {
		return err
	}
	err = client.present(ch.ResolvedFQDN, ch.Key)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *Solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	client, err := newClient(ch.Config)
	if err != nil {
		return err
	}
	err = client.cleanup(ch.ResolvedFQDN, ch.Key)
	if err != nil {
		return err
	}

	return nil
}

func (c *Solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	///// UNCOMMENT THE BELOW CODE TO MAKE A KUBERNETES CLIENTSET AVAILABLE TO
	///// YOUR CUSTOM DNS PROVIDER

	//cl, err := kubernetes.NewForConfig(kubeClientConfig)
	//if err != nil {
	//	return err
	//}
	//
	//c.client = cl

	///// END OF CODE TO MAKE KUBERNETES CLIENTSET AVAILABLE
	return nil
}

func newClient(cfgJSON *extapi.JSON) (*DNSClient, error) {
	client, err := NewACMEProxyDNSProvider(cfgJSON.Raw)
	if err != nil {
		return nil, err
	}

	return client, nil
}
