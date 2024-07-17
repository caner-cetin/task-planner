package api

import (
	"net/http"
	"strconv"
	"time"
	"whatever/src/db"
	"whatever/src/db/models"

	"github.com/labstack/echo/v4"
)

type CreatePlanRequest struct {
	Title             string            `json:"title"`
	Description       *string           `json:"description"`
	Status            models.PlanStatus `json:"status"`
	Time              *time.Time        `json:"time"`
	EstimatedDeadline *time.Time        `json:"estimated_deadline"`
}

func CreatePlan(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	var req CreatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	if req.Time != nil && req.EstimatedDeadline != nil {
		startTime := req.Time
		endTime := req.Time.Add(req.EstimatedDeadline.Sub(*req.Time))
		var plans []models.Plan
		db.DB.Where("student_id = ? AND ((time >= ? AND time <= ?) OR (time >= ? AND time <= ?))", student.ID, startTime, endTime, endTime, startTime).Find(&plans)
		if len(plans) > 0 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "The plan clashes with another plan.", Code: PLAN_CLASH})
		}
	}
	var plan models.Plan
	right_now := time.Now()
	if req.Time == nil {
		plan.Time = right_now
	} else {
		utc_plan_time := (*req.Time).UTC()
		if utc_plan_time.Before(right_now) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Time cannot be in the past.", Code: PLAN_TIME_CANNOT_BE_IN_THE_PAST})
		}
		plan.Time = utc_plan_time
	}
	if req.EstimatedDeadline != nil {
		dd := (req.EstimatedDeadline).UTC()
		plan.EstimatedDeadline = &dd
	}
	plan.CreatedAt = right_now
	plan.UpdatedAt = right_now
	plan.Title = req.Title
	plan.Description = req.Description
	plan.Status = req.Status
	plan.StudentID = student.ID
	plan.Student = *student
	db.DB.Save(&plan)
	if plan.StudentID == 0 {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create a plan.", Code: INTERNAL_SERVER_ERROR})
	}
	return c.JSON(http.StatusCreated, plan)
}

func DeletePlan(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	planID := c.Param("id")
	var plan models.Plan
	result := db.DB.Where("id = ? AND student_id = ?", planID, student.ID).First(&plan)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "Plan not found.", Code: PLAN_NOT_FOUND})
	}
	result = db.DB.Delete(&plan)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: result.Error.Error(), Code: INTERNAL_SERVER_ERROR})
	}
	return c.NoContent(http.StatusNoContent)
}

func GetMyPlans(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	var plans []models.Plan
	db.DB.Where("student_id = ?", student.ID).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func UpdatePlanStatus(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	planID := c.FormValue("id")
	var plan models.Plan
	result := db.DB.Where("id = ? AND student_id = ?", planID, student.ID).First(&plan)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "Plan not found.", Code: PLAN_NOT_FOUND})
	}
	status, err := strconv.Atoi(c.FormValue("status"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid status enum. Please provide a valid status.", Code: PLAN_STATUS_INVALID})
	}
	plan.Status = models.PlanStatus(status)
	plan.UpdatedAt = time.Now()
	db.DB.Save(&plan)
	return c.JSON(http.StatusOK, plan)
}

func UpdatePlanDeadline(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	planID := c.FormValue("id")
	var plan models.Plan
	result := db.DB.Where("id = ? AND student_id = ?", planID, student.ID).First(&plan)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "Plan not found.", Code: PLAN_NOT_FOUND})
	}
	deadline, err := time.Parse(time.RFC3339, c.FormValue("new_deadline"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	deadline = deadline.UTC()
	plan.EstimatedDeadline = &deadline
	plan.UpdatedAt = time.Now()
	db.DB.Save(&plan)
	return c.JSON(http.StatusOK, plan)
}

func GetPlansByInterval(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	startTime, err := time.Parse(time.RFC3339, c.FormValue("start_time"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	endTime, err := time.Parse(time.RFC3339, c.FormValue("end_time"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	var plans []models.Plan
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, startTime, endTime).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByDay(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date, err := time.Parse("2006-01-02", c.FormValue("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	var plans []models.Plan
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, date, date.AddDate(0, 0, 1)).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByMonth(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date, err := time.Parse("2006-01", c.FormValue("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	var plans []models.Plan
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, date, date.AddDate(0, 1, 0)).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByYear(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date, err := time.Parse("2006", c.FormValue("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error(), Code: INVALID_REQUEST})
	}
	var plans []models.Plan
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, date, date.AddDate(1, 0, 0)).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByThisYear(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date := time.Now()
	var plans []models.Plan
	start := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(date.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, start, end).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByThisMonth(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date := time.Now()
	var plans []models.Plan
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, start, end).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByThisWeek(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date := time.Now()
	var plans []models.Plan
	start := date.AddDate(0, 0, -int(date.Weekday()))
	end := start.AddDate(0, 0, 7)
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, start, end).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}

func GetPlansByThisDay(c echo.Context) error {
	student, errorResponse := GetAuthenticatedStudent(c)
	if errorResponse != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}
	date := time.Now()
	var plans []models.Plan
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	db.DB.Where("student_id = ? AND time >= ? AND time <= ?", student.ID, start, end).Find(&plans)
	return c.JSON(http.StatusOK, plans)
}
