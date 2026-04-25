package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Cfg struct {
	//APIの待受ポート
	Port string

	//PostgreSQL接続情報
	PgUser string
	PgPass string
	PgDB   string
	PgHost string
	PgPort int

	//実行環境
	GoEnv     string
	JWTSecret string

	//フロントエンドURL
	FEURL string

	//admin seed
	SeedAdminEmail    string
	SeedAdminPassword string

	RedisHost     string
	RedisPort     string
	RedisDB       int
	RedisPassword string

	GeminiUseMock bool
	GeminiAPIKey  string
	GeminiModel   string
}

func Load() (Cfg, error) {
	_ = godotenv.Load()

	c := Cfg{}

	//APIのポート
	c.Port = getenv("PORT", "8080")

	//DB接続情報
	c.PgUser = mustGet("POSTGRES_USER")
	c.PgPass = mustGet("POSTGRES_PASSWORD")
	c.PgDB = mustGet("POSTGRES_DB")
	c.PgHost = getenv("POSTGRES_HOST", "localhost")

	//実行環境
	c.GoEnv = getenv("GO_ENV", "dev")

	//JWT署名鍵
	c.JWTSecret = mustGet("JWT_SECRET")

	//フロントエンドURL
	c.FEURL = getenv("FE_URL", "http://localhost:3000")

	//DBポート
	p := getenv("POSTGRES_PORT", "5433")
	n, err := strconv.Atoi(p)
	if err != nil {
		n = 5433
	}
	c.PgPort = n

	//admin seed
	c.SeedAdminEmail = getenv("SEED_ADMIN_EMAIL", "admin@test.com")
	c.SeedAdminPassword = getenv("SEED_ADMIN_PASSWORD", "AdminPass123!")

	c.RedisHost = getenv("REDIS_HOST", "localhost")
	c.RedisPort = getenv("REDIS_PORT", "6379")
	c.RedisPassword = getenv("REDIS_PASSWORD", "")
	redisDB, err := strconv.Atoi(getenv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}
	c.RedisDB = redisDB
	c.GeminiUseMock = getenv("GEMINI_USE_MOCK", "true") == "true"
	c.GeminiAPIKey = getenv("GEMINI_API_KEY", "")
	c.GeminiModel = getenv("GEMINI_MODEL", "gemini-1.5-flash")

	return c, nil
}

func getenv(k string, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func mustGet(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}
