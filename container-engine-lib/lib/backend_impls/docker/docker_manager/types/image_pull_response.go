package types

import "strings"

const (
	m1ErrorString = "no matching manifest for linux/arm64/v8"
)

type ImagePullResponse struct {
	Error string `json:"error"`
}

func (resp *ImagePullResponse) IsArchitectureError() bool {
	if resp.Error == "" {
		return false
	}
	return strings.HasPrefix(resp.Error, m1ErrorString)
}
