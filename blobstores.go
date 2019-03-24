package nexus

import (
	"encoding/json"
	"fmt"
)

// BlobStoreTypeFile is used for creating blobstores on the filesystem
var BlobStoreTypeFile = String("File")

// BlobStoreTypeS3 is used for creating S3-backed blob stores
var BlobStoreTypeS3 = String("S3")

var createBlobStoreScriptName = String("nexus3-go-create-blobstore")
var createBlobStoreScript = String(`
import groovy.json.JsonSlurper

parsed_args = new JsonSlurper().parseText(args)
existingBlobStore = blobStore.getBlobStoreManager().get(parsed_args.name)
if (existingBlobStore == null) {
  if (parsed_args.type == "S3") {
      blobStore.createS3BlobStore(parsed_args.name, parsed_args.config)
      msg = "created"
  } else {
      blobStore.createFileBlobStore(parsed_args.name, parsed_args.path)
      msg = "created"
  }
} else {
    msg = "exists"
}
return msg
`)

var deleteBlobStoreScriptName = String("nexus3-go-delete-blobstore")
var deleteBlobStoreScript = String(`
import groovy.json.JsonSlurper

parsed_args = new JsonSlurper().parseText(args)
existingBlobStore = blobStore.getBlobStoreManager().get(parsed_args.name)
if (existingBlobStore != null) {
	if (parsed_args.force) {
		blobStore.getBlobStoreManager().forceDelete(parsed_args.name)
		msg = "deleted"
	} else {
		blobStore.getBlobStoreManager().delete(parsed_args.name)
		msg = "deleted"
	}
} else {
	msg = "not exists"
}
return msg
`)

var listBlobStoreScriptName = String("nexus3-go-list-blobstores")
var listBlobStoreScript = String(`
import groovy.json.JsonOutput

def res = []

blobStore.blobStoreManager.browse()*.each { store ->
	 def storeMap = [:]
   props = store.getProperties()
	 props.each { k, v ->
		 if (v instanceof String || v instanceof Boolean || v instanceof Integer) {
			 storeMap[k] = v
		 } else {
			 storeMap[k] = [:]
			 v.getProperties().each { x, y ->
				 if (y instanceof String || y instanceof Boolean || y instanceof Integer) {
					 storeMap[k][x] = y
				 }
			 }
		 }
	 }
	 res << storeMap
}
def json = JsonOutput.toJson(res)
return json
`)

// BlobStore represents a blobstore instance
type BlobStore struct {
	Groupable        *bool            `json:"groupable"`
	ContentDir       *BlobDir         `json:"contentDir"`
	Metrics          *Metrics         `json:"metrics"`
	BlobIDStream     *BlobIDStream    `json:"blobIdStream"`
	RelativeBlobDir  *BlobDir         `json:"relativeBlobDir"`
	AbsoluteBlobDir  *BlobDir         `json:"absoluteBlobDir"`
	Writable         *bool            `json:"writable"`
	StorageAvailable *bool            `json:"storageAvailable"`
	Config           *BlobStoreConfig `json:"blobStoreConfiguration"`
	Started          *bool            `json:"started"`
	StateGuard       *StateGuard      `json:"stateGuard"`
}

// BlobDir is part of the metadata of a blobstore
type BlobDir struct {
	Absolute                *bool   `json:"absolute"`
	PathForPermissionCheck  *string `json:"pathForPermissionCheck"`
	NameCount               *int    `json:"nameCount"`
	PathForExceptionMessage *string `json:"pathForExceptionMessage"`
	Empty                   *bool   `json:"empty"`
}

// Metrics is part of the metadata of a blobstore
type Metrics struct {
	Unlimited *bool `json:"unlimited"`
}

// BlobIDStream is part of the metadata of a blobstore
type BlobIDStream struct {
	Parallel         *bool `json:"parallel"`
	StreamFlags      *int  `json:"streamFlags"`
	Ordered          *bool `json:"ordered"`
	StreamAndOpFlags *int  `json:"streamAndOpFlags"`
}

// BlobStoreConfig is the configuration of a blob store, and contains fields
// such as the name and type.
type BlobStoreConfig struct {
	Writable *bool   `json:"writable"`
	Type     *string `json:"type"`
	Name     *string `json:"name"`
}

// StateGuard is part of the metadata for a blobstore
type StateGuard struct {
	Current *string `json:"current"`
}

