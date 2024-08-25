// package gms has all functionality related to good morning sunshine
package gms

import (
	"fmt"

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
