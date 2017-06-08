package gautu

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"git.g7n3.com/hamster/gautu/common"
	"git.g7n3.com/hamster/gautu/model"
	"github.com/gin-gonic/gin"
)

//用户登录界面
func loginGet(c *gin.Context) {
	session := AutoDefaultSession(c)
	user := model.User{}
	j := session.Get("LoggedUserInfo")

	if j != "" {
		session.Delete("LoggedUserInfo")
		e := common.JsonDecode(j, &user)
		if e == nil && user.ID.String() != "" {

		}
	}

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}
	session.Set("Form", c.Request.Form.Encode())

	c.HTML(http.StatusOK, "login.tmpl", gin.H{})
}

//登录提交页面
func loginPost(c *gin.Context) {
	session := AutoDefaultSession(c)
	funcWay := make(map[string]func(c *gin.Context, s *Session) bool)
	funcWay["1"] = ValidatePhone
	funcWay["0"] = ValidateUser

	way := c.DefaultPostForm("way", "0")

	if f, b := funcWay[way]; b == true {
		if f(c, session) != true {
			return
		}
	}

	rlt := A_LOGIN_SUCCESS
	url := ParseURL(c, "/home")
	if f := PullForm(session); f != "" {
		url.Path = "authorize"
		url.RawQuery = f
	}
	rlt.Data = map[string]string{
		"URL": url.String(),
	}
	c.JSON(200, rlt)

	return
}

func PullForm(s *Session) string {
	form := s.Get("Form")
	defer s.Delete("Form")
	return form

}

func ParseURL(c *gin.Context, path string) *url.URL {
	u := new(url.URL)
	u.Path = path
	return u
}

func ValidatePhone(c *gin.Context, s *Session) bool {
	mobile := c.DefaultPostForm("mobile", "")
	code := c.DefaultPostForm("code", "")
	if mobile == "" || code == "" {
		//c.Error(errors.New("username or password is wrong"))
		c.JSON(200, A_MOBILE_OR_CODE_CANNOT_NULL)
		return false
	}

	log.Println("ValidatePhone code:" + code)
	log.Println("ValidatePhone mobile:" + mobile)
	user := new(model.User)

	t := model.VerifyAccountType(mobile)
	if t == model.ACCOUNT_TYPE_NONE {
		c.JSON(200, A_MOBILE_LENGTH_WRONG)
		return false
	}

	rmap := common.MessageCheck(mobile, code)

	if rmap == nil {
		c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
		return false
	}

	if v, b := (*rmap)["code"]; b == false || v != "0" {
		c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
		return false
	}

	user.GetUser(mobile, t)
	if user.IsNull() == true {
		user = model.NewUser()
		user.Mobile = mobile
		model.Save(user)
		user.GetUser(mobile, t)
	}
	log.Println("ValidatePhone uid: " + user.ID.String())

	user.UpdateSignInfo(common.ObtainClientIP(c.Request), time.Now().String(), "")
	j, e := common.JsonEncode(user)
	if e == nil {
		s.Set("LoggedUserInfo", j)
	}
	model.Save(user)
	return true
}

// return: continue flag
func ValidateUser(c *gin.Context, s *Session) bool {
	uname := c.DefaultPostForm("username", "")
	password := c.DefaultPostForm("password", "")

	if uname == "" || password == "" {
		//c.Error(errors.New("username or password is wrong"))
		c.JSON(200, A_NAME_OR_PASSWORD_CANNOT_NULL)
		return false
	}
	user := model.User{}

	t := model.VerifyAccountType(uname)
	if t == model.ACCOUNT_TYPE_NONE {
		c.JSON(200, A_NAME_OR_PASSWORD_WRONG)
		return false
	}

	user.GetUser(uname, t)

	if !user.VerifyPassword(password) {
		c.JSON(200, A_NAME_OR_PASSWORD_WRONG)
		return false
	}

	user.UpdateSignInfo(common.ObtainClientIP(c.Request), time.Now().String(), "")
	j, e := common.JsonEncode(user)
	if e == nil {
		s.Set("LoggedUserInfo", j)
	}
	model.Save(user)
	return true
}

