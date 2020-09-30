package function

import (
	"encoding/base64"
	"gin_demo/dao"
	"gin_demo/models"
	"github.com/gin-gonic/gin"
	"strings"
)

var FetchUsername string
var FetchPassword string

func BasicAuth() gin.HandlerFunc {

	return func(c *gin.Context) {
		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			RespondWithError(401, "Unauthorized", c)
			return
		}
		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !AuthenticateUser(pair[0], pair[1]) {
			RespondWithError(401, "Unauthorized", c)
			return
		}

		c.Next()
	}
}

func AuthenticateUser(username, password string) bool {
	var user models.User
	// fetch user from database. Here db.Client() is connection to your database. You will need to import your db package above.
	// This is just for example purpose
	/*err := Config.DB.Where(models.User{EmailAddress: username, Password: password}).FirstOrCreate(&user)
	if err.Error != nil {
		return false
	}
	return true*/
	FetchUsername = username
	FetchPassword = password
	err := dao.DB.Where("email_address=?", username).First(&user).Error
	if err != nil  {
		return false
	}
	valid := ComparePasswords(user.Password, GetPwd(password))
	if valid {
		return true
	} else {
		return false
	}

}

func RespondWithError(code int, message string, c *gin.Context) {
	resp := map[string]string{"error": message}

	c.JSON(code, resp)
	c.Abort()
}
