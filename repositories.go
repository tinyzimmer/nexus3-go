package nexus

import (
	"encoding/json"
)

// Repository represents a Nexus repository
type Repository struct {
	Name   *string `json:"name"`
	Format *string `json:"format"`
	Type   *string `json:"type"`
	URL    *string `json:"url"`
}

// ListRepositories returns a list of the repositories available in Nexus
func (n *Nexus) ListRepositories() (res []*Repository, err error) {
	res = make([]*Repository, 0)
	req, err := n.NewRequest("GET", "service/rest/v1/repositories", nil, nil, "")
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
