package form

// SystemUserParams create body params
type SystemUserParams struct {
	IsActive  bool   `json:"is_active"`                   // Идэвхтэй эсэх
	Email     string `json:"email" binding:"required"`    // Нэр
	Password  string `json:"password" binding:"required"` // Нууц үг
	AccessKey string `json:"access_key" binding:"required"`
	SecretKey string `json:"secret_key" binding:"required"`
}

// SystemUserFilterCols sort hiih bolomjtoi column
type SystemUserFilterCols struct {
	Username  string `json:"username"` // Нэр
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
	IsActive  bool   `json:"is_active"` // Идэвхтэй эсэх
}

// SystemUserFilter sort hiigdej boloh zuils
type SystemUserFilter struct {
	Page   int                  `json:"page"`
	Size   int                  `json:"size"`
	Sort   SortColumn           `json:"sort"`
	Filter SystemUserFilterCols `json:"filter"`
}
