// package gms has all functionality related to good morning sunshine
package gms

import (
	"context"
	"fmt"
	"gms/pkg/auth"
	dbpkg "gms/pkg/db"
	"gms/pkg/email"
	emailmodel "gms/pkg/email/model"
	"gms/pkg/gms/model"
	logs "gms/pkg/logger"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
)

// GoodMrngSunshineJob sends runs  once in a day to check if there is any mail to be sent, if the mail needs to be sent, then it picks it up and sends the email
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
	l := logs.GetLoggerctx(ctx)
	//send mail to non expired mail ID's
	activeRecords, err := ListActiveEmailIDs(ctx)
	if err != nil {
		return err
	}
	for _, ar := range activeRecords {
		randmap := make(map[int64]bool)
		temp := ""
		for _, rn := range ar.RandomNumbers {
			if rn == ',' {
				num, err := strconv.ParseInt(temp, 10, 64) // base 10 and 64-bit integer
				if err != nil {
					l.Sugar().Errorf("error converting string to int", err)
					return err
				}
				randmap[num] = true
				temp = ""
			} else {
				temp += string(rn)
			}
		}

		if len(temp) > 0 {
			num, err := strconv.ParseInt(temp, 10, 64) // base 10 and 64-bit integer
			if err != nil {
				l.Sugar().Errorf("error converting string to int", err)
				return err
			}
			randmap[num] = true
			temp = ""
		}

		//randomly pick a template for that day that index shouldnt be used before
		var randomIndex int64
		for {
			randomIndex = int64(rand.Intn(maxdays)) // generate a random index between 1 and n
			_, ok := randmap[randomIndex]
			if !ok {
				break
			}
		}
		if ar.RandomNumbers == "" {
			ar.RandomNumbers += fmt.Sprintf("%d", randomIndex)
		} else {
			ar.RandomNumbers += "," + fmt.Sprintf("%d", randomIndex)
		}
		err = UpdateEmailRecRandNumber(ctx, ar.ID, ar.RandomNumbers)
		if err != nil {
			continue
		}

		emailbody := email.GetEmailTemplate(randomIndex)
		_ = email.SendEmailUsingGmailSMTP(ctx, &emailmodel.SMTP{
			ToAddress: ar.EmailID,
			EmailBody: emailbody,
			Subject:   "Your Daily Dose of Sunshine from Good Morning Sunshine",
		})

	}

	//Soft Delete expired records
	err = SoftDeleteExpiredEmailIDs(ctx)
	return err
}

// Once the user sign's up this function is called
func MainPageEntry(ctx context.Context, emailID string) error {
	l := logs.GetLoggerctx(ctx)
	//create owner table if not exists
	err := OwnerTable(ctx)
	if err != nil {
		return err
	}

	ownerRecord, err := GetOwnerRecordByEmailID(ctx, emailID)
	if err != nil {
		return err
	}
	if ownerRecord == nil { //creating for the first time
		err = CreateOwnerRecord(ctx, &model.OwnerRecord{
			ID:        uuid.New(),
			EmailID:   emailID,
			RateLimit: 1,
		})
		if err != nil {
			return err
		}
	} else {
		if ownerRecord.RateLimit >= 3 {
			l.Sugar().Info("At max a user " + emailID + " can create 3 new records")
			return fmt.Errorf("At max a user " + emailID + " can create 3 new records")
		} else {
			//update the rate limit count in db
			err = UpdateOwnerRateLimit(ctx, emailID, ownerRecord.RateLimit+1)
			if err != nil {
				return err
			}
		}
	}

	err = emailMainPage(ctx, emailID)
	if err != nil {
		return err
	}

	return nil
}

func emailMainPage(ctx context.Context, emailID string) error {

	url, err := mainPageurl(ctx, emailID)
	if err != nil {
		return err
	}
	err = email.SendEmailUsingGmailSMTP(ctx, &emailmodel.SMTP{
		ToAddress: emailID,
		EmailBody: `<html>
		<body>
		Thank you for joining Good Morning Sunshine. We're delighted to have you on board. To begin sharing morning greetings with your chosen recipient, please use the secure link below:
		<br>
		 <a href="` + url + `">` + url + `</a>
				</body>
		</html>
		`,
		Subject: "Rise & Shine: Your Good Morning Sunshine Link Inside!",
	})
	if err != nil {
		return err
	}
	return nil
}

