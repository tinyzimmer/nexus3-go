package nexus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type UploadComponentInput struct {
	Repository      *string
	ComponentType   *string
	ComponentConfig *map[string]string
	Assets          []*UploadComponentAsset
}

type UploadComponentAsset struct {
	File        *os.File
	AssetConfig *map[string]string
}

// ListComponentsResponse is a response from a ListCompoents call
type ListComponentsResponse struct {
	Items             []*Component `json:"items"`
	ContinuationToken *string      `json:"continuationToken"`
}

// Component is an artifact in the repository with all of it's metadata
// and objects.
type Component struct {
	ID         *string  `json:"id"`
	Repository *string  `json:"repository"`
	Format     *string  `json:"format"`
	Group      *string  `json:"group"`
	Name       *string  `json:"name"`
	Version    *string  `json:"version"`
	Assets     []*Asset `json:"assets"`

	client *Nexus
}

// GetComponentInput is used to provide parameters to GetComponent
type GetComponentInput struct {
	ID *string
}

// DeleteComponentInput is used to provide parameters to DeleteComponent
type DeleteComponentInput struct {
	ID *string
}

// ListComponentsInput is used to provide parameters to ListComponents
type ListComponentsInput struct {
	Repository        *string
	ContinuationToken *string
}

func (n *Nexus) newListComponentsReq(input *ListComponentsInput) (req *http.Request, err error) {
	args := map[string]string{
		"repository": *input.Repository,
	}
	if input.ContinuationToken != nil {
		args["continuationToken"] = *input.ContinuationToken
	}
	req, err = n.NewRequest("GET", "service/rest/v1/components", args, nil, "")
	return
}

func (n *Nexus) newGetComponentReq(input *GetComponentInput) (req *http.Request, err error) {
	endpoint := fmt.Sprintf("service/rest/v1/components/%s", *input.ID)
	req, err = n.NewRequest("GET", endpoint, nil, nil, "")
	return
}

func (n *Nexus) newDeleteComponentReq(input *DeleteComponentInput) (req *http.Request, err error) {
	endpoint := fmt.Sprintf("service/rest/v1/components/%s", *input.ID)
	req, err = n.NewRequest("DELETE", endpoint, nil, nil, "")
	return
}

func (n *Nexus) newUploadBody(input *UploadComponentInput) (bodyBytes []byte, contentType string, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	if input.ComponentConfig != nil {
		for k, v := range *input.ComponentConfig {
			key := fmt.Sprintf("%s.%s", *input.ComponentType, k)
			writer.WriteField(key, v)
		}
	}
	if len(input.Assets) == 1 {
		asset := input.Assets[0]
		key := fmt.Sprintf("%s.asset", *input.ComponentType)
		part, err := writer.CreateFormFile(key, filepath.Base(asset.File.Name()))
		if err != nil {
			return nil, "", err
		}
		_, err = io.Copy(part, asset.File)
		if err != nil {
			return nil, "", err
		}
		if asset.AssetConfig != nil {
			for k, v := range *asset.AssetConfig {
				key := fmt.Sprintf("%s.asset.%s", *input.ComponentType, k)
				writer.WriteField(key, v)
			}
		}
	} else {
		for idx, asset := range input.Assets {
			key := fmt.Sprintf("%s.asset%v", *input.ComponentType, idx)
			part, err := writer.CreateFormFile(key, filepath.Base(asset.File.Name()))
			if err != nil {
				return nil, "", err
			}
			_, err = io.Copy(part, asset.File)
			if err != nil {
				return nil, "", err
			}
			if asset.AssetConfig != nil {
				for k, v := range *asset.AssetConfig {
					key := fmt.Sprintf("%s.asset%v.%s", *input.ComponentType, idx, k)
					writer.WriteField(key, v)
				}
			}
		}
	}
	contentType = writer.FormDataContentType()
	bodyBytes, _ = ioutil.ReadAll(body)
	return
}

func containsKey(dict map[string]string, str string) bool {
	for k := range dict {
		if str == k {
			return true
		}
	}
	return false
}

func hasAllRequiredFields(present map[string]string, required []string) bool {
	for _, x := range required {
		if containsKey(present, x) {
			continue
		} else {
			return false
		}
	}
	return true
}

