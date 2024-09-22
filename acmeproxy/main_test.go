package main

import (
	"os"
)

func main() {
	tmp, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer os.Remove(tmp.Name())

	tmp.WriteString(`
server: 127.0.0.1:8088
providers:
  - zone: example.com
    provider: cloudflare
    config:
      api_token: your_token_here
users:
  - name: example
    token: abc123
    allowedZones:
      - foo.example.com
`)

	err = os.Setenv("CONFIG_PATH", tmp.Name())
	if err != nil {
		panic(err)
	}
}
