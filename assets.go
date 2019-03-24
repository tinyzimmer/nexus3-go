package nexus

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Asset represents an asset in Nexus and it's associated metadata
type Asset struct {
	DownloadURL *string            `json:"downloadUrl"`
	Path        *string            `json:"path"`
	ID          *string            `json:"id"`
	Repository  *string            `json:"repository"`
	Format      *string            `json:"format"`
	Checksum    *map[string]string `json:"checksum"`

	client *Nexus
}

// Delete this asset instance from Nexus
func (a *Asset) Delete() (err error) {
	return a.client.DeleteAsset(&DeleteAssetInput{ID: a.ID})
}

// Download this asset, returns a bytes representation of the object
//
// Example 1
//
// Download an asset to `test.tar.gz`
//
//     res, err := client.GetAsset(
//       &nexus.GetAssetInput{
//         ID: nexus.String("<some asset id>"),
//       },
//     )
//     if err != nil {
//       log.Fatal(err)
//     }
//     data, err := res.Download()
//     if err != nil {
//       return
//     }
//     file, err := os.Create("test.tar.gz")
//     defer file.Close()
//     file.Write(data)
//
// Example 2
//
// Recursively download an entire repository
//
//   input := &nexus.ListAssetsInput{Repository: nexus.String("my-repo")}
//   err := client.ListAssetsPages(input, func(res *nexus.ListAssetsResponse, last bool) (bool, error) {
//     for _, item := range res.Items {
//       data, err := item.Download()
//       if err != nil {
//         return false, err
//       }
//       file, _ := os.Create(*item.ID) // The GetComponents API produces better output
//       defer file.Close()             // and is better suited for an operation like this.
//       file.Write(data)               // The underlying asset objects have the same Download()
//     }                                // methods.
//     return true, nil
//   })
//
func (a *Asset) Download() (data []byte, err error) {
	endpoint := strings.Replace(*a.DownloadURL, a.client.host, "", 1)
	req, err := a.client.NewRequest("GET", endpoint, nil, nil, "")
	if err != nil {
		return
	}
	data, err = a.client.Do(req, nil, false)
	return
}

// GetAssetInput is used to provide parameters to GetAsset
type GetAssetInput struct {
	ID *string
}

// DeleteAssetInput is used to provide parameters to DeleteAsset
type DeleteAssetInput struct {
	ID *string
}

// ListAssetsInput is used to provide parameters to ListAssets
type ListAssetsInput struct {
	Repository        *string
	ContinuationToken *string
}

// ListAssetsResponse is a response from the ListAssets operation
type ListAssetsResponse struct {
	Items             []*Asset `json:"items"`
	ContinuationToken *string  `json:"continuationToken"`
}

func (n *Nexus) newListAssetsReq(input *ListAssetsInput) (req *http.Request, err error) {
	args := map[string]string{
		"repository": *input.Repository,
	}
	if input.ContinuationToken != nil {
		args["continuationToken"] = *input.ContinuationToken
	}
	req, err = n.NewRequest("GET", "service/rest/v1/assets", args, nil, "")
	return
}

func (n *Nexus) newGetAssetReq(input *GetAssetInput) (req *http.Request, err error) {
	endpoint := fmt.Sprintf("service/rest/v1/assets/%s", *input.ID)
	req, err = n.NewRequest("GET", endpoint, nil, nil, "")
	return
}

func (n *Nexus) newDeleteAssetReq(input *DeleteAssetInput) (req *http.Request, err error) {
	endpoint := fmt.Sprintf("service/rest/v1/assets/%s", *input.ID)
	req, err = n.NewRequest("DELETE", endpoint, nil, nil, "")
	return
}

// ListAssets returns a response with up to 10 assets and a token to request the next page.
func (n *Nexus) ListAssets(input *ListAssetsInput) (res *ListAssetsResponse, err error) {
	if input.Repository == nil {
		err = errors.New("Repository is required for ListAssets")
		return
	}
	req, err := n.newListAssetsReq(input)
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		403: fmt.Sprintf("Insufficient permissions to list assets in %s", *input.Repository),
		404: fmt.Sprintf("Repository %s does not exist", *input.Repository),
	}, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}
	if len(res.Items) > 0 {
		for _, x := range res.Items {
			x.client = n
		}
	}
	return
}

// ListAssetsPages iterates over all available asset pages for the given repository.
// The callback function provided on thee command line is called for each page
// with the response and a boolean representing if it's the last page or not.
// The function should return (bool, error). If bool is false or an error is returned
// the next page will not be retrieved.
//
// Example
//
// Iterate responses and print the Path of every asset
//
//   input := &nexus.ListAssetsInput{
//     Repository: nexus.String("my-repo"),
//   }
//   err = client.ListAssetsPages(input, func(res *nexus.ListAssetsResponse, last bool) (bool, error) {
//     for _, item := range res.Items {
//       log.Println(*item.Path)
//     }
//     return true, nil // true ensures we run the next page
//   })
//   if err != nil {
//     log.Fatal(err)
//   }
func (n *Nexus) ListAssetsPages(input *ListAssetsInput, cb func(res *ListAssetsResponse, last bool) (cont bool, err error)) error {
	res, err := n.ListAssets(input)
	if err != nil {
		return err
	}
	if res.ContinuationToken == nil {
		_, err = cb(res, true)
		return err
	}
	cont, err := cb(res, false)
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}
	newInput := &ListAssetsInput{
		Repository:        input.Repository,
		ContinuationToken: res.ContinuationToken,
	}
	return n.ListAssetsPages(newInput, cb)
}

// GetAsset retrieves an asset by the given ID.
func (n *Nexus) GetAsset(input *GetAssetInput) (res *Asset, err error) {
	if input.ID == nil {
		err = errors.New("Asset ID is required for GetAsset")
		return
	}
	req, err := n.newGetAssetReq(input)
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		403: fmt.Sprintf("Insufficient permissions to get asset %s", *input.ID),
		404: fmt.Sprintf("Asset %s does not exist", *input.ID),
		422: fmt.Sprintf("Malformed asset ID: %s", *input.ID),
	}, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}
	res.client = n
	return
}

// DeleteAsset removes an asset with the given ID.
func (n *Nexus) DeleteAsset(input *DeleteAssetInput) (err error) {
	if input.ID == nil {
		err = errors.New("Asset ID is required for DeleteAsset")
		return
	}
	req, err := n.newDeleteAssetReq(input)
	if err != nil {
		return
	}
	_, err = n.Do(req, map[int]string{
		403: fmt.Sprintf("Insufficient permissions to delete %s", *input.ID),
		404: fmt.Sprintf("Asset %s does not exist", *input.ID),
		422: fmt.Sprintf("Malformed asset ID: %s", *input.ID),
	}, false)
	return
}
