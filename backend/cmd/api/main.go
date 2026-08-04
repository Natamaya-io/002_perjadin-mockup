package main

import (
	"context"
	"database/sql"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/auth"
	"github.com/kemnaker/perjadin-backend/internal/domain/dalkot"
	"github.com/kemnaker/perjadin-backend/internal/domain/gup"
	"github.com/kemnaker/perjadin-backend/internal/domain/gupls"
	"github.com/kemnaker/perjadin-backend/internal/domain/gupreport"
	"github.com/kemnaker/perjadin-backend/internal/domain/master"
	"github.com/kemnaker/perjadin-backend/internal/domain/record"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/middleware"
	"github.com/kemnaker/perjadin-backend/internal/seeder"
	"github.com/kemnaker/perjadin-backend/pkg/cache"
	"github.com/kemnaker/perjadin-backend/pkg/database"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func runMigrations(db *sql.DB, sugar *zap.SugaredLogger) {
	sugar.Info("Running migrations...")

	driver, err := postgres.WithInstance(
		db,
		&postgres.Config{},
	)
	if err != nil {
		sugar.Fatalf(
			"Could not create postgres migration driver: %v",
			err,
		)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file:///app/db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		sugar.Fatalf(
			"Migration initialization failed: %v",
			err,
		)
	}

	if err := migration.Up(); err != nil &&
		err != migrate.ErrNoChange {
		if strings.Contains(
			strings.ToLower(err.Error()),
			"dirty database",
		) {
			sugar.Fatalf(
				"Migration database is dirty. Fix the failed migration and force the correct version manually before restarting. Error: %v",
				err,
			)
		}

		sugar.Fatalf(
			"Migration failed: %v",
			err,
		)
	}

	sugar.Info("Migrations applied successfully.")
}

