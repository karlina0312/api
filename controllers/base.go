package controllers

import (
	"errors"
	"net/http"
	"reflect"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"gitlab.com/fibocloud/aws-billing/api_v2/databases"
	"gitlab.com/fibocloud/aws-billing/api_v2/form"
	structs "gitlab.com/fibocloud/aws-billing/api_v2/structs"
	gorm "gorm.io/gorm"
)

// BaseController struct
type BaseController struct {
	Response *structs.Response
	DB       *gorm.DB
}

// DefaultSvc ...
func (co BaseController) DefaultSvc(userID uint) (sess *session.Session, err error) {
	var user databases.SystemUser

	result := co.DB.Preload("AwsCredentials").First(&user, userID)
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return nil, result.Error
	}

	if user.AwsCredentials.Base.ID != 0 {
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(user.AwsRegion), //"us-east-1"
			Credentials: awsCredentials.NewStaticCredentials(user.AwsCredentials.AccessKey, user.AwsCredentials.SecretKey, ""),
		})
		return sess, err

	}
	return sess, errors.New("You don't have a permission to access AWS")
}

// ListResponse ...
type ListResponse struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

// SetBody successfully response
func (co BaseController) SetBody(body interface{}) {
	co.Response.StatusCode = http.StatusOK
	co.Response.Body.StatusCode = 0
	co.Response.Body.ErrorMsg = ""
	co.Response.Body.Body = body
}

// SetError rrror response
func (co BaseController) SetError(code int, message string) {
	co.Response.StatusCode = 200
	co.Response.Body.StatusCode = code
	co.Response.Body.ErrorMsg = message
	co.Response.Body.Body = nil
}

// GetBody in response
func (co BaseController) GetBody() (int, interface{}) {
	return co.Response.StatusCode, co.Response.Body
}

// GetAuth get auth user
func (co BaseController) GetAuth(c *gin.Context) databases.SystemUser {
	if iauth, exists := c.Get("auth"); exists {
		return iauth.(databases.SystemUser)
	}
	return databases.SystemUser{}
}

// Paginate table
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// TableSearch undsen table search hiih
func TableSearch(v reflect.Value, sort form.SortColumn) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		typeOfS := v.Type()
		for i := 0; i < v.NumField(); i++ {

			filterName := typeOfS.Field(i).Name
			filterJSON := typeOfS.Field(i).Tag.Get("json")

			if !strings.Contains(filterName, "External") {
				switch v.Field(i).Interface().(type) {
				case int:
					fitlerValue := v.Field(i).Interface().(int)
					if fitlerValue > 0 {
						db = db.Where(filterJSON+" = ?", fitlerValue)
					}

				case string:
					fitlerValue := v.Field(i).Interface().(string)

					isDate := strings.Contains(strings.ToLower(filterName), "date")

					if fitlerValue != "" && fitlerValue != "0" {
						if filterName == "IsImport" || filterName == "IsActive" {
							boolValue := fitlerValue == "true"
							db = db.Where(filterJSON+" = ?", boolValue)
						}
						if isDate {
							db = db.Where(filterJSON+" >= ?", fitlerValue)
						}

						if !isDate && filterName != "IsImport" && filterName != "IsActive" {
							db = db.Where("LOWER(CAST("+filterJSON+" as TEXT))"+" LIKE ?", "%"+strings.ToLower(fitlerValue)+"%")
						}
					}
				}
			}
		}

		if sort.Field != "" {
			db = db.Order(sort.Field + " " + sort.Order)
		} else {
			db = db.Order("created_date desc")
		}
		return db
	}
}
