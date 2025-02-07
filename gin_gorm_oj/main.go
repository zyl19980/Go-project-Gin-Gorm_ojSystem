package main

import "gin_gorm_oj/router"

func main() {
	//err := models.DB.AutoMigrate(&models.ProblemBasic{})
	//if err != nil {
	//	fmt.Println("建表失败")
	//}
	//fmt.Println("建表成功")

	r := router.Router()
	r.Run()

}
