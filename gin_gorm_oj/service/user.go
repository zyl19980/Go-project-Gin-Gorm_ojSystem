package service

import (
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

// GetUserDetail
// @Summary 用户详情
// @Tags 公共方法
// @Param identity query string false "user identity"
// @Success 200 {string} string "ok"
// @Router /user-detail [get]
func GetUserDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户的唯一标识不能为空",
		})
		return
	}
	data := new(*models.UserBasic)
	err := models.DB.Debug().Where("identity = ?", identity).Find(&data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GetUserDetail by identity Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// Login
// @Summary 用户登录
// @Tags 公共方法
// @Param username formData string false "username"
// @Param password formData string false "password"
// @Success 200 {string} string "ok"
// @Router /login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "必填信息为空",
		})
		return
	}
	//将password用md5加密
	password = helper.GetMd5(password)
	print("=======" + password + "=======")
	data := new(models.UserBasic)
	err := models.DB.Where("name = ? AND password = ?", username, password).First(&data).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "用户名或密码错误",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Get UserBasic Error" + err.Error(),
			})
		}
		return
	}

	token, err := helper.GenerateToken(data.Identity, data.Name, data.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GenerateToken Error" + err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// SendCode
// @Summary 发送验证码
// @Tags 公共方法
// @Param email formData string false "email"
// @Success 200 {string} string "ok"
// @Router /send-code [post]
func SendCode(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	// 生成验证码，并将其存在Redis中
	code := helper.GetRand()
	models.RDB.Set(c, email, code, time.Second*60)
	//发送邮件
	err := helper.SendCode(email, code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Send Code Error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "验证码发送成功",
	})
}

// Register
// @Summary 用户注册
// @Tags 公共方法
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Param code formData string true "code"
// @Param email formData string true "email"
// @Param phone formData string false "phone"
// @Success 200 {string} string "ok"
// @Router /register [post]
func Register(c *gin.Context) {
	//1.拿到Post数据
	username := c.PostForm("username")
	password := c.PostForm("password")
	userCode := c.PostForm("code")
	email := c.PostForm("email")
	phone := c.PostForm("phone")
	if username == "" || email == "" || password == "" || userCode == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	//2.验证验证码
	sysCode, err := models.RDB.Get(c, email).Result()
	if err != nil {
		log.Printf("Get Code Error: %v : \n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "验证码过期，请重新发送验证码：" + err.Error(),
		})
		return
	}
	if sysCode != userCode {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "验证码不正确：" + err.Error(),
		})
		return
	}

	//判断邮箱是否存在
	var cnt int64
	err = models.DB.Where("mail = ?", email).Model(new(models.UserBasic)).Count(&cnt).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get User Error" + err.Error(),
		})
		return
	}
	if cnt > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该邮箱已被注册",
		})
		return
	}

	//3.数据的插入
	data := &models.UserBasic{
		Model:    gorm.Model{},
		Name:     username,
		Password: helper.GetMd5(password),
		Phone:    phone,
		Mail:     email,
		Identity: helper.GetUUID(),
	}
	err = models.DB.Create(data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Create User Error:" + err.Error(),
		})
		return
	}

	//4.生成token
	token, err := helper.GenerateToken(data.Identity, data.Name, data.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Generate Token Error:" + err.Error(),
		})
		return
	}

	//5.返回正确结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})

}

// GetRankList
// @Summary 用户排行榜
// @Tags 公共方法
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} string "ok"
// @Router /rank-list [get]
func GetRankList(c *gin.Context) {
	//分页查询功能，默认页起点，默认页大小
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Get ProblemBasic Page Parse Error", err)
	}
	//其实页位置
	page = (page - 1) * size
	//拿到总的页数
	var count int64
	list := make([]*models.UserBasic, 0)
	err = models.DB.Model(new(models.UserBasic)).Count(&count).Order("pass_num DESC, submit_num ASC").
		Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Rank List Error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"count": count,
			"list":  list,
		},
	})

}