func (n *Nexus) validateComponentFormat(input *UploadComponentInput) (err error) {
	if input.Assets == nil || len(input.Assets) == 0 {
		err = errors.New("At least one asset must be provided to upload a component")
		return
	}
	if input.ComponentType == nil {
		err = errors.New("ComponentType is required for UploadComponent")
		return
	}
	format, err := n.GetFormat(*input.ComponentType)
	if err != nil {
		return
	}
	if format.ComponentFields != nil && len(format.ComponentFields) > 0 {
		requiredComponentFields := make([]string, 0)
		for _, field := range format.ComponentFields {
			if !*field.Optional {
				requiredComponentFields = append(requiredComponentFields, *field.Name)
			}
		}
		if input.ComponentConfig == nil {
			err = fmt.Errorf("%s requires the following component fields: %v", *input.ComponentType, requiredComponentFields)
			return
		}
		if !hasAllRequiredFields(*input.ComponentConfig, requiredComponentFields) {
			err = fmt.Errorf("%s requires the following component fields: %v", *input.ComponentType, requiredComponentFields)
			return
		}
	}
	if format.AssetFields != nil && len(format.AssetFields) > 0 {
		requiredAssetFields := make([]string, 0)
		for _, field := range format.AssetFields {
			if !*field.Optional && *field.Name != "asset" {
				requiredAssetFields = append(requiredAssetFields, *field.Name)
			}
		}
		for _, asset := range input.Assets {
			if len(requiredAssetFields) == 0 {
				continue
			}
			if asset.AssetConfig == nil {
				err = fmt.Errorf("%s requires the following asset fields: %v", *input.ComponentType, requiredAssetFields)
				return
			}
			if !hasAllRequiredFields(*asset.AssetConfig, requiredAssetFields) {
				err = fmt.Errorf("%s requires the following asset fields: %v", *input.ComponentType, requiredAssetFields)
				return
			}
		}
	}
	return
}

func (n *Nexus) newUploadComponentReq(input *UploadComponentInput) (req *http.Request, err error) {
	args := map[string]string{
		"repository": *input.Repository,
	}
	body, ctype, err := n.newUploadBody(input)
	if err != nil {
		return
	}
	req, err = n.NewRequest("POST", "service/rest/v1/components", args, body, ctype)
	return
}

// UploadComponent uploads a component with the given parameters.
func (n *Nexus) UploadComponent(input *UploadComponentInput) (err error) {
	err = n.validateComponentFormat(input)
	if err != nil {
		return
	}
	req, err := n.newUploadComponentReq(input)
	if err != nil {
		return
	}
	_, err = n.Do(req, map[int]string{
		403: "Insufficient permissions to upload component",
		404: fmt.Sprintf("Repository %s does not exist", *input.Repository),
	}, false)
	return
}

// ListComponents returns a response with up to 10 components and a token to request the next page.
func (n *Nexus) ListComponents(input *ListComponentsInput) (res *ListComponentsResponse, err error) {
	if input.Repository == nil {
		err = errors.New("Repository is required for ListComponents")
		return
	}
	req, err := n.newListComponentsReq(input)
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		403: fmt.Sprintf("Insufficient permissions to list components in %s", *input.Repository),
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
			for _, y := range x.Assets {
				y.client = n
			}
		}
	}
	return
}

// ListComponentsPages is identical in usage to ListAssetsPages
func (n *Nexus) ListComponentsPages(input *ListComponentsInput, cb func(res *ListComponentsResponse, last bool) (cont bool, err error)) error {
	res, err := n.ListComponents(input)
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
	newInput := &ListComponentsInput{
		Repository:        input.Repository,
		ContinuationToken: res.ContinuationToken,
	}
	return n.ListComponentsPages(newInput, cb)
}

// GetComponent retrieves a component by the given ID.
func (n *Nexus) GetComponent(input *GetComponentInput) (res *Component, err error) {
	if input.ID == nil {
		err = errors.New("Component ID is required for GetAsset")
		return
	}
	req, err := n.newGetComponentReq(input)
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		403: fmt.Sprintf("Insufficient permissions to get component %s", *input.ID),
		404: fmt.Sprintf("Component %s does not exist", *input.ID),
		422: fmt.Sprintf("Malformed component ID: %s", *input.ID),
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

// DeleteComponent removes a component with the given ID.
func (n *Nexus) DeleteComponent(input *DeleteComponentInput) (err error) {
	if input.ID == nil {
		err = errors.New("Component ID is required for DeleteComponent")
		return
	}
	req, err := n.newDeleteComponentReq(input)
	if err != nil {
		return
	}
	_, err = n.Do(req, map[int]string{
		403: fmt.Sprintf("Insufficient permissions to delete %s", *input.ID),
		404: fmt.Sprintf("Component %s does not exist", *input.ID),
		422: fmt.Sprintf("Malformed component ID: %s", *input.ID),
	}, false)
	return
}