// mainPageurl creates a new jwt token with emailID wrapped into it and attaches the jwt with the url and sends the mail
// This acts as a authentication to authorize only those users who has entered the main page url through their mail
func mainPageurl(ctx context.Context, emailID string) (string, error) {
	l := logs.GetLoggerctx(ctx)

	jwtToken, err := auth.CreateJWTToken(emailID)
	if err != nil {
		l.Sugar().Errorf("creating a new jwt token failed", err)
		return "", err
	}
	//attach this to the url
	baseurl := viper.GetString("app.mailPageurl")
	u, err := url.Parse(baseurl) //parses the url into URL structure
	if err != nil {
		l.Sugar().Errorf("error parsing base url", err)
		return "", err
	}
	//adding a jwt query parameter
	q := u.Query()
	q.Add("tkn", jwtToken)  //tkn is jwt token(key)
	u.RawQuery = q.Encode() //Encode encodes the values into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
	mailPageurl := u.String()
	return mailPageurl, nil
}

func ListMainPage(ctx context.Context, emailID string) ([]*model.EmailRecord, error) {

	err := EmailRecordTable(ctx)
	if err != nil {
		return nil, err
	}
	emailRecords, err := ListEmailRecordByOwnerMailID(ctx, emailID)
	if err != nil {
		return nil, err
	}

	return emailRecords, nil
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

	_, err = stmt.Exec(mailRecord.ID, mailRecord.EmailID, mailRecord.OwnerMailID, mailRecord.ExpiryDate, "", mailRecord.CreatedOn, mailRecord.IsDeleted)
	if err != nil {
		l.Sugar().Errorf("email record table creation failed", err)
		return err
	}

	return nil
}

func UpdateEmailRecRandNumber(ctx context.Context, id uuid.UUID, randstring string) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.UPDATE_EMAIL_RECORD_RANDNUM_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(randstring, id)
	if err != nil {
		l.Sugar().Errorf("update email record random number string failed", err)
		return err
	}

	return nil
}

func OwnerTable(ctx context.Context) error {
	l := logs.GetLoggerctx(ctx)

	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.OWNER_SCHEMA)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		l.Sugar().Errorf("owner table creation failed", err)
		return err
	}

	return nil
}

func CreateOwnerRecord(ctx context.Context, mailRecord *model.OwnerRecord) error {
	l := logs.GetLoggerctx(ctx)

	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.CREATE_OWNER_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(mailRecord.ID, mailRecord.EmailID, mailRecord.RateLimit, time.Now(), time.Now(), false)
	if err != nil {
		l.Sugar().Errorf("owner db record creation failed", err)
		return err
	}

	return nil
}

func UpdateOwnerRateLimit(ctx context.Context, email_id string, rate_limit int) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.UPDATE_OWNER_RATE_LIMIT_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(rate_limit, time.Now(), email_id)
	if err != nil {
		l.Sugar().Errorf("update owner rate limit failed", err)
		return err
	}

	return nil
}

