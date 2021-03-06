package gautu

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.g7n3.com/hamster/gautu/model"

	"net/http"

	"fmt"

	"git.g7n3.com/hamster/gautu/common"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

const TOKEN_PREFIX = "TK_"

func FinishAuthorize(c *gin.Context, u interface{}) (code string, e error) {
	if c.Request.Form == nil {
		c.Request.ParseForm()
	}
	form := c.Request.Form.Encode()

	f, e := url.ParseQuery(form)
	if e != nil {
		return "", ERROR_MAP[E_INVALID_REQUEST]
	}

	if e := CheckClientValidator(f); e != E_INVALID_NONE {
		return "", ERROR_MAP[e]
	}

	if e := c.Request.ParseForm(); e != nil {
		return "", e
	}

	rt := f.Get("response_type")
	log.Println(rt)
	if rt == "code" {
		state := f.Get("state")
		cli := f.Get("client_id")
		uri := f.Get("redirect_uri")
		code = AuthorizeGenerateToken(cli, u.(model.User).ID.String())

		res := common.StitchAddress(
			[]string{"code", code},
			[]string{"state", state},
		)

		c.Redirect(http.StatusFound, strings.Join([]string{uri, res}, "?"))
	} else if rt == "token" {
		cli := f.Get("client_id")
		uri := f.Get("redirect_uri")
		scope := f.Get("scope")
		state := f.Get("state")
		code = AuthorizeGenerateToken(cli, u.(model.User).ID.String())
		acc, _ := AccessGenerateToken(cli, u.(model.User).ID.String(), time.Nanosecond.Nanoseconds(), true)
		res := common.StitchAddress(
			[]string{"access_token", acc},
			[]string{"token_type", "token"},
			[]string{"expires_in", ""},
			[]string{"scope", scope},
			[]string{"state", state},
		)
		c.Redirect(http.StatusFound, strings.Join([]string{uri, res}, "?"))
	}

	return code, nil
}

func CheckClientValidator(form url.Values) int {
	client := new(model.Client)

	id := form.Get("client_id")
	redirect_uri := form.Get("redirect_uri")
	log.Println("CheckClientValidator client_id: ", id)
	log.Println("CheckClientValidator redirect_uri: ", redirect_uri)
	if redirect_uri == "" {
		return E_UNAUTHORIZED_CLIENT
	}

	model.Gorm().First(client, "client_user = ?", id)
	if client.IsNull() {
		log.Println("authorize client is null")
		return E_UNAUTHORIZED_CLIENT
	}

	flag := client.CheckRedirectUri(redirect_uri)

	//	if client.RedirectUri == redirect_uri || client.RedirectUri == url.QueryEscape(redirect_uri) {
	if flag || client.RedirectUri == url.QueryEscape(redirect_uri) {
		log.Println("RedirectUri success")
		return E_INVALID_NONE
	}

	log.Println("first", client.RedirectUri)
	log.Println("second", redirect_uri)
	log.Println("no catched")
	return E_UNAUTHORIZED_CLIENT

}

func AuthorizeGenerateToken(cid, uid string) (code string) {
	buf := bytes.NewBufferString(cid)
	buf.WriteString(uid)

	token := uuid.NewV3(uuid.NewV1(), buf.String())
	code = base64.URLEncoding.EncodeToString(token.Bytes())
	code = TOKEN_PREFIX + strings.ToUpper(strings.TrimRight(code, "="))
	return
}

func AccessGenerateToken(cid, uid string, nano int64, genRefresh bool) (access, refresh string) {
	buf := bytes.NewBufferString(cid)
	buf.WriteString(uid)
	buf.WriteString(strconv.FormatInt(nano, 10))

	access = base64.URLEncoding.EncodeToString(uuid.NewV3(uuid.NewV4(), buf.String()).Bytes())
	access = TOKEN_PREFIX + strings.ToUpper(strings.TrimRight(access, "="))
	if genRefresh {
		refresh = base64.URLEncoding.EncodeToString(uuid.NewV5(uuid.NewV4(), buf.String()).Bytes())
		refresh = TOKEN_PREFIX + strings.ToUpper(strings.TrimRight(refresh, "="))
	}
	return
}

func ResponseError(c *gin.Context, code int) {
	if err := GetError(code); err != nil {
		c.Error(err)
	}

	return
}

//验证客户端信息
//失败跳转到主页
//验证错误跳转到主页
func authorizeGet(c *gin.Context) {
	session := AutoDefaultSession(c)
	j := session.Get("LoggedUserInfo")
	user := new(model.User)
	err := common.JsonDecode(j, user)
	if j == "" && err != nil {
		log.Println("no info")
		redirectToSign(c)
		return
	}

	client := getClient(c)

	if !client.IsNull() && client.Type == 0 {
		log.Println("client authorize")
		code, e := FinishAuthorize(c, *user)
		if e == nil {
			GetRedis().Do("SET", code, j)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error": e.Error(),
			})
			return
		}
	}

	if common.JsonDecode(j, user) == nil {
		c.HTML(http.StatusOK, "authorize.tmpl", gin.H{"user": *user})
		return
	}
	log.Println("default to sign")
	redirectToSign(c)
	return
}

//验证通过回跳
func authorizePost(c *gin.Context) {
	session := AutoDefaultSession(c)

	j := session.Get("LoggedUserInfo")

	user := model.User{}
	if j == "" || common.JsonDecode(j, &user) != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "system error",
		})
		return
	}

	code, e := FinishAuthorize(c, user)
	if e == nil {
		GetRedis().Do("SET", code, j)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error": e.Error(),
		})
	}

	return
}

func getClient(c *gin.Context) *model.Client {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}
	form := c.Request.Form
	fmt.Println(form)
	cid := form.Get("client_id")
	client := new(model.Client)
	model.Gorm().First(client, "client_user = ?", cid)

	return client

}
