package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"whatever/src/db"
	"whatever/src/db/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type JWTCustomClaims struct {
	ID uint
	jwt.RegisteredClaims
}

type SignUpRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func SignUp(c echo.Context) (err error) {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Bad request: %s", err.Error()))
	}
	if len(req.Password) > 36 {
		return c.String(http.StatusBadRequest, "Password is too long")
	}
	pwd_bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	student := models.Student{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Password:  string(pwd_bytes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result := db.DB.Create(&student)
	if result.Error != nil {
		error_text := result.Error.Error()
		if strings.Contains(error_text, "Duplicate") && strings.Contains(error_text, "Username") {
			return c.JSON(http.StatusConflict, ErrorResponse{Message: "This username is already taken.", Code: USERNAME_ALREADY_TAKEN})
		} else {
			// return the error directly to the echo for handling
			return err
		}
	}
	claims := &JWTCustomClaims{
		student.ID,
		jwt.RegisteredClaims{
			ExpiresAt: SESSION_TOKEN_EXPIRES_IN,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encoded_token, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	return c.JSON(http.StatusOK, echo.Map{
		"token": encoded_token,
	})
}

func Login(c echo.Context) (err error) {
	username := c.FormValue("username")
	password := c.FormValue("password")
	if len(password) > 36 {
		return c.String(http.StatusBadRequest, "Password is too long")
	}
	var student models.Student
	result := db.DB.Where("username = ?", username).First(&student)
	if result.Error != nil {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	err = bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(password))
	if err != nil {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	claims := &JWTCustomClaims{
		student.ID,
		jwt.RegisteredClaims{
			ExpiresAt: SESSION_TOKEN_EXPIRES_IN,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encoded_token, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	return c.JSON(http.StatusOK, echo.Map{
		"token": encoded_token,
	})
}

func GetAuthenticatedStudent(c echo.Context) (std *models.Student, err *ErrorResponse) {
	user := c.Get(SESSION_TOKEN_CONTEXT_KEY).(*jwt.Token)
	claims := user.Claims.(*JWTCustomClaims)
	id := claims.ID
	var student models.Student
	db.DB.First(&student, id)
	if student == (models.Student{}) {
		return nil, &ErrorResponse{
			Code:    USER_DOES_NOT_EXIST,
			Message: "Authenticated user does not exist.",
		}
	}
	return &student, nil
}
