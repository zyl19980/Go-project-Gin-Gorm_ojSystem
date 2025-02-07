package service

import (
	"encoding/json"
	"errors"
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

// GetProblemList
// @Summary 问题列表
// @Tags 公共方法
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Param category_identity query string false "category_identity"
// @Success 200 {string} string "ok"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
	//分页查询功能，默认页起点，默认页大小
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Get ProblemBasic Page Parse Error", err)
	}
	//其实页位置
	page = (page - 1) * size
	keyword := c.Query("keyword")
	categoryIdentity := c.Query("category_identity")
	//拿到总的页数
	var count int64
	//问题列表
	problemList := make([]*models.ProblemBasic, 0)
	//拿到查询后的db
	tx := models.GetProblemList(keyword, categoryIdentity)
	//分页查询
	err = tx.Count(&count).Omit("content").Offset(page).Limit(size).Find(&problemList).Error
	if err != nil {
		log.Println("Get ProblemBasic List Error")
		return
	}
	//设置响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"problemList": problemList,
			"count":       count,
		},
	})
}

// GetProblemDetail
// @Summary 问题详情
// @Tags 公共方法
// @Param identity query string false "problem identity"
// @Success 200 {string} string "ok"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "问题的唯一标识不能为空",
		})
		return
	}
	problemBasic := new(models.ProblemBasic)
	err := models.DB.Where("identity = ?", identity).
		Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		First(&problemBasic).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "问题不存在",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Get ProblemDetail Error :" + err.Error(),
			})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": problemBasic,
	})
}

// ProblemCreate
// @Summary 创建问题
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData string true "max_runtime"
// @Param max_memory formData string true "max_memory"
// @Param category_ids formData array false "category_ids"
// @Param test_cases formData array true "test_cases"
// @Success 200 {string} string "ok"
// @Router /admin/problem-create [post]
func ProblemCreate(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(c.PostForm("max_memory"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")

	if title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}
	//生成id
	identity := helper.GetUUID()
	//封装数据
	data := models.ProblemBasic{
		Title:      title,
		Content:    content,
		MaxRuntime: maxRuntime,
		MaxMem:     maxMemory,
		Identity:   identity,
	}
	//处理相关分类
	problemCategories := make([]*models.ProblemCategory, 0)
	for _, id := range categoryIds {
		categoryId, _ := strconv.Atoi(id)
		problemCategories = append(problemCategories, &models.ProblemCategory{
			ProblemId:  data.ID,
			CategoryId: uint(categoryId),
		})
	}
	data.ProblemCategories = problemCategories

	//处理相关测试用例
	testCasesBasics := make([]*models.TestCase, 0)
	for _, testCase := range testCases {
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)
		_, inputOK := caseMap["input"]
		_, outputOK := caseMap["output"]
		if err != nil || !inputOK || !outputOK {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}

		testCaseBasic := &models.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		}
		testCasesBasics = append(testCasesBasics, testCaseBasic)
	}
	data.TestCases = testCasesBasics

	//插入数据
	err := models.DB.Create(&data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Create Error : " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "插入数据成功",
	})
}

// ProblemModify
// @Summary 修改问题
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param title formData string true "title"
// @Param identity formData string true "identity"
// @Param content formData string true "content"
// @Param max_runtime formData string true "max_runtime"
// @Param max_memory formData string true "max_memory"
// @Param category_ids formData []string true "category_ids" collectionFormat(multi)
// @Param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} string "ok"
// @Router /admin/problem-modify [put]
func ProblemModify(c *gin.Context) {
	identity := c.PostForm("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(c.PostForm("max_memory"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")

	log.Println("============>identity:", identity)
	log.Println("============>title:", title)
	log.Println("============>content:", content)
	log.Println("============>maxRuntime:", maxRuntime)
	log.Println("============>maxMemory", maxMemory)
	log.Println("============>categoryIds", categoryIds)
	log.Println("============>testCases", testCases)

	if identity == "" || title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		//1.问题基础信息的保存
		problemBasic := &models.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxRuntime: maxRuntime,
			MaxMem:     maxMemory,
		}

		err := tx.Debug().Where("identity = ?", identity).Updates(problemBasic).Error
		if err != nil {
			log.Println("ProblemModify Error===========> 问题基础信息更新失败")
			return err
		}
		//拿到问题详情
		err = tx.Debug().Where("identity = ?", identity).Find(problemBasic).Error
		if err != nil {
			log.Println("ProblemModify Error===========> 问题查询失败")
			return err
		}
		//2.关联问题分类的更新
		//2.1删除已存在的关联关系
		err = tx.Debug().Where("problem_id = ?", problemBasic.ID).Delete(new(models.ProblemCategory)).Error
		if err != nil {
			log.Println("ProblemModify Error===========> 删除已存在分类关联失败")
			return err
		}
		//2.2增加新的关联关系
		pcs := make([]*models.ProblemCategory, 0)
		for _, id := range categoryIds {
			intId, _ := strconv.Atoi(id)
			pcs = append(pcs, &models.ProblemCategory{
				ProblemId:  problemBasic.ID,
				CategoryId: uint(intId),
			})
		}
		err = tx.Debug().Create(&pcs).Error
		if err != nil {
			log.Println("ProblemModify Error===========> 插入分类关联失败")
			return err
		}
		//3.关联测试案例的更新
		// 3.1删除已存在的关联关系
		err = tx.Debug().Where("problem_identity = ?", problemBasic.Identity).Delete(new(models.TestCase)).Error
		if err != nil {
			log.Println("ProblemModify Error===========> 删除已存在测试案例失败")
			return err
		}

		// 3.2增加新的关联关系
		tcs := make([]*models.TestCase, 0)
		for _, testCase := range testCases {
			caseMap := make(map[string]string)
			err = json.Unmarshal([]byte(testCase), &caseMap)
			if err != nil {
				log.Println("ProblemModify Error===========> 测试案例Json格式转化失败")
				return err
			}
			log.Println("==========>caseMap:", caseMap)
			_, inputOK := caseMap["input"]
			_, outputOK := caseMap["output"]
			if !inputOK || !outputOK {
				return errors.New("测试案例输入输出错误")
			}
			tcs = append(tcs, &models.TestCase{
				Identity:        helper.GetUUID(),
				ProblemIdentity: identity,
				Input:           caseMap["input"],
				Output:          caseMap["output"],
			})
		}
		log.Println("==========>tcs:", tcs)
		err = tx.Create(&tcs).Error
		if err != nil {
			log.Println("ProblemModify Error===========> 插入测试案例失败")
			return err
		}
		return nil
	}); err != nil {
		log.Println("Modify Problem Error：" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "修改问题失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "修改问题成功",
	})
}