func registerGet(c *gin.Context) {
	c.HTML(http.StatusOK, "register.tmpl", gin.H{})
}

func registerPost(c *gin.Context) {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}
	fmt.Println(c.Request.Form)
	uname := c.DefaultPostForm("username", "")
	mobile := c.DefaultPostForm("mobile", "")
	password := c.DefaultPostForm("password", "")
	vpassword := c.DefaultPostForm("vpassword", "")
	code := c.DefaultPostForm("code", "")

	cid := ParseClientUser(c)
	log.Println("register cid: " + cid)

	user := model.NewUser()
	if uname == "" || mobile == "" || password == "" || vpassword == "" {
		c.JSON(http.StatusOK, A_MUST_NOT_BE_NULL)
		return
	}

	if len(uname) < 6 {
		c.JSON(http.StatusOK, A_NAME_LENGTH_WRONG)
		return
	}

	if b, e := model.VerifyUsername(uname); b == false || e != nil {
		c.JSON(http.StatusOK, A_NAME_FORMAT_WRONG)
		return
	}

	user.GetUserByUsername(uname)
	if !user.IsNull() {
		c.JSON(http.StatusOK, A_NAME_IS_EXISTS)
		return
	}

	if b, e := model.VerifyMobile(mobile); b == false || e != nil {
		c.JSON(http.StatusOK, A_MOBILE_LENGTH_WRONG)
		return
	}

	user.GetUserByMobile(mobile)
	if !user.IsNull() {
		c.JSON(http.StatusOK, A_MOBILE_IS_EXISTS)
		return
	}

	rmap := common.MessageCheck(mobile, code)

	if rmap == nil {
		c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
		return
	}

	if v, b := (*rmap)["code"]; b == false || v != "0" {
		c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
		return
	}
	log.Println("register uid: " + user.ID.String())
	if user.IsNull() {
		user.Mobile = mobile
		user.Username = uname
		user.GenerateBCryptPassword(password)
		user.RegisterType = cid
		log.Println("regiseter save: " + user.Username + "|" + user.Mobile)
		model.Gorm().Create(user)
	}
	c.JSON(http.StatusOK, A_REGISTER_SUCCESS)
	return

}

func ParseClientUser(c *gin.Context) string {
	session := AutoDefaultSession(c)
	form := session.Get("Form")

	log.Println("ParseClientUser form: " + form)
	if form == "" {
		return ""
	}

	client := model.NewClient()
	if fmap, e := url.ParseQuery(form); e == nil {

		if cid, b := fmap["client_id"]; b == true {
			model.Gorm().First(client, "client_user = ?", cid)
			if !client.IsNull() {
				return client.ClientUser
			}
		}
	}
	return ""
}

func registerPhoneSend(c *gin.Context) {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}
	fmt.Println(c.Request.Form)
	mobile := c.DefaultPostForm("mobile", "")
	fmt.Println("mobile", mobile, reflect.TypeOf(mobile))
	if b, e := model.VerifyMobile(mobile); b == false || e != nil {
		fmt.Println(b, e, mobile)
		c.JSON(http.StatusOK, A_MOBILE_LENGTH_WRONG)
		return
	}

	common.RegisterSend(mobile)

	c.JSON(http.StatusOK, A_MESSAGE_SEND_SUCCESS)
	return
}

func registerPhoneCheck(c *gin.Context) {

	if c.Request.Form == nil {
		c.Request.ParseForm()
	}
	mobile := c.DefaultPostForm("mobile", "")
	code := c.DefaultPostForm("code", "")

	if b, e := model.VerifyMobile(mobile); b == false || e != nil {
		c.JSON(http.StatusOK, A_MOBILE_LENGTH_WRONG)
	}

	if code == "" {
		c.JSON(http.StatusOK, A_MESSAGE_WRONG)
		return
	}

	rmap := common.MessageCheck(mobile, code)

	if rmap == nil {
		c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
		return
	}
	c.JSON(http.StatusOK, A_MESSAGE_CHECK_SUCCESS)
	return
}

