package models

import "gorm.io/gorm"

type SubmitsBasic struct {
	gorm.Model
	Identity        string        `gorm:"column:identity;type:varchar(36);" json:"identity"`                     // 唯一标识
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`     // 问题表的唯一标识
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:identity;references:problem_identity;" json:"problem_basic"` // 关联问题基础表
	UserIdentity    string        `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`           // 用户表的唯一标识
	UserBasic       *UserBasic    `gorm:"foreignKey:identity;references:user_identity;" json:"user_basic"`       // 关联用户基础表
	Path            string        `gorm:"column:path;type:varchar(255);" json:"path"`                            // 代码存放路径
	Status          int           `gorm:"column:status;type:tinyint(1);" json:"status"`                          // 【-1-待判断，1-答案正确，2-答案错误，3-运行超时，4-运行超内存， 5-编译错误，6-非法代码】
}

func (table *SubmitsBasic) TableName() string {
	return "submits_basic"
}

func GetSubmitList(problemIdentity string, userIdentity string, status int) *gorm.DB {
	tx := DB.Model(new(SubmitsBasic)).Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB {
		return db.Omit("content")
	}).Preload("UserBasic")
	if problemIdentity != "" {
		tx.Where("problem_identity = ?", problemIdentity)
	}
	if userIdentity != "" {
		tx.Where("user_identity = ?", userIdentity)
	}
	if status != 0 {
		tx.Where("status = ?", status)
	}
	return tx
}
