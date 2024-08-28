// package gms has all functionality related to good morning sunshine
package gms

import (
	"context"
	"fmt"
	dbpkg "gms/pkg/db"
	"gms/pkg/gms/model"
	logs "gms/pkg/logger"

	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
)

// EmailSendJob sends runs every minute to check if there is any mail to be sent, if the mail needs to be sent, then it picks it up and sends the email
func GoodMrngSunshine() {

	maxdays := viper.GetInt("gms.maxdays")

	//randomly pick a template for a week
	randomIndex := rand.Intn(maxdays) // generate a random index between 1 and n
	fmt.Println(randomIndex)

}

func CreateNewEmailRecord() {

	db, err := dbpkg.NewdbConnection()
	if err != nil {

	}
	stmt, err := db.Prepare(dbpkg.CREATE_EMAIL_RECORD_QUERY)
	if err != nil {

	}
	stmt.Exec()

}

func CreateEmailRecordTable(ctx context.Context, mailRecord *model.CreateEmailRecord) {
	l := logs.GetLoggerctx(ctx)

	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
	}

	stmt, err := db.Prepare(dbpkg.SCHEMA)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
	}

	_, err = stmt.Exec(mailRecord)
	if err != nil {
		l.Sugar().Errorf("create email record table creation failed", err)
	}

}