func GetOwnerRecordByEmailID(ctx context.Context, emailID string) (*model.OwnerRecord, error) {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.GET_OWNER_DETAILS_BY_EMAILID_QUERY)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return nil, err
	}
	defer stmt.Close()

	dbRecords, err := stmt.Query(emailID)
	if err != nil {
		l.Sugar().Errorf("get owner record by emailID failed", err)
		return nil, err
	}
	defer dbRecords.Close()

	i := model.OwnerRecord{}
	createdOn := ""
	updatedOn := ""
	for dbRecords.Next() {
		if err := dbRecords.Scan(
			&i.ID,
			&i.EmailID,
			&i.RateLimit,
			&createdOn,
			&updatedOn,
			&i.IsDeleted,
		); err != nil {
			l.Sugar().Errorf("scan records failed", err)
			return nil, err
		}
	}

	if err := dbRecords.Close(); err != nil {
		l.Sugar().Errorf("db close failed", err)
		return nil, err
	}
	if err := dbRecords.Err(); err != nil {
		l.Sugar().Errorf("error in db records", err)
		return nil, err
	}

	if i.ID == uuid.Nil && i.EmailID == "" && i.RateLimit == 0 { // there was no record
		return nil, nil
	}

	// Define the layout (format) of the time string
	layout := "2006-01-02 15:04:05.999999-07:00"

	// Parse the time string to time.Time object
	ct, err := time.Parse(layout, createdOn)
	if err != nil {
		l.Sugar().Errorf("error parsing created on time", err)
		return nil, err
	}

	ut, err := time.Parse(layout, updatedOn)
	if err != nil {
		l.Sugar().Errorf("error parsing updated on time", err)
		return nil, err
	}
	i.CreatedOn = ct
	i.UpdatedOn = ut

	return &i, nil
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
		expiryDate := ""
		createdOn := ""
		if err := dbRecords.Scan(
			&i.ID,
			&i.EmailID,
			&i.OwnerMailID,
			&expiryDate,
			&i.RandomNumbers,
			&createdOn,
			&i.IsDeleted,
		); err != nil {
			l.Sugar().Errorf("scan records failed", err)
			return nil, err
		}
		// Define the layout (format) of the time string
		layout := "2006-01-02 15:04:05.999999-07:00"

		// Parse the time string to time.Time object
		ct, err := time.Parse(layout, createdOn)
		if err != nil {
			l.Sugar().Errorf("error parsing created on time", err)
			return nil, err
		}

		ed, err := time.Parse(layout, expiryDate)
		if err != nil {
			l.Sugar().Errorf("error parsing updated on time", err)
			return nil, err
		}
		i.CreatedOn = ct
		i.ExpiryDate = ed

		items = append(items, &i)
	}

	if err := dbRecords.Close(); err != nil {
		l.Sugar().Errorf("db close failed", err)
		return nil, err
	}
	if err := dbRecords.Err(); err != nil {
		l.Sugar().Errorf("db record error", err)
		return nil, err
	}
	return items, nil
}

// ListEmailRecordByOwnerMailID Lists all the email record for a owner mailID
func ListEmailRecordByOwnerMailID(ctx context.Context, emailID string) ([]*model.EmailRecord, error) {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.LIST_ALL_EMAIL_RECORD_FOR_A_OWNER)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return nil, err
	}
	defer stmt.Close()

	dbRecords, err := stmt.Query(emailID)
	if err != nil {
		l.Sugar().Errorf("ListEmailRecordByOwnerMailID failed", err)
		return nil, err
	}
	defer dbRecords.Close()

	items := []*model.EmailRecord{}
	expiryDate := ""
	createdOn := ""
	for dbRecords.Next() {
		var i model.EmailRecord
		if err := dbRecords.Scan(
			&i.ID,
			&i.EmailID,
			&i.OwnerMailID,
			&expiryDate,
			&i.RandomNumbers,
			&createdOn,
			&i.IsDeleted,
		); err != nil {
			l.Sugar().Errorf("scan records failed", err)
			return nil, err
		}
		// Define the layout (format) of the time string
		layout := "2006-01-02 15:04:05.999999-07:00"

		// Parse the time string to time.Time object
		ct, err := time.Parse(layout, createdOn)
		if err != nil {
			l.Sugar().Errorf("error parsing created on time", err)
			return nil, err
		}

		ed, err := time.Parse(layout, expiryDate)
		if err != nil {
			l.Sugar().Errorf("error parsing updated on time", err)
			return nil, err
		}
		i.CreatedOn = ct
		i.ExpiryDate = ed

		items = append(items, &i)
	}

	if err := dbRecords.Close(); err != nil {
		l.Sugar().Errorf("db close failed", err)
		return nil, err
	}
	if err := dbRecords.Err(); err != nil {
		l.Sugar().Errorf("db record error", err)
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

	_, err = stmt.Exec(time.Now())
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

	_, err = stmt.Exec(thresholdTime)
	if err != nil {
		l.Sugar().Errorf("Hard delete expired email id's failed", err)
		return err
	}

	return nil
}

func SoftDeleteRecordsByID(ctx context.Context, recordID string) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.SOFT_DELETE_RECORDS_BY_ID)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(recordID)
	if err != nil {
		l.Sugar().Errorf("soft delete record by record id failed", err)
		return err
	}

	return nil
}

/************************************************************************************/
