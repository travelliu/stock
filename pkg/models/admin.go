package models

type CreateUserReq struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	TushareToken string `json:"tushareToken,omitempty"`
}

type PatchUserReq struct {
	Role         *string `json:"role,omitempty"`
	Disabled     *bool   `json:"disabled,omitempty"`
	TushareToken *string `json:"tushareToken,omitempty"`
}
