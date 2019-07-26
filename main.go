package main

import (
	"harbor/database"
	"harbor/middlewares"
	"harbor/models"
	"harbor/routes"
	"harbor/utils/renders"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "harbor/docs" // docs is generated by Swag CLI, you have to import it
)

// @title EVHarbor API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://obs.casearth.cn/

// @contact.name API Support
// @contact.url http://obs.casearth.cn/
// @contact.email helpdesk@cnic.cn

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @host 127.0.0.1:9000
// @BasePath /
func main() {

	db := database.GetDBDefault()
	defer db.Close()
	db.AutoMigrate(&models.UserProfile{},
		&models.HarborObject{},
		&models.Bucket{},
	)

	app := gin.Default()
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	app.Use(middlewares.BasicAuth(db, &models.UserProfile{ID: 20}))
	routes.Urls(app)
	app.GET("/", index)
	app.Static("/static", "./static") // 设置静态资源
	app.LoadHTMLGlob("views/**/*")    //加载模板

	app.Run(":9000")
}

func index(ctx *gin.Context) {

	renders.HTMLRenderByPongo2(ctx, 200, "views/index/index.html", nil)
}
