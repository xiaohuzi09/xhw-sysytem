package models

// Template 模版模型
type Template struct {
	Model
	UserID  uint    `gorm:"not null;index" json:"user_id"`      // 用户ID
	Name    string  `gorm:"size:100;not null" json:"name"`      // 模版名称
	URL     string  `gorm:"size:500;not null" json:"url"`       // 模版URL
	Width   int     `gorm:"not null;default:0" json:"width"`    // 素材宽度
	Height  int     `gorm:"not null;default:0" json:"height"`   // 素材高度
	OffsetX float64 `gorm:"not null;default:0" json:"offset_x"` // X轴偏移
	OffsetY float64 `gorm:"not null;default:0" json:"offset_y"` // Y轴偏移
	Scale    float64 `gorm:"not null;default:1" json:"scale"`      // 缩放比例
	Rotation float64 `gorm:"not null;default:0" json:"rotation"`   // 旋转角度（度数，顺时针）
}

func (Template) TableName() string {
	return "templates"
}
