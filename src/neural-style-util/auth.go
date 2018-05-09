package NSUtil

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// SecretKey used encode
var (
	SecretKey = os.Getenv("TOKEN_KEY")

	AuthToken = "Token"

	// ErrTokenContextMissing denotes a token was not passed into the parsing
	// middleware's context.
	ErrTokenContextMissing = errors.New("token up for parsing was not passed through the context")

	// ErrTokenInvalid denotes a token was not able to be validated.
	ErrTokenInvalid = errors.New("JWT Token was invalid")

	// ErrTokenExpired denotes a token's expire header (exp) has since passed.
	ErrTokenExpired = errors.New("JWT Token is expired")

	// ErrTokenMalformed denotes a token was not formatted as a JWT token.
	ErrTokenMalformed = errors.New("JWT Token is malformed")

	// ErrTokenNotActive denotes a token's not before header (nbf) is in the
	// future.
	ErrTokenNotActive = errors.New("token is not valid yet")

	// ErrUnexpectedSigningMethod denotes a token was signed with an unexpected
	// signing method.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// CheckToken validate the token
func CheckToken(authString string, log log.Logger) (string, error) {
	authList := strings.Split(authString, " ")
	if len(authList) != 2 || authList[0] != "Bearer" {
		level.Error(log).Log("API", "CheckToken", "info", "No authorization info")
		return "", errors.New("Unkown authorization info")
	}

	tokenString := authList[1]
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		level.Error(log).Log("API", "CheckToken", "info", "Token parse error", "err", err.Error())
		return "", errors.New("Bad Token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		level.Error(log).Log("API", "CheckToken", "info", "Can't access claims", "err", err.Error())
		return "", errors.New("No access claims")
	}
	user := claims["username"].(string)

	return user, nil
}

// GetUsername get parse user name
func GetUsername(authString string, log log.Logger) error {
	_, err := CheckToken(authString, log)
	return err
}

// UsernameMiddleware support user name parsing
func UsernameMiddleware(log log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			tokenString, ok := ctx.Value(AuthToken).(string)
			if !ok || len(tokenString) == 0 {
				return nil, NewErrorWithStatus(http.StatusNonAuthoritativeInfo, "Missing Authorization token")
			}

			err = GetUsername(tokenString, log)
			if err != nil {
				return nil, err
			}

			return next(ctx, request)
		}
	}
}

// ParseToken parse the jwt token and store it in context
func ParseToken(ctx context.Context, req *http.Request) context.Context {
	authString := req.Header.Get("Authorization")
	return context.WithValue(ctx, AuthToken, authString)
}

// AuthMiddleware support authorization
func AuthMiddleware(log log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			tokenString, ok := ctx.Value(AuthToken).(string)
			if !ok || len(tokenString) == 0 {
				return nil, NewErrorWithStatus(http.StatusNonAuthoritativeInfo, "Missing Authorization token")
			}

			err = GetUsername(tokenString, log)
			if err != nil {
				if e, ok := err.(*jwt.ValidationError); ok {
					switch {
					case e.Errors&jwt.ValidationErrorMalformed != 0:
						err = ErrTokenMalformed
					case e.Errors&jwt.ValidationErrorExpired != 0:
						// Token is expired
						err = ErrTokenMalformed
					case e.Errors&jwt.ValidationErrorNotValidYet != 0:
						// Token is not active yet
						err = ErrTokenNotActive
					case e.Inner != nil:
						// report e.Inner
						err = e.Inner
					}
				}

				return nil, NewErrorWithStatus(http.StatusUnauthorized, err.Error())
			}

			return next(ctx, request)
		}
	}
}

// AccessControl control the CORS
func AccessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		h.ServeHTTP(w, r)
	})
}
