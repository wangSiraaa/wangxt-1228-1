package config

import "os"

type Config struct {
	DBDriver string
	MySQLDSN string
	SQLitePath string
	JWTSecret string
	JWTExpireHours int
	ListenAddr string
	CORSAllowOrigin string
}

func Load() *Config {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "mysql"
	}
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/pvgrid?charset=utf8mb4&parseTime=True&loc=Local"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pvgrid-dev-secret-change-me"
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	cors := os.Getenv("CORS_ALLOW_ORIGIN")
	if cors == "" {
		cors = "http://localhost:5173"
	}
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "pvgrid.db"
	}
	return &Config{
		DBDriver:        driver,
		MySQLDSN:        dsn,
		SQLitePath:      sqlitePath,
		JWTSecret:       secret,
		JWTExpireHours:  24,
		ListenAddr:      addr,
		CORSAllowOrigin: cors,
	}
}
