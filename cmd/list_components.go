package main

import (
	"encoding/json"
	"fmt"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func listComponents() {
	client, err := nexus.New(*host, *username, *password)
	checkErr(err)
	input := &nexus.ListComponentsInput{
		Repository: listComponentsRepo,
	}
	components := make([]*nexus.Component, 0)
	err = client.ListComponentsPages(input, func(res *nexus.ListComponentsResponse, last bool) (bool, error) {
		for _, x := range res.Items {
			components = append(components, x)
		}
		return true, nil
	})
	checkErr(err)
	out, err := json.MarshalIndent(components, "", "    ")
	checkErr(err)
	fmt.Println(string(out))
}
