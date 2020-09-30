package main

import (
	"fmt"
	"gin_demo/dao"
	"gin_demo/function"
	"gin_demo/models"
	"github.com/gin-gonic/gin"
	newuuid "github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"net/http"
	"time"
)

var err error

func main() {
	// connect to the DB
	dao.DB, err = gorm.Open("mysql", dao.DbURL(dao.BuildDBConfig()))
	if err != nil {
		fmt.Println("Status:", err)
	}
	defer dao.DB.Close()

	// auto migrate the user structure into the DB table
	dao.DB.AutoMigrate(&models.User{})

	// create a default router
	r := gin.Default()

	r.GET("/hello", func (c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	v1Group := r.Group("v1")
	{
		// add user
		v1Group.POST("/user", func (c *gin.Context) {
			// get the user info from the request
			var user models.User
			c.BindJSON(&user)
			// auto set the id, create time and update time
			user.ID = newuuid.New().String()
			user.AccountCreated = time.Now()
			user.AccountUpdated = time.Now()

			// check the complexity of the password
			// 2 means the password must have at least 2 different char
			code := function.CheckPassword(user.Password, 2)
			switch code {
			case -1: c.JSON(http.StatusOK, gin.H{"error": "the password is too short, please use at least 8 char"})
					return
			case 0: c.JSON(http.StatusOK, gin.H{"error": "the password is too week, please use letters, digits and special char"})
					return
			}

			// if passed the password test, then we can use this password
			var pass = function.HashAndSalt(function.GetPwd(user.Password))
			user.Password = pass

			// check the email is valid or not
			if !function.CheckEmail(user.EmailAddress) {
				c.JSON(http.StatusOK, gin.H{"error": "the email address is not valid"})
				return
			}

			// check for the email address existing, if exist return 400
			var newUser models.User
			if err = dao.DB.Where("email_address=?", user.EmailAddress).First(&newUser).Error; err==nil {
				c.JSON(400, gin.H{"error": "the email address is already exist"})
				return
			}

			// send into the DB, and then response
			if err := dao.DB.Create(&user).Error;err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"msg": "success",
					"data": user,
				})
			}
		})

		// view user
		/*v1Group.GET("/user", func (c *gin.Context) {
			var userList []models.User
			if err = db.Find(&userList).Error;err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"msg": "success",
					"data": userList,
				})
			}
		})*/

		// delete user
		v1Group.DELETE("/user/:email_address", func (c *gin.Context) {
			email, valid := c.Params.Get("email_address")
			if !valid {
				c.JSON(http.StatusOK, gin.H{"error": "email is not exist"})
				return
			}
			if err = dao.DB.Where("email_address=?", email).Delete(models.User{}).Error; err!=nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{email: "Deleted"})
			}
		})
	}

	// Group using gin.BasicAuth() middleware
	// gin.Accounts is a shortcut for map[string]string
	authorized := r.Group("/v1", function.BasicAuth())

	// basic authorized to get a user info
	authorized.GET("/user/self", func(c *gin.Context) {
		//email := c.Params.ByName("email_address")
		email := function.FetchUsername
		var user models.User
		err := dao.DB.Where("email_address=?", email).First(&user).Error
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"id": user.ID,
				"create time": user.AccountCreated,
				"update time": user.AccountUpdated,
				"email": user.EmailAddress,
				"first name": user.FirstName,
				"last name": user.LastName,
			})
		}
	})

	// update user
	authorized.PUT("/user/self", func (c *gin.Context) {
		/*email, valid := c.Params.Get("email_address")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "email address is not exist"})
			return
		}*/
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}	// get the user info based on email
		id := user.ID
		accountCreate := user.AccountCreated
		password := user.Password
		c.BindJSON(&user)
		user.AccountUpdated = time.Now()

		if !function.ComparePasswords(password, function.GetPwd(function.FetchPassword)) {
			code := function.CheckPassword(user.Password, 2)
			switch code {
			case -1: c.JSON(http.StatusOK, gin.H{"error": "the password is too short, please use at least 8 char"})
				return
			case 0: c.JSON(http.StatusOK, gin.H{"error": "the password is too week, please use letters, digits and special char"})
				return
			}

			var pass = function.HashAndSalt(function.GetPwd(user.Password))
			user.Password = pass
		}

		/*
		user.ID = id
		user.AccountCreated = accountCreate*/

		// if user wants to change the email
		if email != *user.EmailAddress || id != user.ID || accountCreate != user.AccountCreated{
			c.JSON(400, gin.H{"error": "The user cannot change the email address, id or create time"})
			return
		}

		if err = dao.DB.Save(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"id": user.ID,
				"create time": user.AccountCreated,
				"update time": user.AccountUpdated,
				"email": user.EmailAddress,
				"first name": user.FirstName,
				"last name": user.LastName,
			})
		}
	})

	r.Run(":9090")
}





