package handlers

import (
	"fmt"
	constants "gms/constant"
	"gms/pkg/auth"
	dbpkg "gms/pkg/db"
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
	IsActive string `validate:"required"`
}

type MainPage struct {
	AuthToken string
	EmailMeta []*EmailMeta
	Timezone  []string
}

type EmailMeta struct {
	RecordID      uuid.UUID
	EmailID       string
	IsExpired     bool
	DaysRemaining int
	EmailTz       string
}

func HomeHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "landing_page.html"))
	if err != nil {
		RenderErrorTemplate(c, "Internal server error occured", err)
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		RenderErrorTemplate(c, "Internal server error occured", err)
		l.Sugar().Errorf("execute template failed", err)
		return
	}

}

func PremiumHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "premium.html"))
	if err != nil {
		RenderErrorTemplate(c, "Internal server error occured", err)
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		RenderErrorTemplate(c, "Internal server error occured", err)
		l.Sugar().Errorf("execute template failed", err)
		return
	}

}

func CheckMailHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "check_mail.html"))
	if err != nil {
		RenderErrorTemplate(c, "Internal server error occured", err)
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	err = c.Request.ParseForm()
	if err != nil {
		RenderErrorTemplate(c, "Parse form failed", err)
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
		RenderErrorTemplate(c, "Input is not valid", err)
		l.Sugar().Errorf("the entered input is not valid", err)
		return
	}

	//send a mail with the link of the website for him to access
	err = gms.MainPageEntry(ctx, emailID)
	if err != nil {
		RenderErrorTemplate(c, "Failed to send mail", err)
		l.Sugar().Errorf("Initial email send failed", err)
		return
	}
	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, nil)
	if err != nil {
		RenderErrorTemplate(c, "Internal Server error occoured", err)
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
		RenderErrorTemplate(c, "Verify JWT token failed", err)
		return
	}
	tokenClaims, err := auth.ExtractClaims(token)
	if err != nil {
		RenderErrorTemplate(c, "Extract JWT Clain failed", err)
		return
	}
	emailID := tokenClaims.EmailID // this is the email id the user has signed up with
	var exp bool
	if tokenClaims.ExpiryDate.Before(time.Now()) {
		exp = true
	}
	emailRecords, err := gms.ListMainPage(ctx, emailID)
	if err != nil {
		RenderErrorTemplate(c, "An Internal server error occoured", err)
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
			EmailTz:       er.TimeZone,
		})
	}
	d.Timezone = constants.TimezonesSlice
	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "mainpage.html"))
	if err != nil {
		RenderErrorTemplate(c, "Parse mainpage template failed", err)
		l.Sugar().Errorf("parse template failed", err)
		return
	}

	// Execute the template and write the output to the response
	err = tmpl.Execute(c.Writer, d)
	if err != nil {
		RenderErrorTemplate(c, "Execute template failed", err)
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
		l.Sugar().Errorf("failed to parse form", err)
		RenderErrorTemplate(c, "Failed to parse form", err)
		return
	}

	emailID := c.PostForm("emailaddress")
	if emailID == "" {
		l.Sugar().Errorf("Email field cannot be empty", err)
		RenderErrorTemplate(c, "Email field cannot be empty", nil)
		return
	}
	timezone := c.PostForm("tz")
	if timezone == "" {
		l.Sugar().Errorf("timezone cannot be empty", err)
		RenderErrorTemplate(c, "Timezone cannot be empty", nil)
		return
	}

	authtoken := c.Query("tkn")
	token, err := auth.VerifyJWTToken(ctx, authtoken)
	if err != nil {
		RenderErrorTemplate(c, "Invalid authentication token", err)
		return
	}

	tokenClaims, err := auth.ExtractClaims(token)
	if err != nil {
		RenderErrorTemplate(c, "Failed to extract token claims", err)
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
		RenderErrorTemplate(c, "The entered input is not valid", err)
		return
	}

	mailTime, err := gms.ConvertMailTime(timezone)
	if err != nil {
		RenderErrorTemplate(c, "Convert Timezone format failed", err)
		return
	}

	id := uuid.New()
	err = gms.EmailRecord(ctx, &model.EmailRecord{
		ID:          id,
		EmailID:     emailID,
		OwnerMailID: ownerMailID,
		ExpiryDate:  time.Now().AddDate(0, 0, 7),
		TimeZone:    timezone,
		CreatedOn:   time.Now(),
		IsDeleted:   false,
	})
	if err != nil {
		RenderErrorTemplate(c, "Email record creation failed ", err)
		return
	}

	cache := dbpkg.GetCacheFromctx(ctx)
	cache.Set(mailTime.Format("15:04"), &dbpkg.CacheEntry{
		RecordID:      id,
		EmailID:       emailID,
		RandomNumbers: "",
		ExpiryDate:    time.Now().AddDate(0, 0, 7),
	})

	c.Redirect(302, fmt.Sprintf("/auth/gms?tkn=%s", authtoken))
}

func ToggleRecordActivityHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	recordID := c.Param("id")
	isActive := c.Param("isactive")
	req := DeactivateRequest{
		RecordID: recordID,
		IsActive: isActive,
	}

	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		l.Sugar().Errorf("the entered record id is not valid", err)
		RenderErrorTemplate(c, "the entered record id is not valid", err)
		return
	}
	authtoken := c.Query("tkn")
	_, err = auth.VerifyJWTToken(ctx, authtoken)
	if err != nil {
		RenderErrorTemplate(c, "Invalid authentication token", err)
		return
	}

	err = gms.ToggleActivityStatus(ctx, recordID, isActive)
	if err != nil {
		RenderErrorTemplate(c, "Error changing Activity Status", err)
		return
	}

	c.Redirect(302, fmt.Sprintf("/auth/gms?tkn=%s", authtoken))
}
