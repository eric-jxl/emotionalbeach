package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate 在创建记录前设置 CreatedAt
func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreatedAt = time.Now()
	return
}

// BeforeUpdate 在更新记录前确保 CreatedAt 不变
func (m *Model) BeforeUpdate(tx *gorm.DB) (err error) {
	// 无需修改 CreatedAt
	return
}

type UserBasic struct {
	Model
	Name          string
	PassWord      string
	Avatar        string
	Gender        string `gorm:"column:gender;default:male;type:varchar(6) comment:'male表示男，female表示女'"` // gorm为数据库字段约束
	Phone         string `gorm:"unique;not null" valid:"matches(^1[3-9]{1}\\d{9}$)"`                           //valid为条件约束
	Email         string `valid:"email"`
	Identity      string
	ClientIp      string `valid:"ipv4"`
	ClientPort    string
	Salt          string     //盐值
	LoginTime     *time.Time `gorm:"column:login_time"`
	HeartBeatTime *time.Time `gorm:"column:heart_beat_time"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time"`
	IsLoginOut    bool
	DeviceInfo    string //登录设备
}

// UserTableName 指定表的名称
func (table *UserBasic) UserTableName() string {
	return "user_basic"
}
