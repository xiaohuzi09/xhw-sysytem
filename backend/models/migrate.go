package models

import (
	"xhw-service/config"

	"log"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	db := config.GetDB()

	err := db.AutoMigrate(
		&User{},
		&Template{},
		&Material{},
	)
	if err != nil {
		return err
	}

	if db.Migrator().HasIndex(&Material{}, "uni_materials_code") {
		if err := db.Migrator().DropIndex(&Material{}, "uni_materials_code"); err != nil {
			return err
		}
	}

	if err := SyncUserRoles(); err != nil {
		return err
	}

	log.Println("Database migration completed successfully")
	return nil
}

// SyncUserRoles 按配置中的管理员白名单同步用户角色
func SyncUserRoles() error {
	db := config.GetDB()

	var users []User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		role := UserRoleUser
		if config.IsAdminUsername(user.Username) {
			role = UserRoleAdmin
		}
		if user.Role == role {
			continue
		}
		if err := db.Model(&User{}).Where("id = ?", user.ID).Update("role", role).Error; err != nil {
			return err
		}
	}

	return nil
}
