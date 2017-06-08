package gautu

import (
	"fmt"
	"log"
	"net/url"
	"testing"

	"os"

	"git.g7n3.com/hamster/gautu/model"
)

func TestParseClientUser(t *testing.T) {
	form := "client_id=cid0001&redirect_uri=https%3A%2F%2Ftest3.mana.cn%2Fauth%2Fmana_callback&response_type=code"
	v, _ := url.ParseQuery(form)
	fmt.Println(v)
}

func TestCreateUser(t *testing.T) {
	user := model.NewUser()
	user.Mobile = "123"
	user.Username = "fdsafdsa"
	user.GenerateBCryptPassword("12345")
	user.RegisterType = "cidejkfdsa"
	log.Println("regiseter save: " + user.Username + "|" + user.Mobile)
	//	model.Gorm().Create(user)
	log.Println(model.Save(user))
}

func TestParseURL(t *testing.T) {
	log.Println(os.ModeAppend)
	log.Println(os.ModePerm)

}
