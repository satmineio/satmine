package main

// swag init
// swag init -g cmd/main.go
// go run ./cmd
import (
	"fmt"
	"os"
	docs "satmine/docs"
	"satmine/rpc"
	"satmine/satmine"
	"satmine/store"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @BasePath /api/v1
// @title BTC Ordinals Mining Protocol
// @description Mining Protocol for Inscription Miners Based on the BTC Ordinals Protocol
// @version 1.01
// @termsOfService https://github.com/SatMine/

// json is an alias for jsoniter library configured to be compatible with the standard library's JSON handling.
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// logger is a global variable for the logger instance
var logger *zap.Logger

// Config represents application configuration
type Config struct {
	Port       int
	Socketport int
	Name       string
	Ordrpc     string
	Dbpath     string
	RecPath    string
	Hookrpc    string
}

// AppConfig holds the global configuration
var AppConfig Config

// cleanUpPreviousData removes previous data in the given path
func cleanUpPreviousData(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Path does not exist, no need to clean
	}

	// Remove the directory and its contents
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove previous data: %w", err)
	}

	return nil
}

// CORS is a middleware to handle Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	var err error
	// Initialize the logger
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flush any buffered log entries

	// Log an entry
	logger.Info("Logger initialized successfully")

	//var configErr error
	viper.SetConfigName("config") // Name of config file (without extension)
	viper.SetConfigType("yaml")   // Or viper.SetConfigType("YAML")
	viper.AddConfigPath(".")      // Path to look for the config file in

	if configErr := viper.ReadInConfig(); configErr != nil {
		logger.Error("Error reading config file", zap.Error(configErr))
	}

	// Bind configuration to struct
	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		logger.Error("Unable to decode into struct, %s", zap.Error(err))
	}

	logger.Info("Configuration: %+v\n", zap.Reflect("config", AppConfig))

	// //Debug used Clean up previous data if exists
	// err = cleanUpPreviousData(AppConfig.Dbpath)
	// if err != nil {
	// 	logger.Error("Error cleaning up previous data", zap.Error(err))
	// 	panic(err)
	// }

	// Initialize the db with the specified database path
	badOpts := badger.DefaultOptions(AppConfig.Dbpath)
	badOpts.VerifyValueChecksum = true
	badOpts.ValueLogFileSize = 2<<30 - 1
	badOpts.MemTableSize = 512 << 20
	db, err := badger.Open(badOpts)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//fmt.Println("AppConfig.RecPath = ", AppConfig.RecPath)

	// Create an instance of satmine.BTOrdIdx using the dbManager
	btOrdIdx := satmine.NewBTOrdIdx(db)

	// Indexing of transaction information records
	recOpts := badger.DefaultOptions(AppConfig.RecPath)
	recOpts.VerifyValueChecksum = true
	recOpts.ValueLogFileSize = 2<<30 - 1
	recOpts.MemTableSize = 512 << 20
	recDb, recErr := badger.Open(recOpts)
	if recErr != nil {
		panic(recErr)
	}
	defer recDb.Close()
	btRecIdx := satmine.NewBTRecIdx(recDb)

	// Retrieve the singleton instance of store.Store and initialize it
	store := store.Instance()
	store.Init(btOrdIdx, btRecIdx, fmt.Sprint(AppConfig.Socketport))

	r := gin.Default()
	r.Use(CORS())

	docs.SwaggerInfo.BasePath = "/api/v1"

	//Registering a Route
	rpc.RegisterRoutes(r)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Static("/assets", "./assets")
	r.Run(":" + strconv.Itoa(AppConfig.Port))
}
