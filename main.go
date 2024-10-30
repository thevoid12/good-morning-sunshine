package main

import (
	"context"
	"gms/client/routes"

	dbpkg "gms/pkg/db"
	"gms/pkg/gms"
	logs "gms/pkg/logger"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("there is a error loading environment variables", err)
		return
	}
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("config/") // path to look for the config file in

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("there is a error in the path of config file", err)
		} else {
			// Config file was found but another error was produced
			log.Println("error laoding config file from viper", err)
		}
	}

	//gms.GoodMrngSunshine()
	l, err := logs.InitializeLogger()
	if err != nil {
		log.Println("error initializing logger", err)
	}
	c := dbpkg.NewCache()
	ctx := context.Background()
	ctx = logs.SetLoggerctx(ctx, l)
	ctx = dbpkg.SetCachectx(ctx, c)

	err = InitializeCache(ctx, c)
	if err != nil {
		l.Sugar().Error("cache initialization failed", err)
		return
	}
	l.Sugar().Info("cache initialized successfully")

	go gms.GoodMrngSunshineJob(ctx) //a go routine to run sending mail job forever

	route := routes.Initialize(ctx, l)
	route.Run(":" + viper.GetString("app.port"))
}

func InitializeCache(ctx context.Context, cache *dbpkg.Cache) error {
	//get all the records
	//load the valid values in the cache
	emailRecords, err := gms.ListActiveEmailIDs(ctx)
	if err != nil {
		return err
	}

	for _, record := range emailRecords {
		mailTime, err := gms.ConvertMailTime(record.TimeZone)
		if err != nil {
			return err
		}
		//cache key is the time and value is the array of details
		cache.Set(mailTime.Format("15:04"), &dbpkg.CacheEntry{
			RecordID:      record.ID,
			EmailID:       record.EmailID,
			RandomNumbers: record.RandomNumbers,
			ExpiryDate:    record.ExpiryDate,
		})

	}

	return nil
}
