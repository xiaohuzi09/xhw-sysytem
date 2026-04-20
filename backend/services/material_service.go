package services

import (
	"errors"
	"strconv"

	"xhw-service/config"
	"xhw-service/models"
)

type MaterialService struct{}

func NewMaterialService() *MaterialService {
	return &MaterialService{}
}

// PageResult 分页结果
type PageResult struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	List     []models.Material `json:"list"`
}

// List 获取素材列表（分页+筛选）
func (s *MaterialService) List(currentUserID uint, currentUserRole string, targetUserID *uint, page, pageSize int, code *int) (*PageResult, error) {
	var materials []models.Material
	var total int64
	db := config.GetDB()

	query := db.Model(&models.Material{})
	if isAdminRole(currentUserRole) {
		if targetUserID != nil && *targetUserID > 0 {
			query = query.Where("user_id = ?", *targetUserID)
		}
	} else {
		if targetUserID != nil && *targetUserID != currentUserID {
			return nil, ErrPermissionDenied
		}
		query = query.Where("user_id = ?", currentUserID)
	}

	if code != nil {
		query = query.Where("code = ?", *code)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&materials).Error; err != nil {
		return nil, err
	}

	return &PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     materials,
	}, nil
}

// Get 获取单个素材
func (s *MaterialService) Get(currentUserID uint, currentUserRole string, id string) (*models.Material, error) {
	materialID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var material models.Material
	db := config.GetDB()
	query := db
	if !isAdminRole(currentUserRole) {
		query = query.Where("user_id = ?", currentUserID)
	}
	if err := query.First(&material, materialID).Error; err != nil {
		return nil, err
	}
	return &material, nil
}

// Create 创建素材（自动生成素材编号）
func (s *MaterialService) Create(currentUserID uint, currentUserRole string, material *models.Material) error {
	db := config.GetDB()

	if !isAdminRole(currentUserRole) || material.UserID == 0 {
		material.UserID = currentUserID
	}

	// 按归属用户获取最大素材编号
	var maxCode int
	db.Model(&models.Material{}).Where("user_id = ?", material.UserID).Select("COALESCE(MAX(code), 0)").Scan(&maxCode)

	// 设置新素材编号为最大编号+1
	material.Code = maxCode + 1

	if err := db.Create(material).Error; err != nil {
		return err
	}
	return nil
}

// Update 更新素材
func (s *MaterialService) Update(currentUserID uint, currentUserRole string, id string, updateData map[string]interface{}) (*models.Material, error) {
	var material models.Material
	db := config.GetDB()

	materialID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	query := db
	if !isAdminRole(currentUserRole) {
		query = query.Where("user_id = ?", currentUserID)
	}
	if err := query.First(&material, materialID).Error; err != nil {
		return nil, err
	}

	// 不允许更新素材编号和用户ID
	removeImmutableFields(updateData, "id", "created_at", "updated_at", "deleted_at", "code", "user_id")

	if err := db.Model(&material).Updates(updateData).Error; err != nil {
		return nil, err
	}
	return &material, nil
}

// Delete 删除素材
func (s *MaterialService) Delete(currentUserID uint, currentUserRole string, id string) error {
	materialID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}

	db := config.GetDB()

	var material models.Material
	query := db
	if !isAdminRole(currentUserRole) {
		query = query.Where("user_id = ?", currentUserID)
	}
	if err := query.First(&material, materialID).Error; err != nil {
		return errors.New("素材不存在或无权限删除")
	}

	if err := db.Delete(&material).Error; err != nil {
		return err
	}
	return nil
}
