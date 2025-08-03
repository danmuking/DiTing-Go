package jwt

import (
	"DiTing-Go/global"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"time"
)

var jwtSecret = []byte(viper.GetString("jwt.secret"))

type JwtClaims struct {
	Uid int64 `json:"uid"`
	jwt.RegisteredClaims
}

// GenerateToken 生成token
func GenerateToken(uid int64) (string, error) {
	registeredClaims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	claims := JwtClaims{
		uid,
		registeredClaims,
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	if err != nil {
		global.Logger.Errorf("generate token failed: %v", err)
	}
	return token, err
}

// ParseToken 解析token
func ParseToken(tokenString string) (*JwtClaims, error) {
	// 解析token
	claims := &JwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		return claims, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("token无法解析")
}
