package models

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

// User 用户模型
type User struct {
	Model
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`
	Nickname string `gorm:"size:50" json:"nickname"`
	Avatar   string `gorm:"size:255" json:"avatar"`
	Role     string `gorm:"size:20;not null;default:user" json:"role"` // admin: 管理员, user: 普通用户
	Status   int    `gorm:"default:1" json:"status"`                   // 1: 正常, 0: 禁用
}

func (User) TableName() string {
	return "users"
}