// BlobStoreQuotaStatus is a response from GetBlobStoreQuotaStatus operation.
type BlobStoreQuotaStatus struct {
	IsViolation   *bool   `json:"isViolation"`
	Message       *string `json:"message"`
	BlobStoreName *string `json:"blobStoreName"`
}

// CreateBlobStoreInput provides parameters to a CreateBlobStore call.
// Type must be one of BlobStoreTypeFile or BlobStoreTypeS3. For a File type
// provide a path, for an s3 type provide an S3BlobStoreConfig.
type CreateBlobStoreInput struct {
	Name     *string            `json:"name"`
	Type     *string            `json:"type"`
	Path     *string            `json:"path"`
	S3Config *S3BlobStoreConfig `json:"config"`
}

// S3BlobStoreConfig represents an S3 bucket configuration for a blob store.
// Expiration is required and is the time (in days deleted blobs last in the bucket.
// To disable supply -1
type S3BlobStoreConfig struct {
	Bucket          *string `json:"bucket"`
	Prefix          *string `json:"prefix"`
	AccessKeyID     *string `json:"accessKeyId"`
	SecretAccessKey *string `json:"secretAccessKey"`
	SessionToken    *string `json:"sessionToken"`
	AssumeRole      *string `json:"assumeRole"`
	Region          *string `json:"region"`
	Endpoint        *string `json:"endpoint"`
	Expiration      *int    `json:"expiration"`
	SignerType      *string `json:"signerType"`
}

// DeleteBlobStoreInput provides paraameters to a DeleteBlobStore call
type DeleteBlobStoreInput struct {
	Name  *string `json:"name"`
	Force *bool   `json:"force"`
}

// ListBlobStores returns a list of the blobstores on the Nexus server
func (n *Nexus) ListBlobStores() (blobstores []*BlobStore, err error) {
	blobstores = make([]*BlobStore, 0)
	script := &Script{
		Name:    listBlobStoreScriptName,
		Type:    ScriptTypeGroovy,
		Content: listBlobStoreScript,
		client:  n,
	}
	res, err := script.ensureAndExecute(nil)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(*res.Result), &blobstores)
	return
}

// GetBlobStoreQuotaStatus retrieves the blobstore quota status for the given id
func (n *Nexus) GetBlobStoreQuotaStatus(id string) (res *BlobStoreQuotaStatus, err error) {
	res = &BlobStoreQuotaStatus{}
	endpoint := fmt.Sprintf("/v1/blobstores/%s/quota-status", id)
	req, err := n.NewRequest("GET", endpoint, nil, nil, "")
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		404: fmt.Sprintf("No status with id %s found", id),
	}, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &res)
	return
}

// GetBlobStore retrieves a blobstore by the given name
func (n *Nexus) GetBlobStore(name string) (store *BlobStore, err error) {
	blobstores, err := n.ListBlobStores()
	if err != nil {
		return
	}
	for _, x := range blobstores {
		if *x.Config.Name == name {
			store = x
			return
		}
	}
	err = fmt.Errorf("Blobstore %s does not exist", name)
	return
}

// CreateBlobStore creates a new blob store with the given parameters
func (n *Nexus) CreateBlobStore(input *CreateBlobStoreInput) (blobstore *BlobStore, err error) {
	script := &Script{
		Name:    createBlobStoreScriptName,
		Type:    ScriptTypeGroovy,
		Content: createBlobStoreScript,
		client:  n,
	}
	res, err := script.ensureAndExecute(input)
	if err != nil {
		return
	}
	if *res.Result == "exists" {
		err = fmt.Errorf("Blobstore %s already exists", *input.Name)
		return
	}
	blobstore, err = n.GetBlobStore(*input.Name)
	return
}

// DeleteBlobStore deletes a blobstore with the given parameters
func (n *Nexus) DeleteBlobStore(input *DeleteBlobStoreInput) (err error) {
	script := &Script{
		Name:    deleteBlobStoreScriptName,
		Type:    ScriptTypeGroovy,
		Content: deleteBlobStoreScript,
		client:  n,
	}
	res, err := script.ensureAndExecute(input)
	if err != nil {
		return
	}
	if *res.Result == "not exists" {
		err = fmt.Errorf("Blobstore %s does not exist", *input.Name)
	}
	return
}
