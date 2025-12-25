// cli/main.go
// Usage examples:
//
//	go run ./cmd/mangahub status
//	go run ./cmd/mangahub --server http://localhost:8080 status
//	MANGAHUB_API=http://localhost:8080 go run ./cmd/mangahub status
//
// Configuration and request example:
//
//	opts := config.LoadOptions{APIEndpoint: "http://localhost:8080"}
//	mgr, _ := config.LoadWithOptions(opts)
//	client := api.NewClient(mgr.Data.BaseURL, mgr.Data.Token)
//	status, _ := client.GetServerStatus(context.Background())
//	fmt.Println(status.Overall)
package main

import "github.com/ngocan-dev/mangahub_/cli/cmd"

func main() {
	cmd.Execute()
}
