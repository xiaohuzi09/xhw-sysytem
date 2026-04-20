package services

import (
	"errors"
	"strconv"

	"xhw-service/config"
	"xhw-service/models"
)

type TemplateService struct{}

func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// List 获取模版列表
func (s *TemplateService) List(currentUserID uint, currentUserRole string, targetUserID *uint) ([]models.Template, error) {
	var templates []models.Template
	db := config.GetDB()

	query := db.Model(&models.Template{})
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

	if err := query.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// Get 获取单个模版
func (s *TemplateService) Get(id string, currentUserID uint, currentUserRole string) (*models.Template, error) {
	templateID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var template models.Template
	db := config.GetDB()
	query := db
	if !isAdminRole(currentUserRole) {
		query = query.Where("user_id = ?", currentUserID)
	}
	if err := query.First(&template, templateID).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// Create 创建模版
func (s *TemplateService) Create(template *models.Template, currentUserID uint, currentUserRole string) error {
	if !isAdminRole(currentUserRole) || template.UserID == 0 {
		template.UserID = currentUserID
	}

	db := config.GetDB()
	if err := db.Create(template).Error; err != nil {
		return err
	}
	return nil
}

// Update 更新模版
func (s *TemplateService) Update(id string, currentUserID uint, currentUserRole string, updateData map[string]interface{}) (*models.Template, error) {
	var template models.Template
	db := config.GetDB()

	templateID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	query := db
	if !isAdminRole(currentUserRole) {
		query = query.Where("user_id = ?", currentUserID)
	}
	if err := query.First(&template, templateID).Error; err != nil {
		return nil, err
	}

	removeImmutableFields(updateData, "id", "created_at", "updated_at", "deleted_at", "user_id")

	if err := db.Model(&template).Updates(updateData).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// Delete 删除模版
func (s *TemplateService) Delete(id string, currentUserID uint, currentUserRole string) error {
	templateID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}

	db := config.GetDB()

	var template models.Template
	query := db
	if !isAdminRole(currentUserRole) {
		query = query.Where("user_id = ?", currentUserID)
	}
	if err := query.First(&template, templateID).Error; err != nil {
		return errors.New("模版不存在或无权限删除")
	}

	if err := db.Delete(&template).Error; err != nil {
		return err
	}
	return nil
}
