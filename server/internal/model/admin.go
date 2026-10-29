package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin" // 超级管理员：全部权限，含账号管理
	RoleAdmin      Role = "admin"       // 管理员：区服配置、操作日志
	RoleOperator   Role = "operator"    // 运营：玩家操作、邮件、公告
	RoleViewer     Role = "viewer"      // 只读
)

var roleLevel = map[Role]int{
	RoleViewer:     1,
	RoleOperator:   2,
	RoleAdmin:      3,
	RoleSuperAdmin: 4,
}

func (r Role) Level() int  { return roleLevel[r] }
func (r Role) Valid() bool { _, ok := roleLevel[r]; return ok }

type AdminUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Nickname     string             `bson:"nickname" json:"nickname"`
	Role         Role               `bson:"role" json:"role"`
	Disabled     bool               `bson:"disabled" json:"disabled"`
	LastLoginAt  *time.Time         `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
