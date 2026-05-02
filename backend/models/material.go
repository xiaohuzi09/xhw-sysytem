package models

// Material 素材模型
type Material struct {
	Model
	UserID  uint   `gorm:"not null;index;uniqueIndex:idx_material_user_code" json:"user_id"` // 用户ID
	URL     string `gorm:"size:500;not null" json:"url"`                                     // 素材URL
	Code    int    `gorm:"not null;uniqueIndex:idx_material_user_code" json:"code"`          // 素材编号
	TitleCN string `gorm:"size:500" json:"title_cn"`                                         // 中文标题
	TitleEN string `gorm:"size:500" json:"title_en"`                                         // 英文标题
}

func (Material) TableName() string {
	return "materials"
}
