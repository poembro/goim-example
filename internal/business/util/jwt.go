package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var Secret = []byte("golangproject") // Secret(盐) 用来加密解密

type TokenInfo struct {
	UserId   string `json:"user_id"`   // userid
	DeviceId string `json:"device_id"` // 设备id
	Nickname string `json:"nickname"`  // 头像
	jwt.StandardClaims
}

// GetToken 获取token
func GetToken(userId, deviceId, nickname string) (string, error) {
	var claims = TokenInfo{
		userId,
		deviceId,
		nickname,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(), //  过期时间
			Issuer:    "golangproject",                             // 签发人
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(Secret))
	if err != nil {
		return "", fmt.Errorf("生成token失败:%s", err.Error())
	}
	return signedToken, nil
}

func DecryptToken(tokenStr string) (*TokenInfo, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenInfo{}, func(token *jwt.Token) (i interface{}, err error) { // 解析token
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*TokenInfo); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
