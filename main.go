package main

import (
	"chatie/internal/config"
	"chatie/internal/handlers"
	"chatie/internal/repository"
	"chatie/internal/server"
	"chatie/internal/services"
	"chatie/internal/ws"
	manager "chatie/pkg/auth"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var addr = flag.String("addr", ":3000", "")

var ctx = context.Background()

const (
	publicConfig  = "./configs/config.yaml"
	privateConfig = ".env"
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
		ForceQuote:    true,
	})
	logger.SetLevel(logrus.DebugLevel)

	cfg, err := config.LoadConfigs(publicConfig, privateConfig)
	if err != nil {
		logger.Debug("parsing config error: ", err)
		return
	}
	// cfg.HTTP.Port = *addr
	logger.Debug("configs parsed", cfg.HTTP.Port)

	dsn := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Name)
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Fatal("connection failed: ", err)
	}
	defer dbpool.Close()
	logger.Debug("postgres connected")

	opt, err := redis.ParseURL("redis://localhost:6364/0")
	if err != nil {
		logger.Fatal(err)
	}
	redis := redis.NewClient(opt)
	logger.Debug("redis connected", redis)

	chatRepo := repository.NewChatRepository(dbpool)
	userRepo := repository.NewUserRepository(dbpool)

	hub := ws.NewWsServer(chatRepo, userRepo, redis)
	go hub.Run()
	logger.Debug("websocket server started")

	tokenManager, _ := manager.NewManager(cfg.Auth.SigningKey)

	userSerice := services.NewUserService(userRepo)
	userHandler := handlers.NewUserhandler(userSerice, tokenManager, cfg)

	router := handlers.Routes(userHandler, hub)

	// router.GET("/ws", middleware.AuthUser(tokenManager), func(c *gin.Context) {
	// 	ws.ServeWS(hub, c)
	// })

	server := server.NewServer(&cfg, router)

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		logger.Fatal("Server Shutdown:", err)
	}

	select {
	case <-ctx.Done():
		logger.Println("timeout of 5 seconds.")
	}
	logger.Println("Server exiting")
}
