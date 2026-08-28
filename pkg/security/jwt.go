package security

import (
	"time"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// Gives JWT functionality the configuration it needs.
type JWTManager struct{
	secretKey     string
	tokenDuration time.Duration
}


//create new manager by giving correct configurations and returns pointer to  new JWT MANAGER
func NewJWTManager(secretKey string,durationHours int) *JWTManager{
	return &JWTManager{
		secretKey: secretKey,
		tokenDuration: time.Duration(durationHours)*time.Hour,
	}
}

//generate jwt 
func (manager *JWTManager) GenerateToken(userID string)(string,error){
	
	//claim contains info written inside token's payload
	claims:=jwt.MapClaims{
		"user_id":userID,
		"exp":     time.Now().Add(manager.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	//create jwt with hs256 crypto signing algo
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	//sign it with secret key and return the signed jwt
	return token.SignedString([]byte(manager.secretKey))
}


//verification of client's jwt
func (manager *JWTManager) VerifyToken(tokenString string)(string,error){
	
	//main verification happens here by taking user tokenstring and secret key if signing algo ok client's token matches the expected signing algo
	token,err:=jwt.Parse(tokenString,func(token *jwt.Token)(interface{},error){

		//checks if signing algo matches or not
		if _,ok:=token.Method.(*jwt.SigningMethodHMAC);!ok{
			return nil,fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		//return the secret key to parse function as argument if algo matches
		return []byte(manager.secretKey),nil


	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	//extract the info inside token
	claims,ok:=token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	//extarct the userID using claim we just extracted
	// Gets user_id from the claims and checks that it is actually a string.
	userID,ok:=claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}
	return userID,nil
}
