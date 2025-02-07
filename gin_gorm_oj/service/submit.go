package service

import (
	"bytes"
	"errors"
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// GetSubmitList
// @Tags 公共方法
// @Summary 提交列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param status query int false "status"
// @Param problem_identity query string false "problem identity"
// @Param user_identity query string false "user identity"
// @Success 200 {string} string "ok"
// @Router /submit-list [get]
func GetSubmitList(c *gin.Context) {
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
	list := make([]models.SubmitsBasic, 0)
	//拿到传入的参数
	problemIdentity := c.Query("problem_identity")
	userIdentity := c.Query("user_identity")
	status, _ := strconv.Atoi(c.Query("status"))
	tx := models.GetSubmitList(problemIdentity, userIdentity, status)

	err = tx.Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Submit list error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"count": count,
		"data": map[string]interface{}{
			"count": count,
			"list":  list,
		},
	})
}

// Submit
// @Tags 用户私有方法
// @Summary 代码提交
// @Param authorization header string true "authorization"
// @Param problem_identity query string false "problem identity"
// @Param code body string true "code"
// @Success 200 {string} string "ok"
// @Router /user/submit [post]
func Submit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	code, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Read code error: " + err.Error(),
		})
		return
	}
	//代码保存
	path, err := helper.CodeSave(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Save code error: " + err.Error(),
		})
		return
	}

	u, _ := c.Get("user")
	userClam := u.(*helper.UserClaims)
	sb := &models.SubmitsBasic{
		Identity:        helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClam.Identity,
		Path:            path,
	}
	//代码判断
	pb := new(models.ProblemBasic)
	err = models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Problem error: " + err.Error(),
		})
		return
	}
	// 答案错误
	WA := make(chan int)
	// 超内存
	OOM := make(chan int)
	// 编译错误
	CE := make(chan int)
	// 测试案例通过的个数
	passCount := 0
	// 提示信息
	msg := "运行通过"
	var lock sync.Mutex
	for _, testCase := range pb.TestCases {
		testCase := testCase
		go func() {
			//执行测试
			// go run code-user/main.go
			cmd := exec.Command("go", "run", path)

			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				log.Fatalln(err)
			}
			_, err = io.WriteString(stdinPipe, testCase.Input)
			if err != nil {
				log.Println(err)
				return
			}
			// 记录内存
			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)
			//根据测试的输入案例运行，拿到输出结果和标准的输出结果进行比对
			if err = cmd.Run(); err != nil {
				log.Println(err, stderr.String())
				//编译出错
				if err.Error() == "exit status 1" {
					msg = stderr.String()
					CE <- 1
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)
			// 答案错误
			if testCase.Output != out.String() {
				msg = "答案错误, 预期结果：" + testCase.Output + "运行结果" + out.String()
				WA <- 1
				return
			}
			// 运行超内存
			if em.Alloc/1024-bm.Alloc/1024 > uint64(pb.MaxMem) {
				msg = "运行超内存"
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			lock.Unlock()
		}()
	}
	select {
	case <-WA:
		sb.Status = 2
	case <-OOM:
		sb.Status = 4
	case <-CE:
		sb.Status = 5
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		if passCount == len(pb.TestCases) {
			sb.Status = 1
		} else {
			sb.Status = 3
			msg = "超时运行"
		}

	}

	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		//保存提交数据
		err = tx.Create(sb).Error
		if err != nil {
			return errors.New("Submit Create Error：" + err.Error())
		}
		//更新用户信息
		m := make(map[string]interface{})
		m["submit_num"] = gorm.Expr("submit_num + ?", 1)
		if sb.Status == 1 {
			m["pass_num"] = gorm.Expr("pass_num + ?", 1)
		}
		err = tx.Model(new(models.UserBasic)).Where("identity = ?", userClam.Identity).Updates(m).Error
		if err != nil {
			return errors.New("UserModel Modify Error：" + err.Error())
		}
		//更新问题列表
		err = tx.Model(new(models.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error
		if err != nil {
			return errors.New("Problem Modify Error：" + err.Error())
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "Submit Error:" + err.Error(),
		})
		return
	}

	//返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"status": sb.Status,
			"msg":    msg,
		},
	})
}
