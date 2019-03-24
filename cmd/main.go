package main

import (
	"fmt"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New(os.Args[0], "A command-line interface for Sonatype Nexus 3.")

	host     = app.Flag("host", "URL of the Nexus host").Short('h').Default("http://localhost:8081").String()
	username = app.Flag("username", "Username to authenticate to Nexus").Short('u').Default("admin").String()
	password = app.Flag("password", "Password to authenticate to Nexus").Short('p').Default("admin123").String()

	scriptExec = app.Command("groovy-exec", "Execute a groovy script on the Nexus host")
	scriptArgs = scriptExec.Arg("commands", "Groovy commands to execute").Strings()
	scriptFile = scriptExec.Flag("script", "A script file to execute").Short('s').ExistingFile()

	listReposCmd      = app.Command("list-repositories", "List the repositories in Nexus")
	listBlobStoresCmd = app.Command("list-blob-stores", "List the blob stores in Nexus")
	listFormatsCmd    = app.Command("list-formats", "List the available component formats")

	listAssetsCmd  = app.Command("list-assets", "List the assets for a given repository")
	listAssetsRepo = listAssetsCmd.Arg("repository", "The repository to list assets for").Required().String()

	listComponentsCmd  = app.Command("list-components", "List the components for a given repository")
	listComponentsRepo = listComponentsCmd.Arg("repository", "The repository to list components for").Required().String()

	uploadComponentCmd  = app.Command("upload-component", "Upload a component to a given repository")
	uploadComponentRepo = uploadComponentCmd.Flag("repository", "The repository to upload the component to").Short('r').Required().String()
	uploadComponentType = uploadComponentCmd.Flag("type", "The type of the component").Short('t').Required().String()
	uploadComponentFile = uploadComponentCmd.Flag("file", "The file to upload").Short('f').Required().ExistingFile()

	createBlobStoreCmd             = app.Command("create-blobstore", "Create a new blob store")
	createBlobStoreName            = createBlobStoreCmd.Flag("name", "The name of the blobstore").Short('n').Required().String()
	createBlobStoreType            = createBlobStoreCmd.Flag("type", "The type of the blob store").Short('t').Default("file").Enum("file", "s3")
	createBlobStorePath            = createBlobStoreCmd.Flag("path", "The path to the blob store when type is file").String()
	createBlobStoreBucket          = createBlobStoreCmd.Flag("bucket", "The s3 bucket when creating an s3 blob store").String()
	createBlobStorePrefix          = createBlobStoreCmd.Flag("prefix", "The s3 bucket prefix for the blob store").String()
	createBlobStoreAccessKeyID     = createBlobStoreCmd.Flag("access-key-id", "The AWS IAM AccessKeyID").String()
	createBlobStoreSecretAccessKey = createBlobStoreCmd.Flag("secret-access-key", "The AWS IAM SecretAccessKey").String()
	createBlobStoreAssumeRole      = createBlobStoreCmd.Flag("assume-role", "The AWS IAM Role to assume").String()
	createBlobStoreRegion          = createBlobStoreCmd.Flag("region", "The AWS region to use").String()
	createBlobStoreExpiration      = createBlobStoreCmd.Flag("expiry-days", "The number of days to wait to expire deleted blobs").Default("-1").Int()

	deleteBlobStoreCmd   = app.Command("delete-blobstore", "Delete a blobstore by the given name")
	deleteBlobStoreName  = deleteBlobStoreCmd.Arg("blobstore", "The name of the blob store to delete").String()
	deleteBlobStoreForce = deleteBlobStoreCmd.Flag("force", "Force deletion of an in-use blobstore").Bool()
)

func checkErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err.Error())
		os.Exit(2)
	}
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case scriptExec.FullCommand():
		executeScript()
	case listReposCmd.FullCommand():
		listRepos()
	case listBlobStoresCmd.FullCommand():
		listBlobStores()
	case listAssetsCmd.FullCommand():
		listAssets()
	case listComponentsCmd.FullCommand():
		listComponents()
	case createBlobStoreCmd.FullCommand():
		createBlobStore()
	case deleteBlobStoreCmd.FullCommand():
		deleteBlobStore()
	case listFormatsCmd.FullCommand():
		listFormats()
	case uploadComponentCmd.FullCommand():
		uploadComponent()
	default:
		app.Usage(nil)
		os.Exit(1)
	}
	os.Exit(0)
}
