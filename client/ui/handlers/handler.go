package handlers

import (
	"fmt"
	"gms/pkg/gms"
	logs "gms/pkg/logger"
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type CheckMail struct {
	Email string `json:"email" validate:"required,email`
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
	// email := c.Request.Form.Get("emailaddress")
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
	err = gms.EmailMainPage(ctx, emailID)
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

func MainPageHandler(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	authtoken := c.Query("tkn")
	fmt.Println(authtoken)
	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "mainpage.html"))
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
