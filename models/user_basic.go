package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID uint `gorm:"primaryKey"`
}

type DatetimeModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserBasic struct {
	BaseModel
	Name          string `gorm:"column:name;type:varchar(255);comment:名称"`
	Password      string `gorm:"column:password"`
	Avatar        string `gorm:"comment:头像"`
	Gender        string `gorm:"column:gender;default:male;type:varchar(6); comment:male表示男，female表示女"`                                                                  // gorm为数据库字段约束
	Phone         string `gorm:"type:varchar(64);index; unique;comment:手机号;not null" valid:"matches(^1[3-9]{1}\\d{9}$)" validate:"required,numeric,len=11,startswith=1"` //valid为条件约束
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
	DatetimeModel
}

// UserTableName 指定表的名称
func (table *UserBasic) UserTableName() string {
	return "user_basic"
}

type Relation struct {
	BaseModel
	OwnerId  uint   //谁的关系信息
	TargetID uint   //对应的谁
	Type     int    //关系类型： 1表示好友关系 2表示群关系
	Desc     string //描述
	DatetimeModel
}

func (r *Relation) RelTableName() string {
	return "relation"
}
