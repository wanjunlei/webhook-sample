package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtkey = []byte("www.kubesphere.io")

type Claims struct {
	UserId uint
	jwt.StandardClaims
}

//°ä·¢token
func generateToken(user string) (string, error) {
	expireTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UserId: 2,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Id:        user,
			IssuedAt:  time.Now().Unix(),
			Issuer:    "127.0.0.1",
			Subject:   "user token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// fmt.Println(token)
	tokenString, err := token.SignedString(jwtkey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func parseToken(user, tokenString string) bool {

	if tokenString == "" {
		return false
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtkey, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	if claims.Id != user {
		return false
	}

	return true
}
