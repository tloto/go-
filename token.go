package gautu

import (
	//	"encoding/json"
	"fmt"
	"net/http"

	"time"

	"log"

	"git.g7n3.com/hamster/gautu/common"
	"git.g7n3.com/hamster/gautu/model"
	"github.com/gin-gonic/gin"
)

func tokenPost(c *gin.Context) {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}

	if e := c.Err(); e != nil {
		log.Println("err1")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
		})
		return
	}
	if e := AccessCheck(c); e != nil {

		log.Println("err2")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
		})
		return
	}

	form := c.Request.Form
	grant_type := form.Get("grant_type")

	if grant_type == "authorization_code" {
		Token(c)
	} else if grant_type == "refresh_token" {
		RefreshToken(c)
	} else {
		log.Println("err3")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ERROR_MAP[E_INVALID_GRANT],
		})
		return
	}

}

//获取Token
func Token(c *gin.Context) {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}

	form := c.Request.Form
	code := form.Get("code")
	redirect_uri := form.Get("redirect_uri")

	if code == "" || redirect_uri == "" {
		log.Println("err3")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ERROR_MAP[E_INVALID_REQUEST],
		})
		return
	}

	j, e := GetRedis().Do("GET", code)
	if e != nil {
		log.Println("err4")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
		})
		return
	}

	user := model.User{}

	common.JsonDecode(string(j.([]byte)), &user)
	if user.IsNull() == true {
		log.Println("err5")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ERROR_MAP[E_INVALID_REQUEST],
		})
		return
	}
	client := new(model.Client)
	log.Println(c.Request.Form.Get("client_id"))
	//gs.GetDB().Where("client_id = ?", c.Request.Form.Get("client_id")).First(&cli)
	model.Gorm().First(client, "client_user = ?", c.Request.Form.Get("client_id"))

	if client.IsNull() {
		log.Println("client is null")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ERROR_MAP[E_INVALID_REQUEST],
		})
		return
	}

	author := new(model.Authorize)
	oid := model.GenerateOpenID(client.ID.String(), user.ID.String())

	author.SubID = oid
	author.ClientID = client.ID
	author.UserID = user.ID

	acc, ref := AccessGenerateToken(client.ClientUser, user.ID.String(), time.Nanosecond.Nanoseconds(), true)

	if exts, author2 := CreateIfNotExist(author); exts {
		author.RefreshToken = ref
		author.AccessToken = acc
		_ = author2
	}

	GetRedis().Do("SET", acc, author.SubID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  acc,
		"token_type":    "bearer",
		"refresh_token": ref,
		"expires_in":    "0",
		//"expires":       time.Nanosecond.Nanoseconds(),
	})
}

//token过期，使用refresh_token重新获取
func RefreshToken(c *gin.Context) {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}

	form := c.Request.Form
	client_id := form.Get("client_id")
	refresh_token := form.Get("refresh_token")

	if refresh_token == "" {
		log.Println("err3")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ERROR_MAP[E_INVALID_REQUEST],
		})
		return
	}

	authorize := new(model.Authorize)

	model.Gorm().First(authorize, "refresh_token = ?", refresh_token)

	acc, ref := AccessGenerateToken(client_id, authorize.UserID.String(), time.Nanosecond.Nanoseconds(), true)

	authorize.AccessToken = acc
	authorize.RefreshToken = ref

	GetRedis().Do("GEt", acc, authorize.SubID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  acc,
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": ref,
		//		"example_parameter": "example_value",
	})
}

/*
validate:
client_id	Required. The client application's id.
client_secret	Required. The client application's client secret .
grant_type	Required. Must be set to authorization_code .
code	Required. The authorization code received by the authorization server.
redirect_uri	Required, if the request URI was included in the authorization request. Must be identical then.
*/

func AccessCheck(c *gin.Context) error {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}

	client_id := c.Request.Form.Get("client_id")
	client_secret := c.Request.Form.Get("client_secret")
	grant_type := c.Request.Form.Get("grant_type")
	//	code := c.Request.Form.Get("code")
	//	redirect_uri := c.Request.Form.Get("redirect_uri")

	if client_id == "" || client_secret == "" || grant_type == "" {
		fmt.Println("missed", c.Request.Form)
		return ERROR_MAP[E_INVALID_REQUEST]
	}

	if c.Request.Method != "POST" {
		fmt.Println("not post")
		return ERROR_MAP[E_INVALID_REQUEST]
	}

	client := new(model.Client)
	model.Gorm().First(client, "client_user = ?", client_id)

	if client.IsNull() {
		fmt.Println("client is null")
		return ERROR_MAP[E_INVALID_CLIENT]
	}

	if client.GetSecret() != client_secret {
		fmt.Println("client secret is wrong!")
		return ERROR_MAP[E_INVALID_CLIENT]
	}

	if grant_type == "authorization_code" {
		flag := client.CheckRedirectUri(c.Request.Form.Get("redirect_uri"))
		if !flag {
			fmt.Println("RedirectUri is wrong！")
			return ERROR_MAP[E_INVALID_GRANT]
		}
	}

	return ERROR_MAP[E_INVALID_NONE]

}

func CreateIfNotExist(author *model.Authorize) (bool, *model.Authorize) {
	existAuthor := new(model.Authorize)
	model.Gorm().First(existAuthor, "sub_id = ?", author.SubID)
	log.Println("CreateIfNotExist ID: " + existAuthor.ID.String())
	log.Println("SubID1: " + author.SubID)
	log.Println("SubID2: " + existAuthor.SubID)
	if existAuthor.SubID == author.SubID {
		return true, existAuthor
	}
	model.Gorm().Create(author)
	return false, nil
}

func RefreshCheck() {

}
