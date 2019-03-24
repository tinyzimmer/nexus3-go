package main

import (
	"fmt"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func deleteBlobStore() {
	client, err := nexus.New(*host, *username, *password)
	checkErr(err)
	err = client.DeleteBlobStore(&nexus.DeleteBlobStoreInput{
		Name: deleteBlobStoreName,
	})
	checkErr(err)
	fmt.Printf("Blobstore %s deleted\n", *deleteBlobStoreName)
}
