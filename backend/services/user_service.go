package services

import (
	"errors"
	"strconv"

	"xhw-service/config"
	"xhw-service/models"
	"xhw-service/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// List 获取用户列表
func (s *UserService) List(currentUserRole string) ([]models.User, error) {
	if !isAdminRole(currentUserRole) {
		return nil, ErrPermissionDenied
	}

	var users []models.User
	db := config.GetDB()
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Get 获取单个用户
func (s *UserService) Get(id string, currentUserID uint, currentUserRole string) (*models.User, error) {
	targetUserID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	if !canAccessUser(currentUserID, uint(targetUserID), currentUserRole) {
		return nil, ErrPermissionDenied
	}

	var user models.User
	db := config.GetDB()
	if err := db.First(&user, targetUserID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (s *UserService) Create(user *models.User, currentUserRole string) error {
	if !isAdminRole(currentUserRole) {
		return ErrPermissionDenied
	}

	db := config.GetDB()

	if err := s.ensureUniqueUser(db, 0, user.Username, user.Email); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.Password = string(hashedPassword)
	user.Role = resolveRoleByUsername(user.Username)

	if err := db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// Update 更新用户
func (s *UserService) Update(id string, currentUserID uint, currentUserRole string, updateData map[string]interface{}) (*models.User, error) {
	targetUserID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	if !canAccessUser(currentUserID, uint(targetUserID), currentUserRole) {
		return nil, ErrPermissionDenied
	}

	var user models.User
	db := config.GetDB()
	if err := db.First(&user, targetUserID).Error; err != nil {
		return nil, err
	}

	removeImmutableFields(updateData, "id", "created_at", "updated_at", "deleted_at", "role")

	nextUsername := user.Username
	if username, ok := updateData["username"].(string); ok && username != "" {
		nextUsername = username
		if !isAdminRole(currentUserRole) && config.IsAdminUsername(nextUsername) {
			return nil, ErrPermissionDenied
		}
	}

	nextEmail := user.Email
	if email, ok := updateData["email"].(string); ok && email != "" {
		nextEmail = email
	}

	if err := s.ensureUniqueUser(db, user.ID, nextUsername, nextEmail); err != nil {
		return nil, err
	}

	if password, ok := updateData["password"].(string); ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		updateData["password"] = string(hashedPassword)
	}
	updateData["role"] = resolveRoleByUsername(nextUsername)

	if err := db.Model(&user).Updates(updateData).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete 删除用户
func (s *UserService) Delete(id string, currentUserRole string) error {
	if !isAdminRole(currentUserRole) {
		return ErrPermissionDenied
	}

	db := config.GetDB()
	if err := db.Delete(&models.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

// GetByUsername 按用户名查询用户
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	var user models.User
	db := config.GetDB()
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Register 用户注册
func (s *UserService) Register(user *models.User) error {
	db := config.GetDB()

	if err := s.ensureUniqueUser(db, 0, user.Username, user.Email); err != nil {
		return err
	}

	// 使用 bcrypt 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.Password = string(hashedPassword)
	user.Role = resolveRoleByUsername(user.Username)

	// 创建用户
	if err := db.Create(user).Error; err != nil {
		return err
	}

	return nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*models.User, string, error) {
	// 查找用户
	user, err := s.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户不存在")
		}
		return nil, "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("密码错误")
	}

	if user.Status != 1 {
		return nil, "", errors.New("用户已被禁用")
	}

	// 生成 JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return user, token, nil
}

func (s *UserService) ensureUniqueUser(db *gorm.DB, ignoreUserID uint, username, email string) error {
	var existingUser models.User

	usernameQuery := db.Where("username = ?", username)
	if ignoreUserID > 0 {
		usernameQuery = usernameQuery.Where("id <> ?", ignoreUserID)
	}
	if err := usernameQuery.First(&existingUser).Error; err == nil {
		return errors.New("用户名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	emailQuery := db.Where("email = ?", email)
	if ignoreUserID > 0 {
		emailQuery = emailQuery.Where("id <> ?", ignoreUserID)
	}
	if err := emailQuery.First(&existingUser).Error; err == nil {
		return errors.New("邮箱已被注册")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	db := config.GetDB()

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	if err := db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetProfile(userID uint) (*models.User, error) {
	var user models.User
	db := config.GetDB()
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateProfile(userID uint, updateData map[string]interface{}) (*models.User, error) {
	var user models.User
	db := config.GetDB()
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	removeImmutableFields(updateData, "id", "created_at", "updated_at", "deleted_at", "role", "password")

	if err := db.Model(&user).Updates(updateData).Error; err != nil {
		return nil, err
	}

	db.First(&user, userID)
	return &user, nil
}

func (s *UserService) UpdateAvatar(userID uint, avatarURL string) error {
	db := config.GetDB()
	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error; err != nil {
		return err
	}
	return nil
}

type UserStats struct {
	TotalUsers     int64 `json:"total_users"`
	ActiveUsers    int64 `json:"active_users"`
	NewUsers       int64 `json:"new_users"`
	AdminUsers     int64 `json:"admin_users"`
	NormalUsers    int64 `json:"normal_users"`
	DisabledUsers  int64 `json:"disabled_users"`
}

type ActiveUsersStats struct {
	DailyActive    int64 `json:"daily_active"`
	WeeklyActive   int64 `json:"weekly_active"`
	MonthlyActive  int64 `json:"monthly_active"`
}

type OverviewStats struct {
	TotalUsers        int64  `json:"total_users"`
	TotalMaterials   int64  `json:"total_materials"`
	TotalTemplates   int64  `json:"total_templates"`
	StorageUsed      int64  `json:"storage_used_bytes"`
}

func (s *UserService) GetUserStats() (*UserStats, error) {
	db := config.GetDB()
	stats := &UserStats{}

	db.Model(&models.User{}).Count(&stats.TotalUsers)
	db.Model(&models.User{}).Where("role = ?", models.UserRoleAdmin).Count(&stats.AdminUsers)
	db.Model(&models.User{}).Where("role = ?", models.UserRoleUser).Count(&stats.NormalUsers)
	db.Model(&models.User{}).Where("status = ?", 0).Count(&stats.DisabledUsers)
	db.Model(&models.User{}).Where("status = ?", 1).Count(&stats.ActiveUsers)

	return stats, nil
}

func (s *UserService) GetActiveUsersStats() (*ActiveUsersStats, error) {
	stats := &ActiveUsersStats{}
	db := config.GetDB()

	var totalUsers int64
	db.Model(&models.User{}).Count(&totalUsers)

	stats.DailyActive = totalUsers
	stats.WeeklyActive = totalUsers
	stats.MonthlyActive = totalUsers

	return stats, nil
}

func (s *UserService) GetOverviewStats() (*OverviewStats, error) {
	stats := &OverviewStats{}
	db := config.GetDB()

	db.Model(&models.User{}).Count(&stats.TotalUsers)
	db.Model(&models.Material{}).Count(&stats.TotalMaterials)
	db.Model(&models.Template{}).Count(&stats.TotalTemplates)

	return stats, nil
}
