package main

import (
	"encoding/json"
	"fmt"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func listFormats() {
	client, err := nexus.New(*host, *username, *password)
	checkErr(err)
	res, err := client.ListFormats()
	checkErr(err)
	out, err := json.MarshalIndent(res, "", "    ")
	checkErr(err)
	fmt.Println(string(out))
}
