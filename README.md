# prefab-cloud-go

Go Client for Prefab Feature Flags, Dynamic log levels, and Config as a Service: https://www.prefab.cloud

## Installation

```bash
go get github.com/prefab-cloud/prefab-cloud-go@latest
```

## Basic example

```go
package main

import (
	"fmt"
	"log"
	"os"

	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
)

func main() {
	apiKey, exists := os.LookupEnv("PREFAB_API_KEY")

	if !exists {
		log.Fatal("API Key not found")
	}

	client, err := prefab.NewClient(prefab.WithAPIKey(apiKey))

	if err != nil {
		log.Fatal(err)
	}

	val, ok, err := client.GetStringValue("my.string.config", prefab.ContextSet{})

	if err != nil {
		log.Fatal(err)
	}

	if !ok {
		log.Fatal("Value not found")
	}

	fmt.Println(val)
}
```

## Documentation

- [API Reference](https://pkg.go.dev/github.com/prefab-cloud/prefab-cloud-go/pkg)

## Notable pending features

- Telemetry
- JSON dump data source
