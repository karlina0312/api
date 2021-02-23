package databases

import "time"

type (
	// SystemUser [ Хэрэглэгч ]
	SystemUser struct {
		Base
		IsActive       bool            `gorm:"column:is_active;default:false" json:"is_active"` // Идэвхтэй эсэх
		Email          string          `gorm:"column:email;unique;not null" json:"email"`       // Нэвтрэх нэр
		Password       string          `gorm:"column:password;" json:"password"`                // Password
		AwsCredentials *AwsCredentials `gorm:"foreignKey:UserID" json:"aws_credentials"`        //
		AwsRegion      string          `gorm:"column:aws_region" json:"aws_region"`
	}

	// ConfirmUser ...
	ConfirmUser struct {
		Base
		User     *SystemUser `gorm:"foreignKey:UserID" json:"user"`     // Үүсгэсэн хэрэглэгч
		UserID   uint        `gorm:"column:user_id" json:"user_id"`     //
		Code     string      `gorm:"column:code" json:"code"`           //
		IsUsed   bool        `gorm:"column:is_used" json:"is_used"`     //
		UsedDate time.Time   `gorm:"column:used_date" json:"used_date"` //
	}

	// AwsCredentials [ AWS эрх ]
	AwsCredentials struct {
		Base
		User        *SystemUser `gorm:"foreignKey:UserID" json:"user"` // Үүсгэсэн хэрэглэгч
		UserID      uint        `gorm:"column:user_id" json:"user_id"` //
		Description string      `gorm:"column:description" json:"description"`
		IsActive    bool        `gorm:"column:is_active" json:"is_active"`
		IsDeleted   bool        `gorm:"column:is_deleted" json:"-"`
		SecretKey   string      `gorm:"column:secret_key" json:"-"`
		AccessKey   string      `gorm:"column:access_key" json:"access_key"`
	}

	// Company ...
	Company struct {
		Base
		IsActive bool   `gorm:"column:is_active;default:false" json:"is_active"` // Идэвхтэй эсэх
		Name     string `gorm:"column:name;unique;not null" json:"name"`         // Нэвтрэх нэр
	}
)
