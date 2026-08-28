package config
import (
	"fmt"
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

type AppConfig struct{
	Port               string
	DatabaseURL        string
	JWTSecret          string
	JWTExpirationHours int
}

func LoadConfig()(*AppConfig,error){

	_=godotenv.Load()

	jwtExpStr:=os.Getenv("JWT_EXPIRATION_HOURS")

	jwtExp,err:=strconv.Atoi(jwtExpStr)
	if err != nil || jwtExp == 0 {
		jwtExp = 24 // Fallback default
	}

	config:=&AppConfig{
		Port:               os.Getenv("PORT"),
		DatabaseURL:        os.Getenv("DB_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTExpirationHours: jwtExp,
	}

	if config.JWTSecret==""{
		return nil, fmt.Errorf("CRITICAL: JWT_SECRET environment variable is not set")
	}

	return config,nil

}