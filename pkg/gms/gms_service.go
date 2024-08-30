// package gms has all functionality related to good morning sunshine
package gms

import (
	"context"
	"fmt"
	dbpkg "gms/pkg/db"
	"gms/pkg/email"
	emailmodel "gms/pkg/email/model"
	"gms/pkg/gms/model"
	logs "gms/pkg/logger"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
)

// EmailSendJob sends runs every minute to check if there is any mail to be sent, if the mail needs to be sent, then it picks it up and sends the email
func GoodMrngSunshineJob(ctx context.Context) {
	l := logs.GetLoggerctx(ctx)

	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), viper.GetInt("gms.mailjobTimer.hour"), viper.GetInt("gms.mailjobTimer.minute"), viper.GetInt("gms.mailjobTimer.second"), 0, now.Location())
	if now.After(nextRun) { // if we already crossed our ticker time then we try  on the next day
		nextRun = nextRun.Add(24 * time.Hour)
	}
	initialDelay := nextRun.Sub(now) // This is the amount of time we need to wait for the ticker to start firing

	// Create a ticker that fires daily
	ticker := time.NewTicker(24 * time.Hour)

	time.Sleep(initialDelay) // Wait for the initial delay
	_ = goodMorningSunshine(ctx)
	go func() {
		for {
			select {
			case <-ticker.C:
				defer ticker.Stop()
				l.Sugar().Info(fmt.Sprintf("the gms job starts at: %v", time.Now()))
				err := goodMorningSunshine(ctx)
				if err != nil {
					continue
				}
				l.Sugar().Info(fmt.Sprintf("the gms job ends at: %v", time.Now()))
			case <-ctx.Done():
				return
			}
		}
	}()
	// Run indefinitely
	select {}
}

func goodMorningSunshine(ctx context.Context) error {
	maxdays := viper.GetInt("gms.maxdays")

	//send mail to non expired mail ID's
	activeRecords, err := ListActiveEmailIDs(ctx)
	if err != nil {
		return err
	}
	for _, ar := range activeRecords {
		//randomly pick a template for that day
		randomIndex := rand.Intn(maxdays) // generate a random index between 1 and n
		emailbody := email.GetEmailTemplate(randomIndex)
		_ = email.SendEmailUsingSMTP(ctx, &emailmodel.SMTP{
			ToAddress: ar.EmailID,
			EmailBody: emailbody,
			Subject:   "This is Your Message of the Day from team Good Moring Sunshine",
		})

	}

	//Soft Delete expired records
	err = SoftDeleteExpiredEmailIDs(ctx)
	return err
}

/*******************************DATABASE *******************************************/

func EmailRecordTable(ctx context.Context) error {
	l := logs.GetLoggerctx(ctx)

	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.SCHEMA)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		l.Sugar().Errorf("email record table creation failed", err)
		return err
	}

	return nil
}

func EmailRecord(ctx context.Context, mailRecord *model.EmailRecord) error {
	l := logs.GetLoggerctx(ctx)

	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.CREATE_EMAIL_RECORD_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(mailRecord.ID, mailRecord.EmailID, mailRecord.ExpiryDate, mailRecord.CreatedOn, mailRecord.IsDeleted)
	if err != nil {
		l.Sugar().Errorf("email record table creation failed", err)
		return err
	}

	return nil
}

// ListActiveEmailIDs Lists all the email id's which are not expired and are in the mailing list
func ListActiveEmailIDs(ctx context.Context) ([]*model.EmailRecord, error) {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.LIST_ACTIVE_EMAIL_RECORD_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return nil, err
	}
	defer stmt.Close()

	dbRecords, err := stmt.Query(time.Now())
	if err != nil {
		l.Sugar().Errorf("list active email ids failed", err)
		return nil, err
	}
	defer dbRecords.Close()

	items := []*model.EmailRecord{}
	for dbRecords.Next() {
		var i model.EmailRecord
		if err := dbRecords.Scan(
			&i.ID,
			&i.EmailID,
			&i.ExpiryDate,
			&i.ExpiryDate,
			&i.CreatedOn,
			&i.IsDeleted,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}

	if err := dbRecords.Close(); err != nil {
		return nil, err
	}
	if err := dbRecords.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// SoftDeleteExpiredEmailIDs expires (soft delete ) email ID's beyond the expiry date
func SoftDeleteExpiredEmailIDs(ctx context.Context) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.SOFT_DELETE_EXPIRED_RECORD_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Query(time.Now())
	if err != nil {
		l.Sugar().Errorf("soft delete expired email record failed", err)
		return err
	}

	return nil
}

// HardDeleteExpiredEmailIDs delete's the email ID's beyond the expiry date
func HardDeleteExpiredEmailIDs(ctx context.Context, thresholdTime time.Time) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}

	stmt, err := db.Prepare(dbpkg.LIST_ACTIVE_EMAIL_RECORD_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}

	_, err = stmt.Query(thresholdTime)
	if err != nil {
		l.Sugar().Errorf("Hard delete expired email id's failed", err)
		return err
	}

	return nil
}

/************************************************************************************/
