package middlewares

import (
	"gin_gorm_oj/helper"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthAdminCheck
// 验证用户是否是admin的中间件
func AuthAdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.AnalyseToken(auth)
		//用户token不存在报错
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized 1 : token error",
			})
			return
		}
		//用户权限不正确报错
		if userClaim.IsAdmin != 1 {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized 2 : user is not admin",
			})
			return
		}
		c.Next()
	}
}
