package form

// CredentialsParams create body params
type CredentialsParams struct {
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// CredentialsUpdateParams create body params
type CredentialsUpdateParams struct {
	UserID      int    `json:"user_id"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// CredentialsUpdateDefaultParams ...
type CredentialsUpdateDefaultParams struct {
	CredentialID int    `json:"credential_id" binding:"required"`
	RegionCode   string `json:"region_code" binding:"required"`
}
