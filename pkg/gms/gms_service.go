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
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

func isTimeSynced() (bool, error) {
	// Check if the system clock has been synchronized using timedatectl this is an ideal check for my rasp pi which can have interupted power supply
	out, err := exec.Command("timedatectl").Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), "System clock synchronized: yes"), nil
}

// GoodMrngSunshineJob sends runs every minute to check if there is any mail to be sent, if the mail needs to be sent, then it picks it up and sends the email
func GoodMrngSunshineJob(ctx context.Context) {
	l := logs.GetLoggerctx(ctx)
	l.Sugar().Info("welcome to good morning sunshine job", time.Now())
	// Wait for time synchronization
	for {
		isTimeSync, err := isTimeSynced()
		if err != nil {
			l.Sugar().Error("Error checking time synchronization:", err)
			return
		}

		if isTimeSync {
			l.Sugar().Info("Time is in perfect sync")
			break // Exit the loop if time is synchronized
		}

		l.Sugar().Info("Waiting for time synchronization...")
		time.Sleep(5 * time.Second) // Retry every 5 seconds until time is synced
	}

	calcNextRunTime := func() time.Time {
		now := time.Now()
		// Round to the next minute
		nextRun := now.Truncate(time.Minute).Add(time.Minute)

		return nextRun
	}

	// Calculate the next run time after sync
	nextRun := calcNextRunTime()
	l.Sugar().Info(fmt.Sprintf("Next run scheduled at: %v", nextRun))

	// Initial delay until the next run at 6 AM
	initialDelay := time.Until(nextRun)

	// Wait until the first scheduled run
	time.Sleep(initialDelay)
	ticker := time.NewTicker(1 * time.Minute) // ticks every minute
	l.Sugar().Info(fmt.Sprintf("First job triggered at: %v", time.Now()))
	cache := dbpkg.GetCacheFromctx(ctx)
	go goodMorningSunshine(ctx, cache) // Run the first job

	// Schedule the job to run every minute
	go func() {
		for {
			select {
			case <-ticker.C:
				//l.Sugar().Info(fmt.Sprintf("The GMS job starts at: %v", time.Now()))
				err := goodMorningSunshine(ctx, cache)
				if err != nil {
					l.Sugar().Error("Error running goodMorningSunshine:", err)
					continue
				}
			//	l.Sugar().Info(fmt.Sprintf("The GMS job ends at: %v", time.Now()))

			case <-ctx.Done():
				ticker.Stop()
				l.Sugar().Info("GMS job stopped.")
				return
			}
		}
	}()

	// Run indefinitely
	select {}
}

func goodMorningSunshine(ctx context.Context, cache *dbpkg.Cache) error {
	maxdays := viper.GetInt("gms.maxdays")
	l := logs.GetLoggerctx(ctx)
	//send mail to non expired mail ID's
	curtime := time.Now().Format("15:04")
	Records := cache.Get(curtime)
	if Records == nil {
		return nil
	}

	randmap := make(map[int64]bool)
	temp := ""
	activeRecords := []*dbpkg.CacheEntry{}

	for _, ar := range Records {
		if ar.ExpiryDate.Before(time.Now()) { //remove the record from the cache
			continue
		}

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

		err := UpdateEmailRecRandNumber(ctx, ar.RecordID, ar.RandomNumbers)
		if err != nil {
			continue
		}

		emailbody := email.GetEmailTemplate(randomIndex)
		_ = email.SendEmailUsingGmailSMTP(ctx, &emailmodel.SMTP{
			ToAddress: ar.EmailID,
			EmailBody: emailbody,
			Subject:   "Your Daily Dose of Sunshine from Good Morning Sunshine",
		})

		activeRecords = append(activeRecords, ar)
	}
	//update the cache with the removed records
	if len(activeRecords) == 0 {
		cache.Delete(curtime)
	} else {
		cache.Update(curtime, activeRecords)
	}

	//Soft Delete expired records
	err := SoftDeleteExpiredEmailIDs(ctx)
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
	go email.SendEmailUsingGmailSMTP(ctx, &emailmodel.SMTP{ // it is a go routine as it takes some tike to send mail and i dont have to wait until it finishes
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

// convert to 6am to their timezone to ist time
func ConvertMailTime(tz string) (time.Time, error) {
	// Load the source timezone
	srcLocation, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, fmt.Errorf("error loading source timezone: %v", err)
	}

	// Load IST timezone
	istLocation, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.Time{}, fmt.Errorf("error loading IST timezone: %v", err)
	}
	now := time.Now()
	srcTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		viper.GetInt("gms.mailjobTimer.hour"),   // 6
		viper.GetInt("gms.mailjobTimer.minute"), // 0
		viper.GetInt("gms.mailjobTimer.second"), // 0
		0, srcLocation,
	)
	// // Set a reference date (January 2, 2006, is the canonical reference in Go)
	// srcTime := time.Date(2006, time.January, 2, hour, minute, 0, 0, srcLocation)

	// Convert the time to IST
	istTime := srcTime.In(istLocation)

	utcTime := istTime.UTC()
	return utcTime, nil
}

func InitializeGmsCache(ctx context.Context, cache *dbpkg.Cache, l *zap.Logger) error {
	//get all the records
	//load the valid values in the cache
	emailRecords, err := ListActiveEmailIDs(ctx)
	if err != nil {
		return err
	}

	for _, record := range emailRecords {
		mailTime, err := ConvertMailTime(record.TimeZone)
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

	cache.CacheStore.Range(func(key, value interface{}) bool {
		entry := value.([]*dbpkg.CacheEntry)
		for _, val := range entry {
			l.Sugar().Info(key, "::::", val.EmailID)
		}
		return true // continue iterating
	})
	return nil
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

	_, err = stmt.Exec(mailRecord.ID, mailRecord.EmailID, mailRecord.OwnerMailID, mailRecord.ExpiryDate, mailRecord.TimeZone, "", mailRecord.CreatedOn, mailRecord.IsDeleted)
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
			&i.TimeZone,
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
			&i.TimeZone,
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
func ToggleActivityStatus(ctx context.Context, recordID string, isActive string) (err error) {
	if isActive == "1" {
		err = SoftDeleteRecordsByID(ctx, recordID)

	} else {
		err = ActivateDeleteRecordsByID(ctx, recordID)

	}
	return err
}

func SoftDeleteRecordsByID(ctx context.Context, recordID string) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.TOGGLE_RECORDS_DELETION_BY_ID)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(true, recordID)
	if err != nil {
		l.Sugar().Errorf("soft delete record by record id failed", err)
		return err
	}

	return nil
}

func ActivateDeleteRecordsByID(ctx context.Context, recordID string) error {
	l := logs.GetLoggerctx(ctx)
	db, err := dbpkg.NewdbConnection()
	if err != nil {
		l.Sugar().Errorf("new db connection creation failed", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(dbpkg.TOGGLE_RECORDS_DELETION_BY_ID)
	if err != nil {
		l.Sugar().Errorf("db prepare failed", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(false, recordID)
	if err != nil {
		l.Sugar().Errorf("activate record by record id failed", err)
		return err
	}

	return nil
}

/************************************************************************************/
