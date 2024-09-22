package main

// see https://github.com/orgs/libdns/repositories
import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Repository struct {
	Name string `json:"name"`
}

func listLibDNSRepos() ([]Repository, error) {
	url := "https://api.github.com/users/libdns/repos"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []Repository
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func importLine(name string) string {
	return fmt.Sprintf("\t\"github.com/libdns/%s\"", name)
}

func returnCase(name string) string {
	return fmt.Sprintf(`
	case "%s":
		return &%s.Provider{}
`, name, name)
}

func main() {
	repos, err := listLibDNSRepos()
	if err != nil {
		fmt.Println(err)
		return
	}

	blackList := make(map[string]string)
	blackList["libdns"] = "this is the interface package"
	blackList["template"] = "this is the template package"
	blackList["acmedns"] = "missing method GetRecords"
	blackList["dode"] = "missing method GetRecords"
	blackList["dnsmadeeasy"] = "verifying module: checksum mismatch"

	var imports []string
	var cases []string
	for _, repo := range repos {
		name := repo.Name

		if reason, ok := blackList[name]; ok {
			log.Printf("skipping %s because %s", name, reason)
			continue
		}
		imports = append(imports, importLine(name))
		cases = append(cases, returnCase(name))
	}

	fmt.Printf("Total repos to use: %d\n", len(imports))

	fileTemplate :=
		`package dns

import (
%s
)

// NewProviderByName see https://github.com/orgs/libdns/repositories
func NewProviderByName(name string) Provider {
	switch name {%s
	}
	return nil
}
`
	fileContent := fmt.Sprintf(fileTemplate, strings.Join(imports, "\n"), strings.Join(cases, ""))

	err = os.WriteFile("./cert-manager-acmeproxy/dns/all_providers.go", []byte(fileContent), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
