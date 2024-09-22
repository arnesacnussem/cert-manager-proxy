## webhook chart

[![Build Image](https://github.com/arnesacnussem/cert-manager-proxy/actions/workflows/docker-build.yaml/badge.svg)](https://github.com/arnesacnussem/cert-manager-proxy/actions/workflows/docker-build.yaml)
[![Build Image](https://github.com/arnesacnussem/cert-manager-proxy/actions/workflows/chart-release.yaml/badge.svg)](https://github.com/arnesacnussem/cert-manager-proxy/actions/workflows/chart-release.yaml)

For more details, see [README.md](https://github.com/arnesacnussem/cert-manager-proxy/blob/main/README.md)

## install & upgrade

```shell
helm upgrade --install -n cert-manager acmeproxy-webhook \
  --repo https://arnesacnussem.github.io/cert-manager-proxy/ acmeproxy-webhook \
  --set groupName=example.com \
  --set image.tag=latest
```
