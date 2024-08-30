package main

import (
	"context"
	"gms/client/routes"

	//"gms/pkg/email"

	logs "gms/pkg/logger"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	err := godotenv.Load()
	//email.SendEmailUsingSMTP()
	// email.SendEmail()
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

	l.Sugar().Info("this is a test logger")
	// err = errors.New("this is a generated error")
	// 	l.Sugar().Errorf("this is a big error", err)
	ctx := context.Background()
	ctx = logs.SetLoggerctx(ctx, l)

	route := routes.Initialize(ctx, l)
	route.Run(":" + viper.GetString("app.port"))
}
