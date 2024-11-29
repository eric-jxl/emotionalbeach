package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserBasic struct {
	Model
	Name          string `gorm:"column:name;type:varchar(255);comment:名称"`
	PassWord      string
	Avatar        string
	Gender        string `gorm:"column:gender;default:male;type:varchar(6); comment:male表示男，female表示女"` // gorm为数据库字段约束
	Phone         string `gorm:"type:varchar(64);index; unique" valid:"matches(^1[3-9]{1}\\d{9}$)"`     //valid为条件约束
	Email         string `valid:"email"`
	Identity      string `gorm:"comment: 密钥"`
	ClientIp      string `valid:"ipv4"`
	ClientPort    string
	Salt          string     //盐值
	LoginTime     *time.Time `gorm:"column:login_time"`
	HeartBeatTime *time.Time `gorm:"column:heart_beat_time"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time"`
	IsLogOut      bool       `gorm:"column:is_logout; comment:是否登出"`
	DeviceInfo    string     `gorm:"column:device_info; comment:设备信息"`
}

// UserTableName 指定表的名称
func (table *UserBasic) UserTableName() string {
	return "user_basic"
}
