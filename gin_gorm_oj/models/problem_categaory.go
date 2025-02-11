package models

import "gorm.io/gorm"

type ProblemCategory struct {
	gorm.Model
	ProblemId     uint             `gorm:"column:problem_id;type:int(11);" json:"problem_id"`           // 问题的ID
	CategoryId    uint             `gorm:"column:category_id;type:int(11);" json:"category_id"`         // 分类的ID
	CategoryBasic *CategoriesBasic `gorm:"foreignKey:id;references:category_id;" json:"category_basic"` // 关联分类的基础信息表
}

func (table *ProblemCategory) TableName() string {
	return "problem_category"
}
