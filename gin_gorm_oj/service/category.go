package service

import (
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// GetCategoryList
// @Summary 分类列表
// @Tags 管理员私有方法
// @Param page query int false "page"
// @Param size query int false "size"
// @Param authorization header string true "authorization"
// @Success 200 {string} string "ok"
// @Router /admin/category-list [get]
func GetCategoryList(c *gin.Context) {
	//分页查询功能，默认页起点，默认页大小
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Get ProblemBasic Page Parse Error", err)
	}
	//其实页位置
	page = (page - 1) * size
	keyword := c.Query("keyword")

	//拿到总的页数
	var count int64

	categoryList := make([]*models.CategoriesBasic, 0)
	err = models.DB.Model(new(models.CategoriesBasic)).Where("name like ?", "%"+keyword+"%").
		Count(&count).Limit(size).Offset(page).Find(&categoryList).Error
	if err != nil {
		log.Println("GetCategoryList Error", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"count": count,
			"list":  categoryList,
		},
	})

}

// CategoryCreate
// @Summary 创建分类
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param name formData string true "name"
// @Param parentId formData string false "parentId"
// @Success 200 {string} string "ok"
// @Router /admin/category-create [post]
func CategoryCreate(c *gin.Context) {
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parentId"))
	categoryBasic := &models.CategoriesBasic{
		Identity: helper.GetUUID(),
		Name:     name,
		ParentId: parentId,
	}

	err := models.DB.Create(categoryBasic).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "创建分类失败" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建分类成功",
	})
}

// CategoryModify
// @Summary 修改分类
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param identity formData string true "identity"
// @Param name formData string true "name"
// @Param parentId formData string false "parentId"
// @Success 200 {string} string "ok"
// @Router /admin/category-modify [put]
func CategoryModify(c *gin.Context) {
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parentId"))
	identity := c.PostForm("identity")
	if name == "" || identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	categoryBasic := &models.CategoriesBasic{
		Identity: identity,
		Name:     name,
		ParentId: parentId,
	}

	err := models.DB.Model(new(models.CategoriesBasic)).Where("identity = ?", identity).Updates(categoryBasic).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "修改分类失败" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "修改分类成功",
	})

}

// CategoryDelete
// @Summary 删除分类
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param identity query string true "identity"
// @Success 200 {string} string "ok"
// @Router /admin/category-delete [delete]
func CategoryDelete(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	var cnt int64
	err := models.DB.Model(new(models.ProblemCategory)).Where("category_id = (SELECT id FROM categories_basic WHERE identity = ? LIMIT 1)", identity).Count(&cnt).Error
	if err != nil {
		log.Println("Get Problem Error", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类关联的问题失败",
		})
		return
	}
	if cnt > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该分类中存在问题，不可删除",
		})
		return
	}
	err = models.DB.Where("identity = ?", identity).Delete(new(models.CategoriesBasic)).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "删除失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})

}
