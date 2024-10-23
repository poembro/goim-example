package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var Secret = []byte("nfGDsxKwFE5yssxax3399PXnzlcY2bZbcp5cE61T0J7gsqZeTxuo5knGt9cbpCfK") // Secret(盐) 用来加密解密

type TokenInfo struct {
	Mid      int64  `json:"mid"`      // userid
	Key      string `json:"key"`      // 设备id
	Nickname string `json:"nickname"` // 头像
	jwt.RegisteredClaims
}

// GetToken 获取token
func GetToken(mid int64, key, nickname string) (string, error) {
	var claims = TokenInfo{
		Mid:      mid,
		Key:      key,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 过期时间
			Issuer:    "h5gameproject",
			IssuedAt:  jwt.NewNumericDate(time.Now()), // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()), // 签发人
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(Secret))
	if err != nil {
		return "", fmt.Errorf("生成token失败:%v", err)
	}
	return token, nil
}

func DecryptToken(tokenStr string) (*TokenInfo, error) {
	tokenItem, err := jwt.ParseWithClaims(tokenStr, &TokenInfo{}, func(token *jwt.Token) (interface{}, error) { // 解析token
		return Secret, nil
	})

	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if !ok {
			return nil, err
		}

		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("Token is invalid")
		}

		if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil, errors.New("token is expired")
		}
		return nil, err
	}

	if !tokenItem.Valid {
		return nil, errors.New("Token is invalid.")
	}

	if tokenItem.Method != jwt.SigningMethodHS256 {
		return nil, errors.New("Wrong signing method")
	}

	if claims, ok := tokenItem.Claims.(*TokenInfo); ok && tokenItem.Valid {
		return claims, nil
	}

	return nil, errors.New("couldn't handle this token")
}
