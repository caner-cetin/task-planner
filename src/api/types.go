package api

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SESSION_TOKEN_EXPIRES_IN = jwt.NewNumericDate(time.Now().Add(time.Hour * 24))
var SESSION_TOKEN_CONTEXT_KEY = "user"