func main() {
	cfg := config.LoadConfig()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = logger.Sync()
	}()

	sugar := logger.Sugar()

	db, err := database.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		sugar.Fatalf(
			"Failed to connect to database: %v",
			err,
		)
	}

	defer func() {
		sugar.Info("Closing database connection...")

		if closeErr := db.Close(); closeErr != nil {
			sugar.Errorf(
				"Error closing database connection: %v",
				closeErr,
			)
		}
	}()

	rdb, err := cache.NewRedisClient(cfg)
	if err != nil {
		sugar.Warnf(
			"Failed to connect to Redis: %v. Running without cache.",
			err,
		)

		rdb = nil
	} else {
		defer func() {
			if closeErr := rdb.Close(); closeErr != nil {
				sugar.Errorf(
					"Error closing Redis connection: %v",
					closeErr,
				)
			}
		}()
	}

	runMigrations(db, sugar)

	if os.Getenv("SEED_DB") == "true" {
		sugar.Info("Seeding database...")
		seeder.Seed(db, rdb)
	}

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(
		userRepo,
		cfg,
		rdb,
	)
	userHandler := user.NewHandler(userSvc)

	masterRepo := master.NewRepository(db)
	masterSvc := master.NewService(
		masterRepo,
		rdb,
	)
	masterHandler := master.NewHandler(masterSvc)

	gupRepo := gup.NewRepository(db)
	gupSvc := gup.NewService(gupRepo)
	gupHandler := gup.NewHandler(gupSvc)

	gupReportRepo := gupreport.NewRepository(db)
	gupReportSvc := gupreport.NewService(gupReportRepo)
	gupReportHandler := gupreport.NewHandler(gupReportSvc)

	gupLsRepo := gupls.NewRepository(db)
	gupLsSvc := gupls.NewService(gupLsRepo)
	gupLsHandler := gupls.NewHandler(gupLsSvc)

	dalkotRepo := dalkot.NewRepository(db)
	dalkotSvc := dalkot.NewService(dalkotRepo)
	dalkotHandler := dalkot.NewHandler(dalkotSvc)

	recordRepo := record.NewRepository(db)
	recordSvc := record.NewService(
		recordRepo,
		cfg,
		rdb,
	)
	recordHandler := record.NewHandler(
		recordSvc,
		userRepo,
		cfg,
		masterSvc,
	)

	authSvc := auth.NewService(
		userRepo,
		cfg,
		rdb,
	)
	authHandler := auth.NewHandler(authSvc)

	if err := os.MkdirAll(
		"uploads",
		0o755,
	); err != nil {
		sugar.Warnf(
			"Failed to create uploads directory: %v",
			err,
		)
	}

	if err := os.MkdirAll(
		"uploads/dalkot",
		0o755,
	); err != nil {
		sugar.Warnf(
			"Failed to create Dalkot uploads directory: %v",
			err,
		)
	}

	e := echo.New()

	e.HideBanner = true
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.BodyLimit("50M"))

	e.Static(
		"/uploads",
		"uploads",
	)

	api := e.Group("/api/v1")

	api.POST(
		"/auth/login",
		authHandler.Login,
	)
	api.POST(
		"/auth/register",
		authHandler.Register,
	)
	api.GET(
		"/auth/demo-users",
		authHandler.GetDemoUsers,
	)

	protected := api.Group("")

	protected.Use(
		middleware.JWTMiddleware(
			cfg,
			userRepo,
		),
	)

	protected.GET(
		"/auth/me",
		authHandler.GetMe,
	)
	protected.PUT(
		"/auth/change-password",
		authHandler.ChangePassword,
	)

	protected.POST(
		"/upload",
		recordHandler.UploadFile,
	)

	protected.GET(
		"/dashboard/summary",
		recordHandler.GetDashboardSummary,
	)

	protected.GET(
		"/records",
		recordHandler.GetRecords,
	)
	protected.GET(
		"/records/paginated",
		recordHandler.GetPaginatedRecords,
	)
	protected.POST(
		"/records",
		recordHandler.CreateRecord,
	)
	protected.GET(
		"/records/:id",
		recordHandler.GetRecordByID,
	)
	protected.GET(
		"/records/:id/spd-stream",
		recordHandler.ExportSpdPDF,
	)
	protected.GET(
		"/records/:id/spd-docx",
		recordHandler.ExportSpdDocx,
	)
	protected.GET(
		"/records/:id/laporan-stream",
		recordHandler.ExportLaporanPDF,
	)
	protected.GET(
		"/records/:id/laporan-docx",
		recordHandler.ExportLaporanDocx,
	)
	protected.GET(
		"/records/:id/rincian-stream",
		recordHandler.ExportRincianPDF,
	)
	protected.GET(
		"/records/:id/rincian-docx",
		recordHandler.ExportRincianDocx,
	)
	protected.PUT(
		"/records/:id",
		recordHandler.UpdateRecord,
	)
	protected.DELETE(
		"/records/:id",
		recordHandler.DeleteRecord,
	)
	protected.DELETE(
		"/records/spd/:spd",
		recordHandler.DeleteRecordsBySpd,
	)
	protected.POST(
		"/records/import",
		recordHandler.ImportExcel,
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)

	protected.GET(
		"/users",
		userHandler.GetUsers,
	)
	protected.POST(
		"/users",
		userHandler.CreateUser,
		middleware.RoleMiddleware(
			"super_admin",
		),
	)
	protected.PUT(
		"/users/:id",
		userHandler.UpdateUser,
		middleware.RoleMiddleware(
			"super_admin",
		),
	)
	protected.DELETE(
		"/users/:id",
		userHandler.DeleteUser,
		middleware.RoleMiddleware(
			"super_admin",
		),
	)

	protected.GET(
		"/master/provinces",
		masterHandler.GetProvinces,
	)
	protected.GET(
		"/master/sbm-rates",
		masterHandler.GetSBMRates,
	)
	protected.GET(
		"/master/settings",
		masterHandler.GetSettings,
	)
	protected.PUT(
		"/master/settings",
		masterHandler.UpdateSettings,
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)

	gupGroup := protected.Group(
		"/gup",
	)

	gupGroup.Use(
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)

	gupGroup.GET(
		"",
		gupHandler.List,
	)
	gupGroup.GET(
		"/master",
		gupHandler.GetMaster,
	)
	gupGroup.GET(
		"/:id",
		gupHandler.GetByID,
	)
	gupGroup.POST(
		"",
		gupHandler.Create,
	)
	gupGroup.PUT(
		"/:id",
		gupHandler.Update,
	)
	gupGroup.DELETE(
		"/:id",
		gupHandler.Delete,
	)

	protected.GET(
		"/gup-report",
		gupReportHandler.GetReport,
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)
	protected.PUT(
		"/gup-report/budget",
		gupReportHandler.SaveBudget,
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)

	protected.GET(
		"/gup-ls",
		gupLsHandler.Get,
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)
	protected.PUT(
		"/gup-ls",
		gupLsHandler.Save,
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
		),
	)

	dalkotGroup := protected.Group(
		"/dalkot",
	)

	dalkotGroup.Use(
		middleware.RoleMiddleware(
			"super_admin",
			"kasubag",
			"protokol",
		),
	)

	dalkotGroup.GET(
		"",
		dalkotHandler.List,
	)
	dalkotGroup.GET(
		"/master",
		dalkotHandler.GetMaster,
	)
	dalkotGroup.GET(
	"/recap",
	dalkotHandler.GetRecap,
	middleware.RoleMiddleware(
		"super_admin",
		"kasubag",
	),
)

	dalkotGroup.GET(
	"/recap/monthly-spj",
	dalkotHandler.GetMonthlySPJRecap,
	middleware.RoleMiddleware(
		"super_admin",
		"kasubag",
	),
)
	dalkotGroup.GET(
		"/:id",
		dalkotHandler.GetByID,
	)
	dalkotGroup.POST(
		"",
		dalkotHandler.Create,
	)
	dalkotGroup.PUT(
		"/:id",
		dalkotHandler.Update,
	)
	dalkotGroup.DELETE(
		"/:id",
		dalkotHandler.Delete,
	)

	go func() {
		sugar.Infof(
			"API server starting on port %s",
			cfg.App.Port,
		)

		if err := e.Start(
			":" + cfg.App.Port,
		); err != nil &&
			err != http.ErrServerClosed {
			sugar.Fatalf(
				"API server failed: %v",
				err,
			)
		}
	}()

	go func() {
		sugar.Info(
			"Starting pprof on port 6060 for internal memory telemetry",
		)

		if err := http.ListenAndServe(
			"0.0.0.0:6060",
			nil,
		); err != nil &&
			err != http.ErrServerClosed {
			sugar.Warnf(
				"pprof server failed: %v",
				err,
			)
		}
	}()

	quit := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	receivedSignal := <-quit

	sugar.Infof(
		"Shutdown signal received: %s",
		receivedSignal.String(),
	)

	shutdownContext, cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	defer cancel()

	if err := e.Shutdown(
		shutdownContext,
	); err != nil {
		sugar.Errorf(
			"Graceful shutdown failed: %v",
			err,
		)

		if closeErr := e.Close(); closeErr != nil {
			sugar.Errorf(
				"Forced server close failed: %v",
				closeErr,
			)
		}
	}

	sugar.Info("Server exited properly.")
}
