package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

type Claims struct { //载荷
	username string
	password string
	jwt.StandardClaims
}

var jwtpwd = "thegua"

func Getjwtpwd() string {
	return jwtpwd
}

func GenerateToken(username string, password string) (string, error) {
	Claims := Claims{
		username: username,
		password: password,
		StandardClaims: jwt.StandardClaims{ //可以直接使用包含的结构
			ExpiresAt: 60 * 60 * 5,
			Issuer:    "gua",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	tokenstring, err := token.SignedString(Getjwtpwd())
	if err != nil {
		fmt.Println("token genrate err:", err)
		return "", err
	}
	return tokenstring, nil
}

func ParseToken(tokenstring string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenstring, Claims{}, func(t *jwt.Token) (interface{}, error) {
		return Getjwtpwd(), nil
	})
	if err != nil {
		fmt.Println("token parse err:", err)
		return "", "", err
	}
	if claims, ok := token.Claims.(Claims); ok && token.Valid { //ok 是指interface转化成功 token.valid 是token自身带的属性 可以用于判断是否过期
		return claims.username, claims.password, nil
	}
	return "", "", err
}