func forgetGet(c *gin.Context) {
	c.HTML(http.StatusOK, "forget.tmpl", gin.H{})
}

func resetGet(c *gin.Context) {
	c.HTML(http.StatusOK, "reset.tmpl", gin.H{})
}

func forgetPost(c *gin.Context) {

	session := AutoDefaultSession(c)
	if c.Request.Form == nil {
		c.Request.ParseForm()
	}

	if uid := session.Get("ForgetUser"); uid == "" {
		fmt.Println(uid)
		mobile := c.DefaultPostForm("mobile", "")
		uname := c.DefaultPostForm("username", "")
		mail := c.DefaultPostForm("mail", "")
		code := c.DefaultPostForm("code", "")
		if mobile == "" && uname == "" && mail == "" && code == "" {
			c.JSON(http.StatusOK, A_FORGET_NEED_IS_NULL)
			return
		}

		user := model.User{}

		if mobile != "" {
			if b, e := model.VerifyMobile(mobile); b == true && e == nil {
				user.GetUser(mobile, model.ACCOUNT_TYPE_MOBILE)
			}
		} else if uname != "" {
			if b, e := model.VerifyUsername(mobile); b == true && e == nil {
				user.GetUser(mobile, model.ACCOUNT_TYPE_UNAME)
			}
		} else {
		}

		if user.Mobile == "" {
			c.JSON(http.StatusOK, A_FORGET_ACCOUNT_WRONG)
			return

		}

		rmap := common.MessageCheck(mobile, code)

		if rmap == nil {
			c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
			return
		}

		if v, b := (*rmap)["code"]; b == false || v != "0" {
			c.JSON(http.StatusOK, A_MESSAGE_CHECK_FAILED)
			return
		}

		session.Set("ForgetUser", user.ID.String())
		c.JSON(http.StatusOK, A_FORGET_CHECK_SUCCESS)
	} else {
		fmt.Println(uid)

		rf := c.DefaultPostForm("type", "")
		pass := c.DefaultPostForm("password", "")
		vpass := c.DefaultPostForm("vpassword", "")
		if rf != "reset" {
			session.Delete("ForgetUser")
			c.JSON(http.StatusOK, A_FORGET_SYSTEM_ERROR)
			return
		}

		if pass == "" || vpass == "" {
			c.JSON(http.StatusOK, A_FORGET_PASSWORD_CANNOT_NULL)
			return
		}

		if pass != vpass {
			c.JSON(http.StatusOK, A_FORGET_PASSWORD_FILL_OUT)
			return

		}

		user := model.User{}

		model.FirstById(&user, uid)

		if user.ID.String() == "" {
			c.JSON(http.StatusOK, A_FORGET_ACCOUNT_WRONG)
			return
		}
		user.UpdateSignInfo(common.ObtainClientIP(c.Request), time.Now().String(), "")
		user.GenerateBCryptPassword(pass)
		model.Save(user)
		session.Delete("ForgetUser")
		c.JSON(http.StatusOK, A_FORGET_CHANGE_SUCCESS)
		return
	}

	return

}

func changeGet(c *gin.Context) {
	c.HTML(http.StatusOK, "change.tmpl", gin.H{})
}

func changePost(c *gin.Context) {

	//session := AutoDefaultSession(c)
	if c.Request.Form == nil {
		c.Request.ParseForm()
	}

	//opass := c.DefaultPostForm("opassword", "")
	//pass := c.DefaultPostForm("password", "")
	//vpass := c.DefaultPostForm("vpassword", "")

	return
}

func logout(c *gin.Context) {
	session := AutoDefaultSession(c)
	session.Delete("LoggedUserInfo")
	redirectToSign(c)
}

//跳转到登陆页面，冰保留uri参数
func redirectToSign(c *gin.Context) {
	u := new(url.URL)
	u.Path = "/login"
	u.RawQuery = c.Request.Form.Encode()
	log.Println("sign: " + u.String())

	c.Writer.Header().Set("Location", u.String())
	c.Writer.WriteHeader(http.StatusFound)
	return
}
