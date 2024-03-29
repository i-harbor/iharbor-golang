package controllers

import (
	"harbor/database"
	"harbor/models"
	"harbor/utils/paginations"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器结构
type UserController struct {
	Controller
}

// UserListJSON 用户列表信息结构
type UserListJSON struct {
	BaseJSON
	Count   uint
	Next    string
	Privous string
	Results []models.UserProfile
}

// NewUserController new controller
func NewUserController() *UserController {
	return &UserController{}
}

// Init 初始化this，子类要重写此方法
func (ctl *UserController) Init() ControllerInterface {

	ctl.this = ctl
	return ctl
}

// GetPermissions return permission
func (ctl UserController) GetPermissions(ctx *gin.Context) []PermissionFunc {

	method := strings.ToUpper(ctx.Request.Method)
	switch method {
	case "GET", "DELETE":
		return []PermissionFunc{IsSuperUser}
	default:
		return []PermissionFunc{}
	}
}

// Get handler for get method
// @Summary 获取用户列表页
// @Description 通过query参数“offset”和“limit”自定义获取用户列表信息
// @Tags user 用户
// @Accept  json
// @Produce  json
// @Param   offset     query    int     true        "The initial index from which to return the results"
// @Param   limit      query    int     true        "Number of results to return per page"
// @Success 200 {object} controllers.UserListJSON
// @Failure 400 {object} controllers.BaseJSON
// @Failure 404 {object} controllers.BaseJSON
// @Security BasicAuth
// @Security ApiKeyAuth
// @Router /api/v1/users/ [get]
func (ctl UserController) Get(ctx *gin.Context) {

	db := database.GetDBDefault()
	paginater := paginations.NewOptimizedLimitOffsetPagination()
	if err := paginater.PrePaginate(ctx); err != nil {
		ctx.JSON(400, BaseJSONResponse(400, err.Error()))
		return
	}
	var users []models.UserProfile
	tableName := models.UserProfile{}.TableName()
	dbQuery := db.Table(tableName).Order("id desc")
	if err := paginater.PaginateDBQuery(&users, dbQuery); err != nil {
		ctx.JSON(500, BaseJSONResponse(500, err.Error()))
		return
	}

	bj := BaseJSONResponse(200, "ok")
	data := UserListJSON{
		BaseJSON: *bj,
		Count:    uint(paginater.GetCount()),
		Results:  users,
		Next:     paginater.GetNextURL(),
		Privous:  paginater.GetPreviousURL(),
	}
	ctx.JSON(200, data)
}

// UserPostForm post form
type UserPostForm struct {
	Username  string `form:"username" json:"username" binding:"max=100,email,required"`
	Password  string `form:"password" json:"password" binding:"min=8,max=128,required"`
	FirstName string `form:"first_name" json:"first_name,omitempty" binding:"max=30"`
	LastName  string `form:"last_name" json:"last_name,omitempty" binding:"max=30"`
	Company   string `form:"company" json:"company,omitempty" binding:"max=255"`
	Telephone string `form:"telephone" json:"telephone,omitempty" binding:"max=11"`
}

func (f *UserPostForm) isValid(ctx *gin.Context) error {

	if err := ctx.ShouldBind(f); err != nil {
		return err
	}
	return f.validate()
}

func (f UserPostForm) validate() error {

	return nil
}

// Post controller
// @Summary 创建用户
// @Description 用户名必须是邮箱
// @Description 密码长度至少8位
// @Tags user 用户
// @Accept  json
// @Produce  json
// @Param   user     body    controllers.UserPostForm     true        "用户名"
// @Success 201 {object} controllers.BaseJSON
// @Failure 400 {object} controllers.BaseJSON
// @Failure 404 {object} controllers.BaseJSON
// @Security BasicAuth
// @Security ApiKeyAuth
// @Router /api/v1/users/ [post]
func (ctl UserController) Post(ctx *gin.Context) {

	form := UserPostForm{}
	if err := form.isValid(ctx); err != nil {
		ctx.JSON(400, BaseJSONResponse(400, err.Error()))
		return
	}

	user := models.UserProfile{}
	db := database.GetDBDefault()
	r := db.Where("username = ?", form.Username).First(&user)
	if r.Error == nil {
		if user.IsActived() {
			ctx.JSON(400, BaseJSONResponse(400, "user already exists"))
			return
		}
	} else if r.RecordNotFound() {
		user.IsActive = false
	}

	user.Username = form.Username
	user.Email = form.Username
	user.FirstName = form.FirstName
	user.LastName = form.LastName
	user.Company = form.Company
	user.Telephone = form.Telephone
	user.DateJoined = models.JSONTimeNow()
	user.SetPassword(form.Password)
	if r := db.Save(&user); r.RowsAffected == 1 && r.Error == nil {
		ctx.JSON(201, BaseJSONResponse(201, "用户创建成功"))
		return
	}
	ctx.JSON(500, BaseJSONResponse(500, "用户创建失败"))
}

// UserRegister 注册用户
// @Description 注册用户
// @Tags Register 注册
// @Accept  json
// @Produce  json
// @Param   some_id     path    string     true        "Some ID"
// @Param   offset     query    int     true        "The initial index from which to return the results"
// @Param   limit      query    int     true        "Number of results to return per page"
// @Success 200 {object} controllers.BaseJSON "ok"
// @Failure 400 {object} controllers.BaseJSON
// @Failure 404 {object} controllers.BaseJSON
// @Router /user/register/ [post]
func UserRegister(ctx *gin.Context) {

	if strings.ToUpper(ctx.Request.Method) == "GET" {
		ctx.HTML(http.StatusOK, "users/register.tmpl", gin.H{})
	}

	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	ctx.JSON(200, gin.H{
		"username": username,
		"password": password,
	})
}

// UserDetailController 用户详情控制器结构
type UserDetailController struct {
	Controller
}

// UserDetailJSON 用户详情信息结构
type UserDetailJSON struct {
	BaseJSON
	User models.UserProfile `json:"user"`
}

// NewUserDetailController new controller
func NewUserDetailController() *UserDetailController {
	return &UserDetailController{}
}

// Init 初始化this，子类要重写此方法
func (ctl *UserDetailController) Init() ControllerInterface {

	ctl.this = ctl
	return ctl
}

// GetPermissions return permission
func (ctl UserDetailController) GetPermissions(ctx *gin.Context) []PermissionFunc {

	method := strings.ToUpper(ctx.Request.Method)
	switch method {
	case "GET", "PATCH":
		return []PermissionFunc{IsAuthenticatedUser}
	case "DELETE":
		return []PermissionFunc{IsStaffSuperUser}
	default:
		return []PermissionFunc{}
	}
}

// Get handler for get method
// @Summary 获取一个用户详细信息
// @Description 通过用户id获取用户详细信息,需要超级用户权限,或id是当前认证用户id
// @Tags user 用户
// @Accept  json
// @Produce  json
// @Param   id     path    int     true        "user id"
// @Success 200 {object} controllers.UserDetailJSON
// @Failure 400 {object} controllers.BaseJSON
// @Failure 404 {object} controllers.BaseJSON
// @Security BasicAuth
// @Security ApiKeyAuth
// @Router /api/v1/users/{id} [get]
func (ctl UserDetailController) Get(ctx *gin.Context) {

	id := ctl.GetParamID(ctx)
	if id == 0 {
		ctx.JSON(400, BaseJSONResponse(400, "invalid param id"))
		return
	}

	u := models.UserProfile{}
	db := database.GetDBDefault()
	if r := db.First(&u, id); r.Error != nil {
		if r.RecordNotFound() {
			ctx.JSON(404, BaseJSONResponse(404, "user not found"))
			return
		}
		ctx.JSON(500, BaseJSONResponse(500, r.Error.Error()))
		return
	}

	// current user is super user or get user is current user
	user := ctl.user
	if IsSuperUser(user) || user.ID == uint(id) {
		bj := BaseJSONResponse(200, "ok")
		data := UserDetailJSON{
			BaseJSON: *bj,
			User:     u,
		}
		ctx.JSON(200, data)
		return
	}
	ctx.JSON(403, BaseJSONResponse(403, "forbidded"))
	return
}

// UserPatchForm patch form
type UserPatchForm struct {
	Password  string `form:"password" json:"password,omitempty" binding:"omitempty,min=8,max=128"`
	FirstName string `form:"first_name" json:"first_name,omitempty" binding:"omitempty,max=30"`
	LastName  string `form:"last_name" json:"last_name,omitempty" binding:"omitempty,max=30"`
	Company   string `form:"company" json:"company,omitempty" binding:"omitempty,max=255"`
	Telephone string `form:"telephone" json:"telephone,omitempty" binding:"omitempty,max=11"`
}

func (f *UserPatchForm) isValid(ctx *gin.Context) error {

	if err := ctx.ShouldBind(f); err != nil {
		return err
	}
	return f.validate()
}

func (f UserPatchForm) validate() error {

	return nil
}

func (f UserPatchForm) updateUser(user *models.UserProfile) error {

	if f.Company != "" {
		user.Company = f.Company
	}
	if f.FirstName != "" {
		user.FirstName = f.FirstName
	}
	if f.LastName != "" {
		user.LastName = f.LastName
	}
	if f.Telephone != "" {
		user.Telephone = f.Telephone
	}
	if f.Password != "" {
		user.SetPassword(f.Password)
	}

	db := database.GetDBDefault()
	if r := db.Save(user); r.Error != nil {
		return r.Error
	}

	return nil
}

// Patch handler for get method
// @Summary 修改用户信息
// @Description 1、超级职员用户拥有所有权限；
// @Description 2、用户拥有修改自己信息的权限；
// @Description 3、超级用户只有修改普通用户信息的权限
// @Tags user 用户
// @Accept  json
// @Produce  json
// @Param   id     path    int     true        "user id"
// @Param   data   body	controllers.UserPatchForm true "change info"
// @Success 200 {object} controllers.UserDetailJSON
// @Failure 400 {object} controllers.BaseJSON
// @Failure 404 {object} controllers.BaseJSON
// @Security BasicAuth
// @Security ApiKeyAuth
// @Router /api/v1/users/{id} [patch]
func (ctl UserDetailController) Patch(ctx *gin.Context) {

	form := UserPatchForm{}
	if err := form.isValid(ctx); err != nil {
		ctx.JSON(400, BaseJSONResponse(400, err.Error()))
		return
	}

	id := ctl.GetParamID(ctx)
	if id == 0 {
		ctx.JSON(400, BaseJSONResponse(400, "invalid param id"))
		return
	}
	u := models.UserProfile{}
	db := database.GetDBDefault()
	if r := db.First(&u, id); r.Error != nil {
		if r.RecordNotFound() {
			ctx.JSON(404, BaseJSONResponse(404, "user not found"))
			return
		}
		ctx.JSON(500, BaseJSONResponse(500, r.Error.Error()))
		return
	}

	user := ctl.user
	// 职员超级用户
	if IsStaffSuperUser(user) ||
		// 超级用户修改普通用户
		(IsSuperUser(user) && u.IsNormalUser()) ||
		// 修改当前用户自己
		(user.ID == u.ID) {
		if err := form.updateUser(&u); err != nil {
			ctx.JSON(500, BaseJSONResponse(500, err.Error()))
			return
		}
		ctx.JSON(200, BaseJSONResponse(200, "ok"))
		return
	}
	ctx.JSON(403, BaseJSONResponse(403, "forbidded"))
	return
}

// Delete handler for get method
// @Summary 删除一个用户
// @Description 通过用户id删除一个用户
// @Tags user 用户
// @Accept  json
// @Produce  json
// @Param   id     path    int     true        "user id"
// @Success 204 {string} string "No content"
// @Failure 400 {object} controllers.BaseJSON
// @Failure 403 {object} controllers.BaseJSON
// @Failure 404 {object} controllers.BaseJSON
// @Security BasicAuth
// @Security ApiKeyAuth
// @Router /api/v1/users/{id} [delete]
func (ctl UserDetailController) Delete(ctx *gin.Context) {

	id := ctl.GetParamID(ctx)
	if id == 0 {
		ctx.JSON(400, BaseJSONResponse(400, "invalid param id"))
		return
	}

	u := models.UserProfile{}
	db := database.GetDBDefault()
	if r := db.First(&u, id); r.Error != nil {
		if r.RecordNotFound() {
			ctx.JSON(404, BaseJSONResponse(404, "user not found"))
			return
		}
		ctx.JSON(500, BaseJSONResponse(500, r.Error.Error()))
		return
	}
	// 改为非激活用户
	if u.IsActived() {
		// u.IsActive = false
		if err := db.Table(u.TableName()).Where("id = ?", u.ID).Update("is_active", "false").Error; err != nil {
			ctx.JSON(500, BaseJSONResponse(500, err.Error()))
			return
		}
	}
	ctx.JSON(204, nil)
}
