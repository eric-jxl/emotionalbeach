package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID uint `gorm:"primaryKey;comment:主键"`
}

type DatetimeModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:milli;comment:修改时间"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserBasic struct {
	BaseModel
	Name          string     `gorm:"column:name;type:varchar(255);comment:名称"`
	Password      string     `gorm:"column:password;comment:密码"`
	Role          string     `gorm:"column:role;default:user;type:varchar(128);size:128;comment:superadmin表示超级管理员, admin表示管理员,user表示普通用户"`
	Avatar        string     `gorm:"comment:头像"`
	Gender        string     `gorm:"column:gender;default:male;type:varchar(6); comment:male表示男，female表示女"`                  // gorm为数据库字段约束
	Phone         string     `gorm:"type:varchar(64);index; unique;comment:手机号;not null" valid:"matches(^1[3-9]{1}\\d{9}$)"` //valid为条件约束
	Email         string     `gorm:"size:255;comment:Email"`
	Identity      string     `gorm:"column:identity;size:128;comment:密钥"`
	ClientIp      string     `gorm:"comment:ip地址" valid:"ipv4"`
	Salt          string     `gorm:"column:salt;comment:密码加盐"`
	LoginTime     *time.Time `gorm:"column:login_time;comment:登陆时间"`
	HeartBeatTime *time.Time `gorm:"column:heart_beat_time;comment:心跳时间"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time;comment:登出时间"`
	IsLogOut      bool       `gorm:"column:is_logout; comment:是否登出"`
	DeviceInfo    string     `gorm:"column:device_info;size:64;comment:设备信息"`
	DatetimeModel
}

// TableName UserTableName 指定表的名称
func (table *UserBasic) TableName() string {
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

func (r *Relation) TableName() string {
	return "relation"
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
