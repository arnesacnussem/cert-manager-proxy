# cert-manager proxy

This project provides a proxy server and a cert-manager webhook to proxy ACME DNS01 challenges.

DNS provider api call is handled by [libdns](https://github.com/libdns/libdns)

The server is intended to be used with the [cert-manager](https://github.com/cert-manager/cert-manager) project.

# ⚠ Warn ⚠

Because my use case, there is no need to issue a valid cert for the proxy server, I just use a self-signed cert, thus
the server only provide a http endpoint.

This project includes basic auth, but absolute no encryption, please use with caution.

It's recommend to use behind a reverse proxy with rate limit or maybe client certificate authentication.

# tl;dr

- write a [config file](#server-config) for the server
- write a config file for the webhook
- run the server
- helm install the webhook

# Why

What I need:

- Fine-grained access control by usernames and allowed sub-zones
- Not expose my DNS provider API token that have much larger scope than a sub-zone I want to allow
- I need a wildcard certificate, but with Let's Encrypt, only DNS01 challenge is supported
- One user can use multiple dns provider

What I have:

- [acmeproxy.pl](https://github.com/madcamel/acmeproxy.pl), but it does not support multiple dns provider, and wants to
  generate a cert for itself
- [lego](https://github.com/go-acme/lego), but it computes the txt record itself, which cert-manager already did

# Config and Run

### server config

This server config example is for using with cloudflare

For a list of supported dns provider, check [libdns](https://github.com/libdns)

Some dns provider is not include,
check [update-libdns-provider-list.go](./update-libdns-provider-list.go) for more details

```yaml
# server listening address
server: ip:port

# List of providers
providers:
  - # zone of the dns provider
    # it is a unique identifier of this provider
    # which also used to match user, the user will matched by suffix
    zone: example.com

    # dns provider name
    # check libdns for a full list of supported provider
    provider: cloudflare

    # configuration for the dns provider
    # please refer to the documentation of the provider at libdns for more details
    config:
      api_token: your_api_token_here

# List of users
users:
  - name: user
    token: token123
    allowedZone:
      - foo.example.com
      - bar.another.com
```

### webhook config

#### example issuer

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: example-issuer
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: user@example.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource that will be used to store the account's private key.
      name: example-issuer-account-key
    solvers:
      - dns01:
          webhook:
            groupName: example.com # groupName must match the one configured on webhook deployment (see Helm chart's values) !
            solverName: acmeproxy
            config:
              server: https://acmeproxy.example.com
              user: example
              token: example
```

or use secret to store auth info
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-issuer-secret
type: kubernetes.io/basic-auth
stringData:
  username: "example"
  password: "example"
  
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: example-issuer
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: user@example.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource that will be used to store the account's private key.
      name: example-issuer-account-key
    solvers:
      - dns01:
          webhook:
            groupName: example.com # groupName must match the one configured on webhook deployment (see Helm chart's values) !
            solverName: acmeproxy
            config:
              server: https://acmeproxy.example.com
              userSecretRef:
                name: example-issuer-secret
                key: username
              tokenSecretRef:
                name: example-issuer-secret
                key: token
```