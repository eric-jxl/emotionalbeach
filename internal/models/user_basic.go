package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID uint `gorm:"primaryKey;comment:主键" json:"id"`
}

type DatetimeModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime:milli;comment:创建时间" json:"created_time"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:milli;comment:修改时间" json:"updated_time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserBasic struct {
	BaseModel
	Name          string     `gorm:"column:name;type:varchar(255);comment:名称" json:"name"`
	Password      string     `gorm:"column:password;comment:密码" json:"password"`
	Role          string     `gorm:"column:role;default:user;type:varchar(128);size:128;comment:superadmin表示超级管理员, admin表示管理员,user表示普通用户" json:"role"`
	Avatar        string     `gorm:"comment:头像" json:"avatar"`
	Gender        string     `gorm:"column:gender;default:male;type:varchar(6); comment:male表示男，female表示女" json:"gender"`              // gorm为数据库字段约束
	Phone         string     `gorm:"type:varchar(64);index; unique;comment:手机号;not null" valid:"matches(^1[3-9]{1}\\d{9}$)" json:"phone"` //valid为条件约束
	Email         string     `gorm:"size:255;comment:Email" json:"email"`
	Identity      string     `gorm:"column:identity;size:128;comment:密钥" json:"identity"`
	ClientIp      string     `gorm:"comment:ip地址"  json:"client_ip"`
	Salt          string     `gorm:"column:salt;comment:密码加盐" json:"salt"`
	LoginTime     *time.Time `gorm:"column:login_time;comment:登陆时间" json:"login_time"`
	HeartBeatTime *time.Time `gorm:"column:heart_beat_time;comment:心跳时间" json:"heart_beat_time"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time;comment:登出时间" json:"login_out_time"`
	IsLogOut      bool       `gorm:"column:is_logout; comment:是否登出" json:"is_logout"`
	DeviceInfo    string     `gorm:"column:device_info;size:64;comment:设备信息" json:"device_info"`
	DatetimeModel
}

// TableName UserTableName 指定表的名称
func (table *UserBasic) TableName() string {
	return "user_basic"
}

type Relation struct {
	BaseModel
	OwnerId  uint   `json:"owner_id"`  //谁的关系信息
	TargetID uint   `json:"target_id"` //对应的谁
	Type     int    `json:"type"`      //关系类型： 1表示好友关系 2表示群关系
	Desc     string `json:"desc"`      //描述
	DatetimeModel
}

func (r *Relation) TableName() string {
	return "relation"
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" `
	Password string `json:"password" binding:"required"`
}
