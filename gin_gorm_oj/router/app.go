package router

import (
	_ "gin_gorm_oj/docs"
	"gin_gorm_oj/middlewares"
	"gin_gorm_oj/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	//swagger配置
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//路由规则
	r.GET("/ping", service.Ping)
	//公用方法
	//问题相关路由
	r.GET("/problem-list", service.GetProblemList)
	r.GET("/problem-detail", service.GetProblemDetail)

	//用户相关路由
	r.GET("/user-detail", service.GetUserDetail)
	r.POST("/login", service.Login)
	r.POST("/send-code", service.SendCode)
	r.POST("/register", service.Register)

	//排行榜
	r.GET("/rank-list", service.GetRankList)
	//提交记录
	r.GET("/submit-list", service.GetSubmitList)

	//管理员私有方法
	authAdmin := r.Group("/admin", middlewares.AuthAdminCheck())
	//创建问题
	authAdmin.POST("/problem-create", service.ProblemCreate)
	//问题修改
	authAdmin.PUT("/problem-modify", service.ProblemModify)
	//分类列表
	authAdmin.GET("/category-list", service.GetCategoryList)
	//分类创建
	authAdmin.POST("/category-create", service.CategoryCreate)
	//分类的修改
	authAdmin.PUT("/category-modify", service.CategoryModify)
	//分类的删除
	authAdmin.DELETE("/category-delete", service.CategoryDelete)

	//用户私有方法
	authUser := r.Group("/user", middlewares.AuthUserCheck())
	//代码提交
	authUser.POST("/submit", service.Submit)
	r.Run(":8080")

	return r
}
