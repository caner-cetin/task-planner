package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
	"whatever/src/api"
	"whatever/src/db"
	"whatever/src/db/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dtlc := os.Getenv("DOTENV_LOCATION")
	var err error
	if dtlc == "" {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(dtlc)
	}
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(mysql)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_DATABASE"))
	sqdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	sqdb.AutoMigrate(&models.Plan{}, &models.Student{})
	pool, err := sqdb.DB()
	if err != nil {
		log.Fatal(err)
	}
	pool.SetMaxIdleConns(10)
	pool.SetMaxOpenConns(100)
	pool.SetConnMaxLifetime(time.Hour)
	db.Pool = pool
	db.DB = sqdb

	e := echo.New()

	// Configure middleware with the custom claims type
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(api.JWTCustomClaims)
		},
		SigningKey: []byte(os.Getenv("JWT_SECRET_KEY")),
		ContextKey: api.SESSION_TOKEN_CONTEXT_KEY,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("request-id", c.Response().Header().Get(echo.HeaderXRequestID)),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("request-id", c.Response().Header().Get(echo.HeaderXRequestID)),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "bu sarki yasayip da olenler icin")
	})
	// unauthenticated
	auth := e.Group("/auth")
	auth.POST("/login", api.Login)
	auth.POST("/signup", api.SignUp)

	students := e.Group("/students")
	students.Use(echojwt.WithConfig(jwtConfig))

	me := students.Group("/me")
	me.GET("", api.StudentMe)
	me.PATCH("", api.StudentUpdate)
	me.DELETE("", api.StudentDelete)

	plans := e.Group("/plans", echojwt.WithConfig(jwtConfig))
	plans.GET("/", api.GetMyPlans)
	plans.POST("/create", api.CreatePlan)
	plans.DELETE("/:id", api.DeletePlan)
	plans.PATCH("/status", api.UpdatePlanStatus)
	plans.PATCH("/deadline", api.UpdatePlanDeadline)
	plans.GET("/interval", api.GetPlansByInterval)

	by := plans.Group("/by", middleware.BodyLimit("256K"))
	by.GET("/year", api.GetPlansByYear)
	by.GET("/month", api.GetPlansByMonth)
	by.GET("/day", api.GetPlansByDay)

	within := plans.Group("/within", middleware.BodyLimit("0K")) // no body allowed
	within.GET("/year", api.GetPlansByThisYear)
	within.GET("/month", api.GetPlansByThisMonth)
	within.GET("/week", api.GetPlansByThisWeek)
	within.GET("/day", api.GetPlansByThisDay)

	e.Logger.Fatal(e.Start(":1323"))
}
