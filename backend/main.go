// Package main wires up the BaseAdmin backend: config, database, redis,
// middleware stack, module routes, and the migrate/seed CLI subcommands.
//
// @title						BaseAdmin API
// @version					1.0
// @description				Admin Panel Base Project API
// @host						localhost:8880
// @BasePath					/api/v1
// @securityDefinitions.apikey	SessionCookie
// @in							cookie
// @name						session
//
//go:generate swag init -g main.go -o docs
package main

import (
	"log"
	"os"
	"time"

	"baseadmin/backend/config"
	"baseadmin/backend/database/migrations"
	"baseadmin/backend/database/seeders"
	_ "baseadmin/backend/docs"
	"baseadmin/backend/middleware"
	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/modules/auth"
	"baseadmin/backend/modules/bandara"
	"baseadmin/backend/modules/blazer_sizes"
	"baseadmin/backend/modules/dashboard"
	"baseadmin/backend/modules/kota_asal"
	"baseadmin/backend/modules/menu_sections"
	"baseadmin/backend/modules/menus"
	"baseadmin/backend/modules/peserta"
	"baseadmin/backend/modules/qr_gate"
	"baseadmin/backend/modules/qris_cross_border"
	"baseadmin/backend/modules/roles"
	"baseadmin/backend/modules/settings"
	"baseadmin/backend/modules/sliders"
	"baseadmin/backend/modules/users"
	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.Load()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			db := config.ConnectDatabase(cfg)
			migrations.Run(db)
			return
		case "seed":
			db := config.ConnectDatabase(cfg)
			seeders.Run(db)
			return
		}
	}

	db := config.ConnectDatabase(cfg)
	rdb := config.ConnectRedis(cfg)
	activity_logs.StartWorker(db)

	fileStorage := storage.NewS3Storage(cfg.S3Endpoint, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.StorageMaxSize)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware(cfg.FrontendURL))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestSizeLimit(50 << 20))
	router.Use(middleware.CSRF(cfg.AppEnv == "production", cfg.CookieDomain))
	router.Use(middleware.XSSSanitizer())
	router.Use(middleware.MaintenanceMode(rdb, db))
	router.Use(middleware.RateLimit(rdb, "api", cfg.RateLimitAPI, time.Minute, func(c *gin.Context) string { return c.ClientIP() }))

	sessionAuth := middleware.SessionAuth(rdb, db)
	guard := func(menuPath, action string) gin.HandlerFunc {
		return func(c *gin.Context) {
			sessionAuth(c)
			if c.IsAborted() {
				return
			}
			middleware.RequirePermission(db, menuPath, action)(c)
		}
	}
	requireSuperadmin := func(c *gin.Context) {
		if !utils.IsSuperadmin(c) {
			utils.Error(c, 403, "superadmin access required")
			c.Abort()
			return
		}
		c.Next()
	}

	if cfg.AppEnv != "production" {
		router.GET("/swagger/*any", sessionAuth, requireSuperadmin, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	uploadLimiter := middleware.RateLimit(rdb, "upload", 10, time.Minute, func(c *gin.Context) string {
		return c.ClientIP()
	})

	api := router.Group("/api/v1")

	authHandler := auth.NewHandler(db, rdb, cfg)
	authHandler.RegisterRoutes(api, sessionAuth)

	usersHandler := users.NewHandler(db, cfg, fileStorage, rdb)
	usersHandler.RegisterRoutes(api, guard, uploadLimiter)

	rolesHandler := roles.NewHandler(db)
	rolesHandler.RegisterRoutes(api, guard)

	menusHandler := menus.NewHandler(db)
	menusHandler.RegisterRoutes(api, sessionAuth, guard)

	menuSectionsHandler := menu_sections.NewHandler(db)
	menuSectionsHandler.RegisterRoutes(api, guard)

	kotaAsalHandler := kota_asal.NewHandler(db)
	kotaAsalHandler.RegisterRoutes(api, sessionAuth, guard)

	bandaraHandler := bandara.NewHandler(db)
	bandaraHandler.RegisterRoutes(api, sessionAuth, guard)

	blazerSizesHandler := blazer_sizes.NewHandler(db)
	blazerSizesHandler.RegisterRoutes(api, sessionAuth, guard)

	pesertaHandler := peserta.NewHandler(db, fileStorage, rdb)
	pesertaHandler.RegisterRoutes(api, sessionAuth, guard, uploadLimiter)

	qrGateHandler := qr_gate.NewHandler(db)
	qrGateHandler.RegisterRoutes(api, sessionAuth, guard)

	dashboardHandler := dashboard.NewHandler(db)
	dashboardHandler.RegisterRoutes(api, sessionAuth)

	ocrAPIKey, ocrModel := cfg.AnthropicAPIKey, cfg.AnthropicModel
	if cfg.OCRProvider == "deepseek" {
		ocrAPIKey, ocrModel = cfg.DeepSeekAPIKey, cfg.DeepSeekModel
	}
	qrisCrossBorderHandler := qris_cross_border.NewHandler(db, fileStorage, cfg.OCRProvider, ocrAPIKey, ocrModel)
	qrisCrossBorderHandler.RegisterRoutes(api, sessionAuth, guard, uploadLimiter)

	settingsHandler := settings.NewHandler(db, rdb, fileStorage)
	settingsHandler.RegisterRoutes(api, guard, uploadLimiter)

	slidersHandler := sliders.NewHandler(db, fileStorage)
	slidersHandler.RegisterRoutes(api, sessionAuth, guard, uploadLimiter)

	activityLogsHandler := activity_logs.NewHandler(db, cfg)
	activityLogsHandler.RegisterRoutes(api, guard)

	log.Printf("BaseAdmin backend listening on :%s (env=%s)", cfg.AppPort, cfg.AppEnv)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
