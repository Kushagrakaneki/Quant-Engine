package api

import (
	"context"
	"net/http"
	

	"github.com/Kushagrakaneki/Quant-Engine/pkg/security"
)

type contextKey string

const UserContextKey=contextKey("userID")

func JWTMiddleware(jwtManager *security.JWTManager) func(http.Handler)http.Handler{
	return func(next http.Handler) http.Handler{
		return http.HandlerFunc(func(w http.ResponseWriter,req *http.Request){
			
			tokenString:=req.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
				return
			}

			userID,err:=jwtManager.VerifyToken(tokenString)
			if err!=nil{
				http.Error(w,"Invalid or Expired Token",http.StatusUnauthorized)
				return
			}

			ctx:=context.WithValue(req.Context(),UserContextKey,userID)

			next.ServeHTTP(w,req.WithContext(ctx))


		})
	}
}