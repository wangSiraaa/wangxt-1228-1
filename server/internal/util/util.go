package util

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"pvgrid/internal/config"
)

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 业务错误码与 §5 业务规则对齐
const (
	CodeOK                = "ok"
	CodeCapacityInsufficient = "capacity_insufficient"
	CodeAlarmUnhandled    = "alarm_unhandled"
	CodeBadReq            = "bad_request"
	CodeNotFound           = "not_found"
	CodeForbidden          = "forbidden"
	CodeUnauthorized        = "unauthorized"
	CodeInternal            = "internal_error"
)

// BizError 业务错误，携带 code
type BizError struct {
	Code    string
	Message string
	Status  int
}

func (e *BizError) Error() string { return e.Message }

func NewBizError(code, message string, status int) *BizError {
	return &BizError{Code: code, Message: message, Status: status}
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeOK, Message: "success", Data: data})
}

func Fail(c *gin.Context, berr *BizError) {
	c.JSON(berr.Status, Response{Code: berr.Code, Message: berr.Message})
}

// JWTClaims 自定义声明
type JWTClaims struct {
	UserID uint64 `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(cfg *config.Config, userID uint64, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pvgrid",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseToken 解析并校验 JWT
func ParseToken(cfg *config.Config, tokenStr string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
