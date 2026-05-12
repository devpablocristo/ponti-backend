package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/core/security/go/contextkeys"
)

func (h *Handler) registerMeContextRoute() {
	group := h.gsv.GetRouter().Group(h.acf.APIBaseURL(), h.mws.GetValidation()...)
	group.GET("/me/context", h.GetMeContext)
}

func (h *Handler) GetMeContext(c *gin.Context) {
	actor, _ := c.Request.Context().Value(ctxkeys.Actor).(string)
	if strings.TrimSpace(actor) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "authentication context required"})
		return
	}

	type userRow struct {
		ID       uuid.UUID `json:"id"`
		IDPSub   string    `json:"idp_sub" gorm:"column:idp_sub"`
		IDPEmail string    `json:"idp_email" gorm:"column:idp_email"`
		Email    string    `json:"email"`
	}
	var user userRow
	if err := h.db.
		WithContext(c.Request.Context()).
		Table("users").
		Select("id, idp_sub, idp_email, email").
		Where("idp_sub = ?", actor).
		Limit(1).
		Take(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "local user not found"})
		return
	}

	type tenantRow struct {
		TenantID uuid.UUID `json:"tenant_id"`
		Name     string    `json:"name"`
		RoleID   uuid.UUID `json:"-"`
		RoleName string    `json:"role"`
	}
	var tenants []tenantRow
	if err := h.db.
		WithContext(c.Request.Context()).
		Table("auth_memberships AS m").
		Select("m.tenant_id, t.name, m.role_id, r.name AS role_name").
		Joins("JOIN auth_tenants t ON t.id = m.tenant_id").
		Joins("JOIN auth_roles r ON r.id = m.role_id").
		Where("m.user_id = ? AND m.status = 'active'", user.ID).
		Order("t.name ASC").
		Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to load memberships"})
		return
	}

	type permissionRow struct {
		RoleID uuid.UUID
		Name   string
	}
	var roleIDs []uuid.UUID
	for _, tenant := range tenants {
		roleIDs = append(roleIDs, tenant.RoleID)
	}
	permissionsByRole := map[uuid.UUID][]string{}
	if len(roleIDs) > 0 {
		var perms []permissionRow
		if err := h.db.
			WithContext(c.Request.Context()).
			Table("auth_role_permissions rp").
			Select("rp.role_id, p.name").
			Joins("JOIN auth_permissions p ON p.id = rp.permission_id").
			Where("rp.role_id IN ?", roleIDs).
			Order("p.name ASC").
			Find(&perms).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to load permissions"})
			return
		}
		for _, perm := range perms {
			permissionsByRole[perm.RoleID] = append(permissionsByRole[perm.RoleID], perm.Name)
		}
	}

	currentTenantID, _ := c.Request.Context().Value(ctxkeys.OrgID).(uuid.UUID)
	tenantPayload := make([]gin.H, 0, len(tenants))
	for _, tenant := range tenants {
		tenantPayload = append(tenantPayload, gin.H{
			"id":          tenant.TenantID,
			"name":        tenant.Name,
			"role":        tenant.RoleName,
			"permissions": permissionsByRole[tenant.RoleID],
			"is_current":  tenant.TenantID == currentTenantID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        user.ID,
			"idp_sub":   user.IDPSub,
			"idp_email": user.IDPEmail,
			"email":     user.Email,
		},
		"current_tenant_id": currentTenantID,
		"tenants":           tenantPayload,
	})
}
