package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"kota-siaga/infrastructure/database"
	"kota-siaga/internal/integrations/apiindonesia"
	"kota-siaga/internal/router"
	"kota-siaga/pkg/config"
	"kota-siaga/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func loadEnvFile(path string) error {
	err := godotenv.Load(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("load %s: %w", path, err)
}
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}
	gin.SetMode(os.Getenv("GIN_MODE"))

	if timeZone, err := time.LoadLocation("Asia/Jakarta"); err != nil {
		logger.WriteLog(logger.LogLevelError, "time.LoadLocation - Error: "+err.Error())
	} else {
		time.Local = timeZone
	}

	myAddr := "unknown"
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			myAddr = ipNet.IP.String()
			break
		}
	}
	if len(myAddr) < 15 {
		myAddr += strings.Repeat(" ", 15-len(myAddr))
	}
	os.Setenv("ServerIP", myAddr)
	logger.WriteLog(logger.LogLevelInfo, "Server IP: "+myAddr)

	var port, appName string
	flag.StringVar(&port, "port", os.Getenv("PORT"), "port of the service")
	flag.StringVar(&appName, "appname", os.Getenv("APP_NAME"), "service name")
	flag.Parse()
	logger.WriteLog(logger.LogLevelInfo, "APP: "+appName+"; PORT: "+port)

	FailOnError(config.ValidateStartupConfig(port), "Invalid app configuration")

	apiClient, err := apiindonesia.NewClient(config.LoadAPIIndonesiaConfig())
	FailOnError(err, "Invalid API Indonesia configuration")

	redisClient, err := database.InitRedis()
	if err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Redis unavailable; continuing without optional Redis features")
	} else {
		defer func() {
			if closeErr := redisClient.Close(); closeErr != nil {
				logger.WriteLog(logger.LogLevelError, "Failed to close Redis connection: "+closeErr.Error())
			}
		}()
		logger.WriteLog(logger.LogLevelInfo, "Redis initialized")
	}

	routes := router.NewRoutes(redisClient, apiClient)
	logger.WriteLog(logger.LogLevelInfo, "Public routes registered")

	FailOnError(routes.App.Run(fmt.Sprintf(":%s", port)), "Failed to run service")
}
