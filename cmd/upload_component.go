package main

import (
	"fmt"
	"os"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func uploadComponent() {
	client, err := nexus.New(*host, *username, *password)
	checkErr(err)
	file, err := os.Open(*uploadComponentFile)
	checkErr(err)
	asset := &nexus.UploadComponentAsset{
		File: file,
	}
	err = client.UploadComponent(&nexus.UploadComponentInput{
		Repository:    uploadComponentRepo,
		ComponentType: uploadComponentType,
		Assets: []*nexus.UploadComponentAsset{
			asset,
		},
	})
	checkErr(err)
	fmt.Println("Component uploaded successfully")
}
