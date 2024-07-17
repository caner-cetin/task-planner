package api

import (
	"net/http"
	"whatever/src/db"
	"whatever/src/db/models"

	"github.com/labstack/echo/v4"
)

func StudentMe(c echo.Context) (err error) {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	return c.JSON(http.StatusOK, student)
}

type StudentUpdateRequest struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Username  *string `json:"username"`
}

func StudentUpdate(c echo.Context) (err error) {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	var req models.Student
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	if req.FirstName == "" && req.LastName == "" && req.Username == "" {
		return c.NoContent(http.StatusNoContent)
	}
	if req.FirstName != "" {
		if student.FirstName == req.FirstName {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "First name is the same as the current one.", Code: USER_FIRST_NAME_IS_SAME_AS_CURRENT_ONE})
		}
		student.FirstName = req.FirstName
	}
	if req.LastName != "" {
		if student.LastName == req.LastName {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Last name is the same as the current one.", Code: USER_LAST_NAME_IS_SAME_AS_CURRENT_ONE})
		}
		student.LastName = req.LastName
	}
	if req.Username != "" {
		if student.Username == req.Username {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Username is the same as the current one.", Code: USER_NAME_IS_SAME_AS_CURRENT_ONE})
		}
		student.Username = req.Username
	}
	result := db.DB.Save(&student)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: result.Error.Error(), Code: INTERNAL_SERVER_ERROR})
	}
	return c.JSON(http.StatusOK, student)
}

func StudentDelete(c echo.Context) (err error) {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	result := db.DB.Delete(&student)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: result.Error.Error(), Code: INTERNAL_SERVER_ERROR})
	}
	return c.NoContent(http.StatusNoContent)
}
