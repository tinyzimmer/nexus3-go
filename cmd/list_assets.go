package main

import (
	"encoding/json"
	"fmt"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func listAssets() {
	client, err := nexus.New(*host, *username, *password)
	checkErr(err)
	input := &nexus.ListAssetsInput{
		Repository: listAssetsRepo,
	}
	assets := make([]*nexus.Asset, 0)
	err = client.ListAssetsPages(input, func(res *nexus.ListAssetsResponse, last bool) (bool, error) {
		for _, x := range res.Items {
			assets = append(assets, x)
		}
		return true, nil
	})
	checkErr(err)
	out, err := json.MarshalIndent(assets, "", "    ")
	checkErr(err)
	fmt.Println(string(out))
}
