package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/repository"
	"slg_game_gm/server/pkg/errcode"
	"slg_game_gm/server/pkg/jwtutil"
)

type AdminService struct {
	repo *repository.AdminRepo
	jwt  *jwtutil.Manager
}

func NewAdminService(repo *repository.AdminRepo, jwt *jwtutil.Manager) *AdminService {
	return &AdminService{repo: repo, jwt: jwt}
}

func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// EnsureSuperAdmin 首次启动时创建初始超级管理员。
func (s *AdminService) EnsureSuperAdmin(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return nil
	}
	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now()
	if err := s.repo.Create(ctx, &model.AdminUser{
		Username: username, PasswordHash: hash, Nickname: "超级管理员",
		Role: model.RoleSuperAdmin, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return err
	}
	slog.Warn("initial super admin created, please change the password", "username", username)
	return nil
}

func (s *AdminService) Login(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	u, err := s.repo.FindByUsername(ctx, req.Username)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errcode.LoginFailed
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return nil, errcode.LoginFailed
	}
	if u.Disabled {
		return nil, errcode.UserDisabled
	}
	token, expireAt, err := s.jwt.Generate(u.ID.Hex(), u.Username)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u.LastLoginAt = &now
	if err := s.repo.Update(ctx, u.ID, bson.M{"last_login_at": now}); err != nil {
		slog.Warn("update last login failed", "err", err)
	}
	return &dto.LoginResp{Token: token, ExpireAt: expireAt, User: u}, nil
}

// Authenticate 校验 token 并加载当前用户，每次请求都读库以便及时感知禁用和角色变更。
func (s *AdminService) Authenticate(ctx context.Context, token string) (*model.AdminUser, error) {
	claims, err := s.jwt.Parse(token)
	if err != nil {
		return nil, errcode.Unauthorized
	}
	id, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, errcode.Unauthorized
	}
	u, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errcode.Unauthorized
	}
	if err != nil {
		return nil, err
	}
	if u.Disabled {
		return nil, errcode.UserDisabled
	}
	return u, nil
}

func (s *AdminService) ChangePassword(ctx context.Context, u *model.AdminUser, req dto.ChangePasswordReq) error {
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)) != nil {
		return errcode.InvalidParams.WithMsg("原密码错误")
	}
	return s.setPassword(ctx, u.ID, req.NewPassword)
}

func (s *AdminService) setPassword(ctx context.Context, id primitive.ObjectID, pw string) error {
	hash, err := hashPassword(pw)
	if err != nil {
		return err
	}
	return repoErr(s.repo.Update(ctx, id, bson.M{"password_hash": hash, "updated_at": time.Now()}))
}

func (s *AdminService) List(ctx context.Context, q dto.AdminListQuery) ([]model.AdminUser, int64, error) {
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *AdminService) Create(ctx context.Context, req dto.CreateAdminReq) (*model.AdminUser, error) {
	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u := &model.AdminUser{
		Username: req.Username, PasswordHash: hash, Nickname: req.Nickname,
		Role: req.Role, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, errcode.Conflict.WithMsg("用户名已存在")
		}
		return nil, err
	}
	return u, nil
}

func (s *AdminService) Update(ctx context.Context, operator *model.AdminUser, id primitive.ObjectID, req dto.UpdateAdminReq) error {
	set := bson.M{"updated_at": time.Now()}
	if req.Nickname != nil {
		set["nickname"] = *req.Nickname
	}
	if req.Role != nil {
		if operator.ID == id && *req.Role != operator.Role {
			return errcode.InvalidParams.WithMsg("不能修改自己的角色")
		}
		set["role"] = *req.Role
	}
	if req.Disabled != nil {
		if operator.ID == id && *req.Disabled {
			return errcode.InvalidParams.WithMsg("不能禁用自己")
		}
		set["disabled"] = *req.Disabled
	}
	return repoErr(s.repo.Update(ctx, id, set))
}

func (s *AdminService) ResetPassword(ctx context.Context, id primitive.ObjectID, req dto.ResetPasswordReq) error {
	return s.setPassword(ctx, id, req.Password)
}

func (s *AdminService) Delete(ctx context.Context, operator *model.AdminUser, id primitive.ObjectID) error {
	if operator.ID == id {
		return errcode.InvalidParams.WithMsg("不能删除自己")
	}
	return repoErr(s.repo.Delete(ctx, id))
}
