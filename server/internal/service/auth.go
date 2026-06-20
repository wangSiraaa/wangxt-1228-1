package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pvgrid/internal/config"
	"pvgrid/internal/dto"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type AuthService struct {
	cfg   *config.Config
	users repository.UserRepo
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg, users: repository.UserRepo{}}
}

func (s *AuthService) Login(req dto.LoginReq) (*dto.LoginResp, error) {
	u, err := s.users.FindByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.NewBizError(util.CodeUnauthorized, "phone or password incorrect", 401)
		}
		return nil, util.NewBizError(util.CodeInternal, err.Error(), 500)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, util.NewBizError(util.CodeUnauthorized, "phone or password incorrect", 401)
	}
	token, err := util.GenerateToken(s.cfg, u.ID, u.Role)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, "token generate failed", 500)
	}
	return &dto.LoginResp{
		Token: token,
		User:  dto.LoginUser{ID: u.ID, Phone: u.Phone, Name: u.Name, Role: u.Role},
	}, nil
}

// HashPassword 生成 bcrypt 密码哈希（供 seed 使用）
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}
