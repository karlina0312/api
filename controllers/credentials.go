package controllers

import (
	"net/http"
	"time"

	gin "github.com/gin-gonic/gin"
	databases "gitlab.com/fibocloud/aws-billing/api_v2/databases"
	form "gitlab.com/fibocloud/aws-billing/api_v2/form"
	structs "gitlab.com/fibocloud/aws-billing/api_v2/structs"
)

// CredentialsController struct
type CredentialsController struct {
	BaseController
}

// ListCredentials ...
type ListCredentials struct {
	Total int64                      `json:"total"`
	List  []databases.AwsCredentials `json:"list"`
}

// Init Controller
func (co CredentialsController) Init(router *gin.RouterGroup) {
	router.GET("/list", co.List)                     // List
	router.GET("get/:id", co.Get)                    // Show
	router.POST("", co.Create)                       // Create
	router.PUT("/:id", co.Update)                    // Update
	router.POST("/update/default", co.UpdateDefault) // Update
	router.DELETE("/:id", co.Delete)                 // Delete
}

// List credentials
// @Summary List credentials
// @Description Get credentials
// @Tags Credentials
// @Accept json
// @Produce json
// @Success 200 {object} structs.ResponseBody{body=[]databases.credentials}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /credentials/list [get]
func (co CredentialsController) List(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var credentials []databases.AwsCredentials
	co.DB.Where("user_id = ?", co.GetAuth(c).Base.ID).Find(&credentials)

	co.SetBody(credentials)
	return
}

// Get credentials
// @Summary Get credentials
// @Description Show credentials
// @Tags Credentials
// @Accept json
// @Produce json
// @Param id path uint true "credentials ID"
// @Success 200 {object} structs.ResponseBody{body=databases.credentials}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /credentials/{id} [get]
func (co CredentialsController) Get(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var credentials databases.AwsCredentials
	result := co.DB.First(&credentials, c.Param("id"))
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(credentials)
	return
}

// Create credentials
// @Summary Create credentials
// @Description Add credentials
// @Tags Credentials
// @Accept json
// @Produce json
// @Param credentials body form.credentialsParams true "credentials"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /credentials [post]
func (co CredentialsController) Create(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.CredentialsParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	credentials := databases.AwsCredentials{
		UserID:      co.GetAuth(c).Base.ID,
		Description: params.Description,
		IsActive:    true,
		SecretKey:   params.SecretKey,
		AccessKey:   params.AccessKey,
		Base: databases.Base{
			CreatedDate: time.Now(),
		},
	}

	result := co.DB.Create(&credentials)
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})
	return
}

// UpdateDefault credentials
// @Summary UpdateDefault credentials
// @Description UpdateDefault credentials
// @Tags Credentials
// @Accept json
// @Produce json
// @Param credentials body form.credentialsParams true "credentials"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /credentials/update/default [post]
func (co CredentialsController) UpdateDefault(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	tx := co.DB.Begin()

	var params form.CredentialsUpdateDefaultParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	result := tx.Model(&databases.AwsCredentials{}).Where("user_id = ?", co.GetAuth(c).Base.ID).Update("is_active", false)
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	result = tx.Where("id = ?", params.CredentialID).Updates(&databases.AwsCredentials{IsActive: true})
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	result = tx.Where("id = ?", co.GetAuth(c).Base.ID).Updates(&databases.SystemUser{AwsRegion: params.RegionCode})
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})
	tx.Commit()
	return
}

// Update credentials
// @Summary Update credentials
// @Description Edit credentials
// @Tags Credentials
// @Accept json
// @Produce json
// @Param id path uint true "credentials ID"
// @Param credentials body form.credentialsParams true "credentials"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /credentials/{id} [put]
func (co CredentialsController) Update(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.CredentialsUpdateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	var credentials databases.AwsCredentials
	result := co.DB.First(&credentials, c.Param("id"))
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	credentials.Description = params.Description
	credentials.IsActive = params.IsActive
	credentials.SecretKey = params.SecretKey
	credentials.AccessKey = params.AccessKey

	credentials.Base.ModifiedDate = time.Now()

	result = co.DB.Save(&credentials)
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})
	return
}

// Delete credentials
// @Summary Delete credentials
// @Description Remove credentials
// @Tags Credentials
// @Accept json
// @Produce json
// @Param id path uint true "credentials ID"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /credentials/{id} [delete]
func (co CredentialsController) Delete(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()
	result := co.DB.Delete(&databases.AwsCredentials{}, c.Param("id"))
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}
	co.SetBody(structs.SuccessResponse{
		Success: true,
	})
	return
}
