package gautu

import "github.com/gin-gonic/gin"

func home(c *gin.Context) {
	c.HTML(200, "home.tmpl", gin.H{})
}
