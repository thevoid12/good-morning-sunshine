// this package helps in sending email

package email

import (
	"context"
	"fmt"
	emailmodel "gms/pkg/email/model"
	logs "gms/pkg/logger"
	"net/smtp"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/spf13/viper"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "goodmrngsunshine@gmail.com"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	Recipient = ""

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	Subject = "This is a trail msg"

	// The HTML body for the email.
	HtmlBody = "<h1>hello world! this is a test email sent by void </h1>"

	//The email body for recipients with non-HTML email clients.
	TextBody = "this is a test email."

	// The character encoding for the email.
	CharSet = "UTF-8"
)

func SendEmailusingSES() {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("")},
	)

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return
	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)
}

func SendEmailUsingGmailSMTP(ctx context.Context, smtpStruct *emailmodel.SMTP) error {
	l := logs.GetLoggerctx(ctx)
	from := os.Getenv("FROM")
	user := os.Getenv("FROM")
	password := os.Getenv("GMS_GMAIL_PASS")

	to := []string{
		smtpStruct.ToAddress,
	}

	addr := viper.GetString("mail.gmailsmtp.address")
	host := viper.GetString("mail.gmailsmtp.host")
	contentType := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(
		"Subject: " + smtpStruct.Subject + "\r\n" +
			contentType + "\r\n" +
			smtpStruct.EmailBody + "\r\n")

	auth := smtp.PlainAuth("", user, password, host)

	err := smtp.SendMail(addr, auth, from, to, msg)
	if err != nil {
		l.Sugar().Error("send mail to "+smtpStruct.ToAddress+" failed", err)
		return err
	}

	l.Sugar().Info("Email Sent Successfully to " + smtpStruct.ToAddress)
	return nil
}
