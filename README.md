# cert-manager proxy

This project provides a proxy server and a cert-manager webhook to proxy ACME DNS01 challenges.

DNS provider api call is handled by [libdns](https://github.com/libdns/libdns)

The server is intended to be used with the [cert-manager](https://github.com/cert-manager/cert-manager) project.

# ⚠ Warn

Because my use case, there is no need to issue a valid cert for the proxy server, I just use a self-signed cert, thus
the server only provide a http endpoint.

This project includes basic auth, but absolute no encryption, please use with caution.

It's recommend to use behind a reverse proxy with rate limit or maybe client certificate authentication.

# tl;dr

- write a [server config file](#example-server-config)
- [run the server](#acmeproxy-server)
- [install the webhook](#install-webhook)
- [create an issuer](#config-an-example-issuer) in your cluster

# Why

What I need:

- Fine-grained access control by usernames and allowed sub-zones
- Not expose my DNS provider API token that have much larger scope than a sub-zone I want to allow
- I need a wildcard certificate, but with Let's Encrypt, only DNS01 challenge is supported
- One user can use multiple dns provider
- Just found I also need to match a request by regex :)

What I have:

- [acmeproxy.pl](https://github.com/madcamel/acmeproxy.pl), but it does not support multiple dns provider, and wants to
  generate a cert for itself
- [lego](https://github.com/go-acme/lego), but it computes the txt record itself, which cert-manager already did

# Config and Run

## acmeproxy server

the image hosted on `ghcr.io`, config file location can be changed with env variable `CONFIG_PATH` pointing to a `.yaml`
config file, or use default config path `/config/config.yaml`

```shell
docker pull ghct.io/arnesacnussem/cert-manager-proxy/acmeproxy:latest
```

```shell
docker run -dp 8088:8088 \
  -v "$(pwd)/acmeproxy-config":/config \
  ghct.io/arnesacnussem/cert-manager-proxy/acmeproxy:latest
```

### example server config

For a list of supported dns provider, check [libdns](https://github.com/libdns).
Some dns provider is not include,
check [update-libdns-provider-list.go](./update-libdns-provider-list.go) for more details

```yaml
# server listening address
# this directly pass to gin
# see https://pkg.go.dev/github.com/gin-gonic/gin#Engine.Run for more detail
# e.g., server: 0.0.0.0:8088
# e.g., server: :8088
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
    allowedZones:
      # match zone by suffix
      - zone: foo.example.com # match *.foo.example.com or foo.example.com
      - zone: bar.another.com # match *.bar.another.com or bar.another.com

      # or use regex to match
      # be careful when using a regex, the code just run "regexp.Match"
      - regex: ^.+-foo.bar.com$ # match *-foo.bar.com
        # this time, the zone must be one in provider.zone
        zone: example.com
```

### webhook config

#### install webhook

```shell
helm upgrade --install -n cert-manager acmeproxy-webhook \
  --repo https://arnesacnussem.github.io/cert-manager-proxy/ acmeproxy-webhook \
  --set groupName=example.com \
  --set image.tag=latest
```

#### config an example issuer

by simply write your secret in it

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

or use a secret

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