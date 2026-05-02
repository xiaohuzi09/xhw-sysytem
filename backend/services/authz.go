package services

import (
	"errors"

	"xhw-service/config"
	"xhw-service/models"
)

var ErrPermissionDenied = errors.New("无权限操作")

func isAdminRole(role string) bool {
	return role == models.UserRoleAdmin
}

func resolveRoleByUsername(username string) string {
	if config.IsAdminUsername(username) {
		return models.UserRoleAdmin
	}
	return models.UserRoleUser
}

func canAccessUser(currentUserID, targetUserID uint, currentUserRole string) bool {
	return isAdminRole(currentUserRole) || currentUserID == targetUserID
}

func removeImmutableFields(updateData map[string]interface{}, fields ...string) {
	for _, field := range fields {
		delete(updateData, field)
	}
}
