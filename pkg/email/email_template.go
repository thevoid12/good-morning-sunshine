package email

func GetEmailTemplate(template TemplateType) string {
	switch template {
	case GmSunshine:
		body := Template1()
		return body
	case Thanks:
		body := Template2()
		return body
	}
	return ""
}

func Template1() string {
	body := `<h1>Good Morning Sunshine!!!</h1>`
	return body
}

func Template2() string {
	body := `<h1>Thanks for subscribing for gms.Have A great day</h1>`
	return body
}
