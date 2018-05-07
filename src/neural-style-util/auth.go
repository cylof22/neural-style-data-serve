package NSUtil

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// SecretKey used encode
const (
	SecretKey = "Tulian is great"
)

// CheckToken validate the token
func CheckToken(authString string) (bool, string) {
	authList := strings.Split(authString, " ")
	if len(authList) != 2 || authList[0] != "Bearer" {
		fmt.Println("No authorization info")
		return false, ""
	}

	tokenString := authList[1]
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		fmt.Println("parse claims failed: ", err)
		return false, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Can't access claims")
		return false, ""
	}
	user := claims["username"].(string)

	return true, user
}

// GetUsername get parse user name
func GetUsername(req *http.Request) string {
	var authString = ""
	if len(req.Header) > 0 {
		for k, v := range req.Header {
			if k == "Authorization" {
				authString = v[0]
			}
		}
	}

	if authString == "" {
		return ""
	}

	_, username := CheckToken(authString)
	return username
}

// UsernameMiddleware support user name parsing
func UsernameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetUsername(r)
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware support authorization
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := GetUsername(r)
		if username == "" {
			//w.Header().Set("Content-Type", "text/html; text/javascript; text/css; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// CreateToken create time-limited token
func CreateToken(userName string) string {
	claims := make(jwt.MapClaims)
	claims["username"] = userName
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() //72小时有效期，过期需要重新登录获取token
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		fmt.Println("Error for sign token: ")
		fmt.Println(err)
		return ""
	}

	return tokenString
}

// AccessControl control the CORS
func AccessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		h.ServeHTTP(w, r)
	})
}
