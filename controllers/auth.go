package controllers

//packages
import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	gin "github.com/gin-gonic/gin"
	"gitlab.com/fibocloud/aws-billing/api_v2/databases"
	"gitlab.com/fibocloud/aws-billing/api_v2/form"
	"gitlab.com/fibocloud/aws-billing/api_v2/structs"
	"gitlab.com/fibocloud/aws-billing/api_v2/utils"
)

// AuthController struct
type AuthController struct {
	BaseController
}

// Init Controller
func (co AuthController) Init(router *gin.RouterGroup) {
	router.GET("/admin", co.Admin)         //Admin
	router.POST("/login", co.Login)        //Login
	router.POST("/register", co.Register)  //Register
	router.GET("/confirm/:id", co.Confirm) //Confirm
}

// LoginParams create body params
type LoginParams struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResult create body params
type LoginResult struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

// Login user
// @Summary Sign in user
// @Description Sign in user
// @Tags Auth
// @Accept json
// @Produce json
// @Param auth body LoginParams true "Auth"
// @Success 200 {object} structs.ResponseBody{body=LoginResult}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /auth/login [post]
func (co AuthController) Login(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params LoginParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	var user databases.SystemUser
	result := co.DB.Preload("AwsCredentials", "is_active = ?", true).Where("email = ?", params.Email).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			co.SetError(http.StatusNotFound, "Хэрэглэгч олдсонгүй")
		}
		return
	}

	if !user.IsActive {
		co.SetError(http.StatusNotFound, "Хэрэглэгчийн эрх баталгаажаагүй байна")
		return
	}

	if valid, err := utils.ComparePassword(user.Password, params.Password); !valid {
		if err != nil {
			co.SetError(http.StatusNotFound, "Нэвтрэх нэр эсвэл нууц үг буруу байна")
		}
		// co.SetError(http.StatusUnauthorized, err.Error())
		return
	}

	accessToken, refreshToken := utils.GenerateToken(user)
	co.SetBody(LoginResult{Token: accessToken, Refresh: refreshToken})
	return
}

// Confirm user
// @Summary Confirm user
// @Description Confirm user
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} structs.ResponseBody{body=LoginResult}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /auth/confirm/{id} [get]
func (co AuthController) Confirm(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	tx := co.DB.Begin()

	var confirm databases.ConfirmUser
	result := tx.Where("code = ?", c.Param("id")).First(&confirm)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			tx.Rollback()
			co.SetError(http.StatusNotFound, "Буруу код байна")
		}
		return
	}

	if confirm.IsUsed {
		tx.Rollback()
		co.SetError(http.StatusNotFound, "Хэрэглэгдсэн код байна")
		return
	}

	result = tx.Where("id = ?", confirm.UserID).Updates(&databases.SystemUser{IsActive: true})
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusNotFound, result.Error.Error())
	}

	confirm.IsUsed = true
	confirm.UsedDate = time.Now()

	result = tx.Save(confirm)
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusNotFound, result.Error.Error())
	}

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})

	tx.Commit()
	return
}

// Register systemUser
// @Summary Register systemUser
// @Description Add systemUser
// @Tags SystemUser
// @Accept json
// @Produce json
// @Param systemUser body form.SystemUserParams true "systemUser"
// @Success 200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /auth/register [post]
func (co AuthController) Register(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.SystemUserParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}

	tx := co.DB.Begin()

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

	result := tx.Create(&systemUser)
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	credentials := databases.AwsCredentials{
		UserID:    systemUser.Base.ID,
		IsActive:  false,
		SecretKey: params.SecretKey,
		AccessKey: params.AccessKey,
		Base: databases.Base{
			CreatedDate: time.Now(),
		},
	}

	result = tx.Create(&credentials)
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	code := utils.StringWithCharset(10)

	confirm := databases.ConfirmUser{
		UserID: systemUser.Base.ID,
		Code:   code,
		Base: databases.Base{
			CreatedDate: time.Now(),
		},
	}
	result = tx.Create(&confirm)
	if result.Error != nil {
		tx.Rollback()
		co.SetError(http.StatusInternalServerError, result.Error.Error())
		return
	}

	url := "https://crvoy9sfk2.execute-api.ap-east-1.amazonaws.com/default/fibobill-confirm-user"
	method := "POST"

	payload := strings.NewReader(`{"email": "` + params.Email + `", "code": "` + code + `"}`)

	fmt.Println(payload)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))

	co.SetBody(structs.SuccessResponse{
		Success: true,
	})

	tx.Commit()
	return
}

// Admin create
// @Summary Init admin account
// @Description Init admin account
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} structs.ResponseBody{body=databases.SystemUser}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /auth/admin [get]
func (co AuthController) Admin(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	hashPwd, err := utils.GenerateHash("Mongol123@")
	if err != nil {
		co.SetError(http.StatusInternalServerError, err.Error())
		return
	}

	user := databases.SystemUser{
		IsActive: true,
		Email:    "admin",
		Password: hashPwd,
		Base: databases.Base{
			CreatedDate: time.Now(),
		},
	}

	resultSystemUser := co.DB.Create(&user)
	if resultSystemUser.Error != nil {
		co.SetError(http.StatusInternalServerError, resultSystemUser.Error.Error())
		return
	}

	co.SetBody(user)

	return
}
