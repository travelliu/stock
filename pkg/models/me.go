package models

import "time"

type ChangePasswordReq struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type SetTushareTokenReq struct {
	Token string `json:"token"`
}

type IssueTokenReq struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type IssueTokenResp struct {
	Token    string    `json:"token"`
	Metadata *APIToken `json:"metadata"`
}
