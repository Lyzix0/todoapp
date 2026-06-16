package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/Lyzix0/todoapp/internal/core/config"
	core_logger "github.com/Lyzix0/todoapp/internal/core/logger"
	core_pgx_pool "github.com/Lyzix0/todoapp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/Lyzix0/todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/Lyzix0/todoapp/internal/core/transport/http/server"
	stats_postres_repository "github.com/Lyzix0/todoapp/internal/features/stats/repository/postgres"
	stats_service "github.com/Lyzix0/todoapp/internal/features/stats/service"
	stats_transport_http "github.com/Lyzix0/todoapp/internal/features/stats/transport/http"
	tasks_postgres_repository "github.com/Lyzix0/todoapp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/Lyzix0/todoapp/internal/features/tasks/service"
	tasks_transport_http "github.com/Lyzix0/todoapp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/Lyzix0/todoapp/internal/features/users/repository/postgres"
	users_service "github.com/Lyzix0/todoapp/internal/features/users/service"
	users_transport_http "github.com/Lyzix0/todoapp/internal/features/users/transport/http"
	web_fs_repository "github.com/Lyzix0/todoapp/internal/features/web/repository/file_system"
	web_service "github.com/Lyzix0/todoapp/internal/features/web/service"
	web_transport_http "github.com/Lyzix0/todoapp/internal/features/web/transport/http"
	"go.uber.org/zap"

	_ "github.com/Lyzix0/todoapp/docs"
)

// @title Golang Todoapp api
// @version 1.0
// @description Todoapp REST-API scheme
// @host 127.0.0.1:5050
// @BasePath /api/v1

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

	logger.Debug("initializing postgres connection pool")

	pool, err := core_pgx_pool.NewConnectionPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))

	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))

	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)
	tasksService := tasks_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing feature", zap.String("feature", "stats"))

	statsRepository := stats_postres_repository.NewStatsRepository(pool)
	statsService := stats_service.NewStatsService(statsRepository)
	statsTransportHTTP := stats_transport_http.NewStatsHTTPHandler(statsService)

	logger.Debug("ininitalizing feature", zap.String("feature", "web"))
	webRepository := web_fs_repository.NewWebRepository()
	webService := web_service.NewWebService(webRepository)
	webTransportHTTP := web_transport_http.NewWebHTTPHandler(webService)

	logger.Debug("initializing HTTP server")

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouterV1.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(statsTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouterV1)
	httpServer.RegisterRoutes(webTransportHTTP.Routes()...)
	httpServer.RegisterSwagger()

	// apiVersionRouterV2 := core_http_server.NewAPIVersionRouter(
	// 	core_http_server.ApiVersion2,
	// 	core_http_middleware.Dummy("api v2 middleware"),
	// )
	// apiVersionRouterV2.RegisterRoutes(usersTransportHTTP.Routes()...)
	// httpServer.RegisterAPIRouters(apiVersionRouterV2)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
