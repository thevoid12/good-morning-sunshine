package handlers

import (
	"gms/pkg/auth"
	"gms/pkg/gms"
	"gms/pkg/gms/model"
	logs "gms/pkg/logger"
	"html/template"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type CheckMail struct {
	Email string `json:"email" validate:"required,email`
}
type DeactivateRequest struct {
	RecordID string `validate:"required,uuid"`
}

type MainPage struct {
	AuthToken string
	EmailMeta []*EmailMeta
}

type EmailMeta struct {
	RecordID      uuid.UUID
	EmailID       string
	IsExpired     bool
	DaysRemaining int
}

func HomeHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "landing_page.html"))
	if err != nil {
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		l.Sugar().Errorf("execute template failed", err)
		return
	}
}

func CheckMailHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "check_mail.html"))
	if err != nil {
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	err = c.Request.ParseForm()
	if err != nil {
		l.Sugar().Errorf("parse form failed", err)
		return
	}

	emailID := c.PostForm("emailaddress")
	cm := CheckMail{
		Email: emailID,
	}
	validate := validator.New()
	err = validate.Struct(cm)
	if err != nil {
		l.Sugar().Errorf("the entered input is not valid", err)
		return
	}

	//send a mail with the link of the website for him to access
	err = gms.MainPageEntry(ctx, emailID)
	if err != nil {
		l.Sugar().Errorf("initial email send failed", err)
		return
	}
	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		l.Sugar().Errorf("execute template failed", err)
		return
	}
}

// initial load of the MainPage
func MainPageHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	authtoken := c.Query("tkn")
	token, err := auth.VerifyJWTToken(ctx, authtoken)
	if err != nil {
		return
	}
	tokenClaims, err := auth.ExtractClaims(token)
	if err != nil {
		return
	}
	emailID := tokenClaims.EmailID // this is the email id the user has signed up with
	var exp bool
	if tokenClaims.ExpiryDate.Before(time.Now()) {
		exp = true
	}
	emailRecords, err := gms.ListMainPage(ctx, emailID)
	if err != nil {
		return
	}

	d := MainPage{
		AuthToken: authtoken,
		EmailMeta: []*EmailMeta{},
	}
	for _, er := range emailRecords {
		exp = false
		if er.ExpiryDate.Before(time.Now()) || er.IsDeleted {
			exp = true
		}
		daysRem := 0
		if !exp {

			duration := time.Until(er.ExpiryDate)

			// Convert the duration to days
			daysRem = int(duration.Hours()/24) + 1
		}
		d.EmailMeta = append(d.EmailMeta, &EmailMeta{
			RecordID:      er.ID,
			EmailID:       er.EmailID,
			IsExpired:     exp,
			DaysRemaining: daysRem,
		})
	}
	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "mainpage.html"))
	if err != nil {
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, d)
	if err != nil {
		l.Sugar().Errorf("execute template failed", err)
		return
	}
}

// Create a new email record
func NewMailRecordHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	err := c.Request.ParseForm()
	if err != nil {
		l.Sugar().Errorf("parse form failed", err)
		return
	}
	emailID := c.PostForm("emailaddress")
	if emailID == "" {
		l.Sugar().Errorf("email field cannot be empty")
		return
	}

	authtoken := c.Param("tkn")
	token, err := auth.VerifyJWTToken(ctx, authtoken)
	if err != nil {
		return
	}
	tokenClaims, err := auth.ExtractClaims(token)
	if err != nil {
		return
	}
	ownerMailID := tokenClaims.EmailID // this is the email id the user has signed up with

	cm := CheckMail{
		Email: emailID,
	}
	validate := validator.New()
	err = validate.Struct(cm)
	if err != nil {
		l.Sugar().Errorf("the entered input is not valid", err)
		return
	}

	err = gms.EmailRecord(ctx, &model.EmailRecord{
		ID:          uuid.New(),
		EmailID:     emailID,
		OwnerMailID: ownerMailID,
		ExpiryDate:  time.Now().AddDate(0, 0, 7),
		CreatedOn:   time.Now(),
		IsDeleted:   false,
	})
	if err != nil {
		l.Sugar().Errorf("email record creation  failed", err)
		return
	}

	emailRecords, err := gms.ListMainPage(ctx, ownerMailID)
	if err != nil {
		return
	}

	d := MainPage{
		AuthToken: authtoken,
		EmailMeta: []*EmailMeta{},
	}

	for _, er := range emailRecords {
		exp := false
		if er.ExpiryDate.Before(time.Now()) || er.IsDeleted {
			exp = true
		}
		daysRem := 0
		if !exp {

			duration := time.Until(er.ExpiryDate)

			// Convert the duration to days
			daysRem = int(duration.Hours()/24) + 1
		}
		d.EmailMeta = append(d.EmailMeta, &EmailMeta{
			RecordID:      er.ID,
			EmailID:       er.EmailID,
			IsExpired:     exp,
			DaysRemaining: daysRem,
		})
	}
	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "mainpage.html"))
	if err != nil {
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, d)
	if err != nil {
		l.Sugar().Errorf("execute template failed", err)
		return
	}
}

func DeactivateRecordHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	recordID := c.Param("id")
	req := DeactivateRequest{
		RecordID: recordID,
	}

	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		l.Sugar().Errorf("the entered record id is not valid", err)
		return
	}
	authtoken := c.Param("tkn")
	token, err := auth.VerifyJWTToken(ctx, authtoken)
	if err != nil {
		return
	}
	tokenClaims, err := auth.ExtractClaims(token)
	if err != nil {
		return
	}
	ownerMailID := tokenClaims.EmailID // this is the email id the user has signed up with
	err = gms.SoftDeleteRecordsByID(ctx, recordID)
	if err != nil {
		return
	}

	emailRecords, err := gms.ListMainPage(ctx, ownerMailID)
	if err != nil {
		return
	}

	d := MainPage{
		AuthToken: authtoken,
		EmailMeta: []*EmailMeta{},
	}

	for _, er := range emailRecords {
		exp := false
		if er.ExpiryDate.Before(time.Now()) || er.IsDeleted {
			exp = true
		}
		daysRem := 0
		if !exp {

			duration := time.Until(er.ExpiryDate)

			// Convert the duration to days
			daysRem = int(duration.Hours()/24) + 1
		}
		d.EmailMeta = append(d.EmailMeta, &EmailMeta{
			RecordID:      er.ID,
			EmailID:       er.EmailID,
			IsExpired:     exp,
			DaysRemaining: daysRem,
		})
	}
	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "mainpage.html"))
	if err != nil {
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, d)
	if err != nil {
		l.Sugar().Errorf("execute template failed", err)
		return
	}

}
