package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/config"
	"pvgrid/internal/util"
)

// CORS 跨域中间件，允许前端 localhost:5173
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = cfg.CORSAllowOrigin
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

const ctxUserID = "userID"
const ctxRole = "role"

// JWTAuth 校验 Bearer Token
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			util.Fail(c, util.NewBizError(util.CodeUnauthorized, "missing or invalid authorization header", http.StatusUnauthorized))
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := util.ParseToken(cfg, tokenStr)
		if err != nil {
			util.Fail(c, util.NewBizError(util.CodeUnauthorized, "invalid token", http.StatusUnauthorized))
			c.Abort()
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// RequireRole 角色中间件，admin 拥有全部权限
func RequireRole(roles ...string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get(ctxRole)
		roleStr, _ := role.(string)
		if roleStr == "admin" {
			c.Next()
			return
		}
		if _, ok := allow[roleStr]; !ok {
			util.Fail(c, util.NewBizError(util.CodeForbidden, "insufficient role: "+roleStr, http.StatusForbidden))
			c.Abort()
			return
		}
		c.Next()
	}
}

// CurrentUserID 从上下文取当前用户ID
func CurrentUserID(c *gin.Context) uint64 {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(uint64)
	return id
}

// CurrentRole 从上下文取当前角色
func CurrentRole(c *gin.Context) string {
	v, _ := c.Get(ctxRole)
	r, _ := v.(string)
	return r
}
