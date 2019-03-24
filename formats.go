package nexus

import (
	"encoding/json"
	"fmt"
)

type Format struct {
	Name            *string           `json:"format"`
	MultipleUpload  *bool             `json:"multipleUpload"`
	ComponentFields []*ComponentField `json:"componentFields"`
	AssetFields     []*AssetField     `json:"assetFields"`
}

type ComponentField struct {
	Name        *string `json:"name"`
	Type        *string `json:"type"`
	Description *string `json:"description"`
	Optional    *bool   `json:"optional"`
	Group       *string `json:"group"`
}

type AssetField struct {
	Name        *string `json:"name"`
	Type        *string `json:"type"`
	Description *string `json:"description"`
	Optional    *bool   `json:"optional"`
	Group       *string `json:"group"`
}

// GetFormat retrieves a single repository type's format
func (n *Nexus) GetFormat(format string) (res *Format, err error) {
	endpoint := fmt.Sprintf("service/rest/v1/formats/%s/upload-specs", format)
	req, err := n.NewRequest("GET", endpoint, nil, nil, "")
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		404: fmt.Sprintf("The format %s does not exist", format),
	}, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &res)
	return
}

// ListFormats returns a list of the available repository formats
func (n *Nexus) ListFormats() (res []*Format, err error) {
	res = make([]*Format, 0)
	req, err := n.NewRequest("GET", "service/rest/v1/formats/upload-specs", nil, nil, "")
	if err != nil {
		return
	}
	body, err := n.Do(req, nil, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &res)
	return
}
