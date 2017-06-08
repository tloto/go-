package gautu

import (
	"log"

	"github.com/gin-gonic/gin"
)

func SignCheck(c *gin.Context) {
	log.Println("SignCheck")
	session := AutoDefaultSession(c)
	user := session.Get("LoggedUserInfo")
	if user == "" {
		if c.Request.Form == nil {
			c.Request.ParseForm()
		}
		session.Set("Form", c.Request.Form.Encode())
		redirectToSign(c)
		c.Abort()
		return
	}
	c.Next()
}

func VisitLog(c *gin.Context) {
	log.Println("VisitLog URL: ", c.Request.URL.RawQuery)
	c.Next()
}
