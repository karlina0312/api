package controllers

import (
	"net/http"
	"reflect"
	"time"

	gin "github.com/gin-gonic/gin"
	databases "gitlab.com/fibocloud/aws-billing/api_v2/databases"
	form "gitlab.com/fibocloud/aws-billing/api_v2/form"
	structs "gitlab.com/fibocloud/aws-billing/api_v2/structs"
	utils "gitlab.com/fibocloud/aws-billing/api_v2/utils"
)

// UserController struct
type UserController struct {
	BaseController
}

// ListSystemUsers ...
type ListSystemUsers struct {
	Total int64                  `json:"total"`
	List  []databases.SystemUser `json:"list"`
}

// Init Controller
func (co UserController) Init(router *gin.RouterGroup) {
	router.POST("/list", co.List)    // List
	router.GET("get/:id", co.Get)    // Show
	router.POST("", co.Create)       // Create
	router.PUT("/:id", co.Update)    // Update
	router.DELETE("/:id", co.Delete) // Delete
	router.GET("/me", co.Me)         // Me
}

// List systemUser
// @Summary List systemUser
// @Description Get systemUser
// @Tags SystemUser
// @Accept json
// @Produce json
// @Param filter body form.SystemUserFilter true "filter"
// @Success 200 {object} structs.ResponseBody{body=[]databases.SystemUser}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /systemUser/list [post]
func (co UserController) List(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var count int64
	var params form.SystemUserFilter
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	db := co.DB

	// filter hiij bgaa heseg
	v := reflect.ValueOf(params.Filter)

	db = db.Scopes(TableSearch(v, params.Sort))
	db = db.Scopes(Paginate(params.Page, params.Size))

	var listRepsonse ListSystemUsers

	var systemUsers []databases.SystemUser
	db.Find(&systemUsers)

	db.Table("med_base_systemUser").Count(&count)

	listRepsonse.List = systemUsers
	listRepsonse.Total = count

	co.SetBody(listRepsonse)
	return
}

// Get systemUser
// @Summary Get systemUser
// @Description Show systemUser
// @Tags SystemUser
// @Accept json
// @Produce json
// @Param id path uint true "systemUser ID"
// @Success 200 {object} structs.ResponseBody{body=databases.SystemUser}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /systemUser/{id} [get]
func (co UserController) Get(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var systemUser databases.SystemUser
	result := co.DB.First(&systemUser, c.Param("id"))
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(systemUser)
	return
}

// Create systemUser
// @Summary Create systemUser
// @Description Add systemUser
// @Tags SystemUser
// @Accept json
// @Produce json
// @Param systemUser body form.SystemUserParams true "systemUser"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /systemUser [post]
func (co UserController) Create(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.SystemUserParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	hashPwd, err := utils.GenerateHash(params.Password)
	if err != nil {
		co.SetError(http.StatusInternalServerError, err.Error())
		return
	}

	systemUser := databases.SystemUser{
		IsActive: params.IsActive,
		Email:    params.Email,
		Password: hashPwd,
		Base: databases.Base{
			CreatedDate: time.Now(),
		},
	}

	result := co.DB.Create(&systemUser)
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})
	return
}

// Update systemUser
// @Summary Update systemUser
// @Description Edit systemUser
// @Tags SystemUser
// @Accept json
// @Produce json
// @Param id path uint true "systemUser ID"
// @Param systemUser body form.SystemUserParams true "systemUser"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /systemUser/{id} [put]
func (co UserController) Update(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.SystemUserParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	var systemUser databases.SystemUser
	result := co.DB.First(&systemUser, c.Param("id"))
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	systemUser.IsActive = params.IsActive
	systemUser.Email = params.Email

	systemUser.Base.ModifiedDate = time.Now()

	result = co.DB.Save(&systemUser)
	if result.Error != nil {
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})
	return
}

// Delete systemUser
// @Summary Delete systemUser
// @Description Remove systemUser
// @Tags SystemUser
// @Accept json
// @Produce json
// @Param systemUser body form.DeleteParams true "systemUser"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /systemUser/{id} [delete]
func (co UserController) Delete(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.DeleteParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	for _, v := range params.IDs {
		result := co.DB.Delete(&databases.SystemUser{}, v)
		if result.Error != nil {
			co.SetError(http.StatusInternalServerError, result.Error.Error())
			return
		}
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})

	return
}

// Me get auth systemUser
// @Summary Get auth
// @Description Show auth
// @Tags SystemUser
// @Accept json
// @Produce json
// @Success 200 {object} structs.ResponseBody{body=databases.SystemUser}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /systemUser/me [get]
func (co UserController) Me(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var user databases.SystemUser
	co.DB.Debug().Preload("AwsCredentials", "is_active = ?", true).First(&user, co.GetAuth(c).Base.ID)

	co.SetBody(user)

	return
}
