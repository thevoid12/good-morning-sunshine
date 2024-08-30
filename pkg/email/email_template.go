package email

func GetEmailTemplate(randomNumber int) string {
	switch {
	case randomNumber == 0:
		body := Template0()
		return body
	case randomNumber == 1:
		body := Template1()
		return body
	}
	//add how many templates we want
	//TODO: later generate the template through LLMs
	return ""
}

func Template0() string {
	body := `<h1>Good Morning Sunshine!!!</h1>`
	return body
}

func Template1() string {
	body := `<h1>Thanks for subscribing for gms.Have A great day</h1>`
	return body
}
