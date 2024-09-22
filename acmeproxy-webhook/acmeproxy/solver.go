package acmeproxy

import (
	"context"
	"encoding/json"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
)

type Solver struct {
	kubeClient *kubernetes.Clientset
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
	client, err := c.getClient(ch)
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
	client, err := c.getClient(ch)
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
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.kubeClient = cl
	return nil
}

func (c *Solver) getSecretVal(selector cmmetav1.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := c.kubeClient.CoreV1().Secrets(ns).Get(context.TODO(), selector.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %s", ns+"/"+selector.Name)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}

	return nil, errors.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}

func (c *Solver) getClient(ch *v1alpha1.ChallengeRequest) (*DNSClient, error) {
	config := &DNSProviderConfig{}
	err := json.Unmarshal(ch.Config.Raw, &config)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal config json")
	}

	if &config.TokenSecretRef != nil && &config.UserSecretRef != nil {
		data, err := c.getSecretVal(config.TokenSecretRef, ch.ResourceNamespace)
		if err != nil {
			return nil, err
		}

		config.Token = string(data)
	}

	if &config.UserSecretRef != nil {
		data, err := c.getSecretVal(config.UserSecretRef, ch.ResourceNamespace)
		if err != nil {
			return nil, err
		}
		config.User = string(data)
	}

	client := &DNSClient{
		client: &http.Client{},
		cfg:    config,
	}
	return client, nil
}
