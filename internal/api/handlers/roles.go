// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"v/internal/logger"
)

// Role represents a user role.
type Role struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	UserCount   int      `json:"user_count"`
	IsSystem    bool     `json:"is_system"`
}

// RoleHandler handles role-related requests.
type RoleHandler struct {
	logger logger.Logger
	roles  map[int64]*Role
}

// NewRoleHandler creates a new RoleHandler.
func NewRoleHandler(log logger.Logger) *RoleHandler {
	// Initialize with default roles
	roles := map[int64]*Role{
		1: {
			ID:          1,
			Name:        "admin",
			Description: "系统管理员，拥有所有权限",
			Permissions: []string{"*"},
			UserCount:   1,
			IsSystem:    true,
		},
		2: {
			ID:          2,
			Name:        "user",
			Description: "普通用户，可以管理自己的代理",
			Permissions: []string{"proxy:read", "proxy:write", "profile:read", "profile:write"},
			UserCount:   0,
			IsSystem:    true,
		},
		3: {
			ID:          3,
			Name:        "viewer",
			Description: "只读用户，只能查看信息",
			Permissions: []string{"proxy:read", "profile:read", "stats:read"},
			UserCount:   0,
			IsSystem:    true,
		},
	}

	return &RoleHandler{
		logger: log,
		roles:  roles,
	}
}

// ListRoles returns all roles.
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles := make([]*Role, 0, len(h.roles))
	for _, role := range h.roles {
		roles = append(roles, role)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    roles,
	})
}

// GetRole returns a specific role.
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid role id",
		})
		return
	}

	role, exists := h.roles[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "role not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    role,
	})
}

// CreateRoleRequest represents a create role request.
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// CreateRole creates a new role.
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	// Check if role name already exists
	for _, role := range h.roles {
		if role.Name == req.Name {
			c.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": "role name already exists",
			})
			return
		}
	}

	// Generate new ID
	var maxID int64
	for id := range h.roles {
		if id > maxID {
			maxID = id
		}
	}

	newRole := &Role{
		ID:          maxID + 1,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		UserCount:   0,
		IsSystem:    false,
	}

	h.roles[newRole.ID] = newRole

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "role created",
		"data":    newRole,
	})
}

// UpdateRoleRequest represents an update role request.
type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// UpdateRole updates an existing role.
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid role id",
		})
		return
	}

	role, exists := h.roles[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "role not found",
		})
		return
	}

	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "cannot modify system role",
		})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Permissions != nil {
		role.Permissions = req.Permissions
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "role updated",
		"data":    role,
	})
}

// DeleteRole deletes a role.
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid role id",
		})
		return
	}

	role, exists := h.roles[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "role not found",
		})
		return
	}

	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "cannot delete system role",
		})
		return
	}

	if role.UserCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "cannot delete role with assigned users",
		})
		return
	}

	delete(h.roles, id)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "role deleted",
	})
}

// GetPermissions returns all available permissions.
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	permissions := []map[string]string{
		{"key": "*", "name": "所有权限", "description": "拥有系统所有权限"},
		{"key": "proxy:read", "name": "查看代理", "description": "查看代理配置"},
		{"key": "proxy:write", "name": "管理代理", "description": "创建、修改、删除代理"},
		{"key": "user:read", "name": "查看用户", "description": "查看用户列表"},
		{"key": "user:write", "name": "管理用户", "description": "创建、修改、删除用户"},
		{"key": "role:read", "name": "查看角色", "description": "查看角色列表"},
		{"key": "role:write", "name": "管理角色", "description": "创建、修改、删除角色"},
		{"key": "stats:read", "name": "查看统计", "description": "查看系统统计数据"},
		{"key": "system:read", "name": "查看系统", "description": "查看系统信息"},
		{"key": "system:write", "name": "管理系统", "description": "修改系统设置"},
		{"key": "profile:read", "name": "查看个人信息", "description": "查看个人资料"},
		{"key": "profile:write", "name": "修改个人信息", "description": "修改个人资料"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    permissions,
	})
}
