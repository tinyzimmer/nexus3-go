package main

import (
	"encoding/json"
	"fmt"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func createBlobStore() {
	client, err := nexus.New(*host, *username, *password)
	var btype *string
	if *createBlobStoreType == "file" {
		btype = nexus.BlobStoreTypeFile
	} else {
		btype = nexus.BlobStoreTypeS3
	}
	checkErr(err)
	store, err := client.CreateBlobStore(&nexus.CreateBlobStoreInput{
		Name: createBlobStoreName,
		Type: btype,
		Path: createBlobStorePath,
		S3Config: &nexus.S3BlobStoreConfig{
			Bucket:          createBlobStoreBucket,
			Prefix:          createBlobStorePrefix,
			AccessKeyID:     createBlobStoreAccessKeyID,
			SecretAccessKey: createBlobStoreSecretAccessKey,
			AssumeRole:      createBlobStoreAssumeRole,
			Region:          createBlobStoreRegion,
			Expiration:      createBlobStoreExpiration,
		},
	})
	checkErr(err)
	out, err := json.MarshalIndent(store, "", "    ")
	checkErr(err)
	fmt.Println(string(out))
}
