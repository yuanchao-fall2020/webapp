package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gin_demo/dao"
	"gin_demo/function"
	"gin_demo/logger"
	"gin_demo/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/go-sql-driver/mysql"
	"gopkg.in/alexcesaro/statsd.v2"
	//"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	newuuid "github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	//"logger"
)

var err error
var AccessKeyID string
var SecretAccessKey string
var MyRegion string
var MyBucket string
var filepath string

func main() {

	logger.Log.Printf("web app is starting...")

	// connect to the DB
	rootCertPool := x509.NewCertPool()
	pem := `-----BEGIN CERTIFICATE-----
MIIEBjCCAu6gAwIBAgIJAMc0ZzaSUK51MA0GCSqGSIb3DQEBCwUAMIGPMQswCQYD
VQQGEwJVUzEQMA4GA1UEBwwHU2VhdHRsZTETMBEGA1UECAwKV2FzaGluZ3RvbjEi
MCAGA1UECgwZQW1hem9uIFdlYiBTZXJ2aWNlcywgSW5jLjETMBEGA1UECwwKQW1h
em9uIFJEUzEgMB4GA1UEAwwXQW1hem9uIFJEUyBSb290IDIwMTkgQ0EwHhcNMTkw
ODIyMTcwODUwWhcNMjQwODIyMTcwODUwWjCBjzELMAkGA1UEBhMCVVMxEDAOBgNV
BAcMB1NlYXR0bGUxEzARBgNVBAgMCldhc2hpbmd0b24xIjAgBgNVBAoMGUFtYXpv
biBXZWIgU2VydmljZXMsIEluYy4xEzARBgNVBAsMCkFtYXpvbiBSRFMxIDAeBgNV
BAMMF0FtYXpvbiBSRFMgUm9vdCAyMDE5IENBMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEArXnF/E6/Qh+ku3hQTSKPMhQQlCpoWvnIthzX6MK3p5a0eXKZ
oWIjYcNNG6UwJjp4fUXl6glp53Jobn+tWNX88dNH2n8DVbppSwScVE2LpuL+94vY
0EYE/XxN7svKea8YvlrqkUBKyxLxTjh+U/KrGOaHxz9v0l6ZNlDbuaZw3qIWdD/I
6aNbGeRUVtpM6P+bWIoxVl/caQylQS6CEYUk+CpVyJSkopwJlzXT07tMoDL5WgX9
O08KVgDNz9qP/IGtAcRduRcNioH3E9v981QO1zt/Gpb2f8NqAjUUCUZzOnij6mx9
McZ+9cWX88CRzR0vQODWuZscgI08NvM69Fn2SQIDAQABo2MwYTAOBgNVHQ8BAf8E
BAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUc19g2LzLA5j0Kxc0LjZa
pmD/vB8wHwYDVR0jBBgwFoAUc19g2LzLA5j0Kxc0LjZapmD/vB8wDQYJKoZIhvcN
AQELBQADggEBAHAG7WTmyjzPRIM85rVj+fWHsLIvqpw6DObIjMWokpliCeMINZFV
ynfgBKsf1ExwbvJNzYFXW6dihnguDG9VMPpi2up/ctQTN8tm9nDKOy08uNZoofMc
NUZxKCEkVKZv+IL4oHoeayt8egtv3ujJM6V14AstMQ6SwvwvA93EP/Ug2e4WAXHu
cbI1NAbUgVDqp+DRdfvZkgYKryjTWd/0+1fS8X1bBZVWzl7eirNVnHbSH2ZDpNuY
0SBd8dj5F6ld3t58ydZbrTHze7JJOd8ijySAp4/kiu9UfZWuTPABzDa/DSdz9Dk/
zPW4CXXvhLmE02TA9/HeCw3KEHIwicNuEfw=
-----END CERTIFICATE-----`

	if ok := rootCertPool.AppendCertsFromPEM([]byte(pem)); !ok {
		log.Fatal("Failed to append PEM.")
	}

	mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:      rootCertPool,
	})


	// try to connect to mysql database.
	cfg := mysql.Config{
		User:   "csye6225fall2020",
		Passwd: "Y940519a",
		Addr:   fmt.Sprintf("%s:%d", dao.GetHostname(), 3306), //IP:PORT
		Net:    "tcp",
		DBName: "csye6225",
		Loc: time.Local,
		AllowNativePasswords: true,
		TLSConfig: "custom",
	}




	str := cfg.FormatDSN()

	// db, err := gorm.Open("mysql", str)

	//dao.DB, err = gorm.Open("mysql", dao.DbURL(dao.BuildDBConfig()))
	dao.DB, err = gorm.Open("mysql", str)
	if err != nil {
		fmt.Println("Status:", err)
	}
	defer dao.DB.Close()

	// auto migrate the user structure into the DB table
	dao.DB.AutoMigrate(&models.User{})
	dao.DB.AutoMigrate(&models.Question{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	dao.DB.AutoMigrate(&models.Answer{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT").AddForeignKey("question_id", "questions(id)", "RESTRICT", "RESTRICT")
	dao.DB.AutoMigrate(&models.Category{})
	dao.DB.AutoMigrate(&models.QuestionCategory{}).AddForeignKey("question_id", "questions(id)", "RESTRICT", "RESTRICT").AddForeignKey("category_id", "categories(id)", "RESTRICT", "RESTRICT")

	dao.DB.AutoMigrate(&models.FileQuestion{}).AddForeignKey("question_id", "questions(id)", "RESTRICT", "RESTRICT")
	dao.DB.AutoMigrate(&models.FileAnswer{}).AddForeignKey("answer_id", "answers(id)", "RESTRICT", "RESTRICT")

	// create a default router
	r := gin.Default()

	r.GET("/hello", func (c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	d, err := statsd.New() // Connect to the UDP port 8125 by default.
	if err != nil {
		// If nothing is listening on the target port, an error is returned and
		// the returned client does nothing but is still usable. So we can
		// just log the error and go on.
		logger.Log.Printf(err.Error())
	}
	defer d.Close()

	num1 := 0
	num2 := 0
	num3 := 0
	num4 := 0
	num5 := 0
	num6 := 0
	num7 := 0
	num8 := 0
	num9 := 0
	num10 := 0
	num11 := 0
	num12 := 0
	num13 := 0
	num14 := 0
	num15 := 0
	num16 := 0
	num17 := 0
	num18 := 0
	num19 := 0

	v1Group := r.Group("/v1")
	{
		// add user
		v1Group.POST("/user", func (c *gin.Context) {

			logger.Log.Printf("POST a user is starting...")

			num1++
			// Time something.
			t := d.NewTiming()

			d.Count("foo.counter", num1)

			/*// It can also be used as a one-liner to easily time a function.
			pingHomepage := func() {
				defer d.NewTiming().Send("homepage.response_time")

				//print("http://example.com/")
			}
			pingHomepage()

			// Cloning a Client allows using different parameters while still using the
			// same connection.
			// This is way cheaper and more efficient than using New().
			stat := d.Clone(statsd.Prefix("http"), statsd.SampleRate(0.2))
			stat.Increment("view") // Increments http.view*/

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
				logger.Log.Printf("error: the password is too short")
				return
			case 0: c.JSON(http.StatusOK, gin.H{"error": "the password is too week, please use letters, digits and special char"})
				logger.Log.Printf("error: the password is too week")
				return
			}

			// if passed the password test, then we can use this password
			var pass = function.HashAndSalt(function.GetPwd(user.Password))
			user.Password = pass

			// check the email is valid or not
			if !function.CheckEmail(user.EmailAddress) {
				c.JSON(http.StatusOK, gin.H{"error": "the email address is not valid"})
				logger.Log.Printf("error: the email address is not valid")
				return
			}

			t2 := d.NewTiming()

			// check for the email address existing, if exist return 400
			// ???????
			var newUser models.User
			if err = dao.DB.Where("email_address=?", user.EmailAddress).First(&newUser).Error; err==nil {
				c.JSON(400, gin.H{"error": "the email address is already exist"})
				logger.Log.Printf("error: the email address is exit")
				return
			}

			// send into the DB, and then response
			if err := dao.DB.Create(&user).Error;err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				logger.Log.Printf(err.Error())
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

			t2.Send("db_response_time")
			t.Send("api_response_time")
			logger.Log.Printf("POST a user is done...")
		})

		// view user
		v1Group.GET("/user/", func (c *gin.Context) {

			logger.Log.Printf("View a user is starting...")

			num2++
			// Time something.
			t := d.NewTiming()

			// Increment a counter.
			//d.Increment("foo.counter")
			d.Count("foo.counter", num2)

			var userList []models.User
			if err = dao.DB.Find(&userList).Error;err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				logger.Log.Printf(err.Error())
			} else {
				c.JSON(http.StatusOK, gin.H{
					"msg": "success",
					"data": userList,
				})
			}

			t.Send("db_response_time")
			t.Send("api_response_time")
			logger.Log.Printf("View a user is done...")
		})

		// get a user info by id
		v1Group.GET("/user/:id", func(c *gin.Context) {
			logger.Log.Printf("Get a user is starting...")
			num3++
			// Time something.
			t := d.NewTiming()

			// Increment a counter.
			//d.Increment("foo.counter")
			d.Count("foo.counter", num3)

			id := c.Params.ByName("id")
			var user models.User
			err := dao.DB.Where("id=?", id).First(&user).Error

			t.Send("db_response_time")

			if err != nil {
				c.JSON(404, gin.H{"error": err.Error()})
				logger.Log.Printf(err.Error())
			} else {
				c.JSON(http.StatusOK, gin.H{
					"data": user,
				})
			}

			t.Send("api_response_time")
			logger.Log.Printf("Get a user is done...")
		})

		// delete user
		v1Group.DELETE("/user/:email_address", func (c *gin.Context) {
			logger.Log.Printf("Delete a user is starting...")
			num4++
			// Time something.
			t := d.NewTiming()

			// Increment a counter.
			//d.Increment("foo.counter")
			d.Count("foo.counter", num4)

			email, valid := c.Params.Get("email_address")
			if !valid {
				c.JSON(404, gin.H{"error": "email is not exist"})
				logger.Log.Printf(err.Error())
				return
			}

			t2 := d.NewTiming()

			if err = dao.DB.Where("email_address=?", email).Delete(models.User{}).Error; err!=nil {
				c.JSON(404, gin.H{"error": err.Error()})
				logger.Log.Printf(err.Error())
			} else {
				c.JSON(http.StatusOK, gin.H{email: "Deleted"})
			}

			t2.Send("db_response_time")
			t.Send("api_response_time")
			logger.Log.Printf("Delete a user is done...")
		})

		// get a question
		v1Group.GET("/question/:question_id", func(c *gin.Context) {
			logger.Log.Printf("Get a question is starting...")
			num5++
			// Time something.
			t := d.NewTiming()

			// Increment a counter.
			//d.Increment("foo.counter")
			d.Count("foo.counter", num5)

			// get the question_id
			questionId, valid := c.Params.Get("question_id")
			if !valid {
				c.JSON(204, gin.H{"error": "cannot get the question_id"})
				logger.Log.Printf(err.Error())
				return
			}

			t2 := d.NewTiming()

			// check the question_id exist or not
			var question models.Question
			if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}

			// now, we can get a question info
			// first, get categories
			// get from qc
			var qcArr []models.QuestionCategory
			if err = dao.DB.Where("question_id=?", questionId).Find(&qcArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}
			// get from category table
			var cateArr []models.Category
			for i := range qcArr {
				var category models.Category
				if err = dao.DB.Where("id=?", qcArr[i].CategoryID).Find(&category).Error; err != nil {
					c.JSON(404, gin.H{"error": "The question_id is not exist"})
					logger.Log.Printf(err.Error())
					return
				}
				cateArr = append(cateArr, category)
			}
			// second, get answers
			var answerArr []models.Answer
			if err = dao.DB.Where("question_id=?", questionId).Find(&answerArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}
			// add the file info into the answer
			var fileQuestionArr []models.FileQuestion
			if err = dao.DB.Where("question_id=?", questionId).Find(&fileQuestionArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The answer id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}

			t2.Send("db_response_time")

			question.AnswerArr = answerArr
			question.CategoryArr = cateArr
			question.FileArr = fileQuestionArr
			// then, print out
			c.JSON(http.StatusOK, gin.H{
				"question": question,
			})

			t.Send("api_response_time")
			logger.Log.Printf("Get a question is done...")
		})

		// get all questions
		v1Group.GET("/question", func(c *gin.Context) {
			logger.Log.Printf("Get all questions is starting...")
			num6++
			// Time something.
			t := d.NewTiming()

			// Increment a counter.
			//d.Increment("foo.counter")
			d.Count("foo.counter", num6)

			// get all the questions first
			var questionArr []models.Question
			if err = dao.DB.Find(&questionArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}

			// get the answers and cates based on each question
			for i := range questionArr {

				t2 := d.NewTiming()

				// first, get categories
				// get from qc
				var qcArr []models.QuestionCategory
				if err = dao.DB.Where("question_id=?", questionArr[i].ID).Find(&qcArr).Error; err!=nil {
					c.JSON(404, gin.H{"error": "The question_id is not exist"})
					logger.Log.Printf(err.Error())
					return
				}
				// get from category table
				var cateArr []models.Category
				for i := range qcArr {
					var category models.Category
					if err = dao.DB.Where("id=?", qcArr[i].CategoryID).Find(&category).Error; err != nil {
						c.JSON(404, gin.H{"error": "The question_id is not exist"})
						logger.Log.Printf(err.Error())
						return
					}
					cateArr = append(cateArr, category)
				}
				// second, get answers
				var answerArr []models.Answer
				if err = dao.DB.Where("question_id=?", questionArr[i].ID).Find(&answerArr).Error; err!=nil {
					c.JSON(404, gin.H{"error": "The question_id is not exist"})
					logger.Log.Printf(err.Error())
					return
				}

				// add the file info into the answer
				for j := range answerArr {
					var fileAnswerArr []models.FileAnswer
					if err = dao.DB.Where("answer_id=?", answerArr[j].ID).Find(&fileAnswerArr).Error; err != nil {
						c.JSON(404, gin.H{"error": "The answer id is not exist"})
						logger.Log.Printf(err.Error())
						return
					}
					answerArr[j].FileArr = fileAnswerArr
				}
				// add the file info into the answer
				var fileQuestionArr []models.FileQuestion
				if err = dao.DB.Where("question_id=?", questionArr[i].ID).Find(&fileQuestionArr).Error; err!=nil {
					c.JSON(404, gin.H{"error": "The answer id is not exist"})
					logger.Log.Printf(err.Error())
					return
				}

				t2.Send("db_response_time")

				questionArr[i].CategoryArr = cateArr
				questionArr[i].AnswerArr = answerArr
				questionArr[i].FileArr = fileQuestionArr
			}
			c.JSON(http.StatusOK, gin.H{
				"questions": questionArr,
			})

			t.Send("api_response_time")
			logger.Log.Printf("Get all questions is done...")
		})

		// get a question's answer
		v1Group.GET("/question/:question_id/answer/:answer_id", func(c *gin.Context) {
			logger.Log.Printf("Get an answer is starting...")
			num7++
			// Time something.
			t := d.NewTiming()

			// Increment a counter.
			//d.Increment("foo.counter")
			d.Count("foo.counter", num7)

			// get the question_id and answer_id
			questionId, valid := c.Params.Get("question_id")
			if !valid {
				c.JSON(204, gin.H{"error": "cannot get the question_id"})
				logger.Log.Printf(err.Error())
				return
			}
			answerId, valid := c.Params.Get("answer_id")
			if !valid {
				c.JSON(204, gin.H{"error": "cannot get the answer_id"})
				logger.Log.Printf(err.Error())
				return
			}

			t2 := d.NewTiming()

			// check the question_id and answer_id exist or not
			var answer models.Answer
			if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The answer_id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}
			if answer.QuestionID != questionId {
				c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
				logger.Log.Printf(err.Error())
				return
			}

			// add the file info into the answer
			var fileAnswerArr []models.FileAnswer
			if err = dao.DB.Where("answer_id=?", answerId).Find(&fileAnswerArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The answer id is not exist"})
				logger.Log.Printf(err.Error())
				return
			}

			t2.Send("db_response_time")

			answer.FileArr = fileAnswerArr

			// now, we can get the answer
			c.JSON(200, gin.H{
				"answer": answer,
			})

			t.Send("api_response_time")
			logger.Log.Printf("Get an answer is done...")
		})
	}

	// Group using gin.BasicAuth() middleware
	// gin.Accounts is a shortcut for map[string]string
	authorized := r.Group("/v1", function.BasicAuth())

	// basic authorized to get a user info
	authorized.GET("/user_auth/self", func(c *gin.Context) {
		logger.Log.Printf("Get a user is starting...")
		num8++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num8)

		//email := c.Params.ByName("email_address")
		email := function.FetchUsername
		var user models.User

		t2 := d.NewTiming()

		err := dao.DB.Where("email_address=?", email).First(&user).Error

		t2.Send("db_response_time")

		if err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
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

		t.Send("api_response_time")
		logger.Log.Printf("Get a user is done...")
	})

	// update a question
	authorized.PUT("/question/:question_id", func(c *gin.Context) {
		logger.Log.Printf("Update a question is starting...")
		num9++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num9)

		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email

		// get the question_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			logger.Log.Printf(err.Error())
			return
		}

		// check the question_id exist or not
		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}

		// check authenticated or not
		if question.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the question does not belong to this user"})
			logger.Log.Printf("the question does not belong to this user")
			return
		}

		// update question
/*		questionText := question.QuestionText
		cateArr := question.CategoryArr*/
		var newQuestion models.Question
		var flag1 bool = false
		c.BindJSON(&newQuestion)
		// categories := question.CategoryArr
		// check content is empty or not
		if newQuestion.QuestionText == "" && newQuestion.CategoryArr == nil{
			c.JSON(204, gin.H{"error": "no content"})
			logger.Log.Printf("no content")
			return
		}
		if newQuestion.QuestionText != "" {
			question.QuestionText = newQuestion.QuestionText
		}
		if newQuestion.CategoryArr != nil {
			question.CategoryArr = newQuestion.CategoryArr
		} else {
			flag1 = true	// which means do not need to update the categories
		}
		categories := question.CategoryArr
		question.QuestionUpdated = time.Now()

		if !flag1 {
			// delete all the categories in question in qc table
			// check the question has categories or not
			var flag2 bool = true // have category
			var qc models.QuestionCategory

			t2 := d.NewTiming()

			if err = dao.DB.Where("question_id=?", question.ID).First(&qc).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"msg": "The question does not have category"})
				flag2 = false // no category
			}
			// if the question has category, delete in qc table
			if flag2 {
				if err = dao.DB.Where("question_id=?", question.ID).Delete(models.QuestionCategory{}).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					logger.Log.Printf(err.Error())
				}
			}

			t2.Send("db_response_time")

			val := function.CheckCategoryDuplicate(categories)
			if !val {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate categories!"})
				logger.Log.Printf("error: Duplicate catefories")
				return
			}

			// update categories
			for i := range categories {
				var category models.Category

				t3 := d.NewTiming()

				dao.DB.Where("category_name=?", categories[i].CategoryName).First(&category)

				t3.Send("db_response_time")

				if category.ID == "" {
					category.ID = newuuid.New().String()
					category.CategoryName = categories[i].CategoryName
					if err := dao.DB.Create(&category).Error; err != nil {
						c.JSON(404, gin.H{"error": err.Error()})
						logger.Log.Printf(err.Error())
					}
				}
				categories[i].ID = category.ID
			}
		}

		t4 := d.NewTiming()

		// send into the DB, and then response
		if err := dao.DB.Save(&question).Error;err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Updated a question"})
		}

		t4.Send("db_response_time")

		if !flag1 {
			for i := range categories {
				var qc models.QuestionCategory
				qc.CategoryID = categories[i].ID
				qc.QuestionID = question.ID
				if err := dao.DB.Create(&qc).Error; err != nil {
					c.JSON(404, gin.H{"error": err.Error()})
					logger.Log.Printf(err.Error())
				}
			}
		}

		t.Send("api_response_time")
		logger.Log.Printf("Update a question is done...")
	})

	// Delete a question
	authorized.DELETE("/question/:question_id", func(c *gin.Context) {
		logger.Log.Printf("Delete a question is starting...")
		num10++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num10)

		email := function.FetchUsername
		var user models.User

		t2 := d.NewTiming()

		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email

		t2.Send("db_response_time")

		// get the question_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			logger.Log.Printf("error: cannot get hte question_id")
			return
		}

		// check the question_id exist or not
		var question models.Question

		t3 := d.NewTiming()

		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}


		t3.Send("db_response_time")

		// check authenticated or not
		if question.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the question does not belong to this user"})
			logger.Log.Printf("error: this question does not belong to this user")
			return
		}

		// check the question has answers or not
		var answer models.Answer

		t4 := d.NewTiming()

		if err = dao.DB.Where("question_id=?", question.ID).First(&answer).Error; err==nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The answer is exist, user cannot delete the question"})
			logger.Log.Printf(err.Error())
			return
		}

		// check the question has categories or not
		var flag bool = true // have category
		var qc models.QuestionCategory
		if err = dao.DB.Where("question_id=?", question.ID).First(&qc).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "The question does not have category"})
			logger.Log.Printf(err.Error())
			flag = false // no category
		}
		// now, the user can delete the question without any answers
		// delete in qc table
		// if the question has category, delete
		if flag {
			if err = dao.DB.Where("question_id=?", question.ID).Delete(models.QuestionCategory{}).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				logger.Log.Printf(err.Error())
			}
		}

		// delete the file if exist
		var fileQuestionArr []models.FileQuestion
		if err = dao.DB.Where("question_id=?", questionId).Find(&fileQuestionArr).Error; err!=nil {
			c.JSON(200, gin.H{"msg": "cannot find the file for this answer"})
			logger.Log.Printf(err.Error())
			return
		}
		// delete the file in mysql
		if err = dao.DB.Where("question_id=?", questionId).Delete(models.FileQuestion{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an question's file in mysql"})
		}


		t4.Send("db_response_time")

		// delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"

		t5 := d.NewTiming()

		for _, fileQuestion := range fileQuestionArr {
			DeleteFile(S3Bucket, fileQuestion.S3ObjectName)
		}

		t5.Send("S3_response_time")

		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})

		// then, delete the question
		if err = dao.DB.Where("id=?", question.ID).Delete(models.Question{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted a question"})
		}

		t.Send("api_response_time")
		logger.Log.Printf("Delete a question is done...")
	})

	// Delete a question's answer, delete the file if exist
	authorized.DELETE("/question/:question_id/answer/:answer_id", func(c *gin.Context) {
		logger.Log.Printf("Delete an answer is starting...")
		num11++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num11)

		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email

		// get the question_id and answer_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			logger.Log.Printf("error: cannot get the question_id")
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the answer_id"})
			logger.Log.Printf("error: cannot get the answer_id")
			return
		}

		// check the question_id and answer_id exist or not
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		if answer.QuestionID != questionId {
			c.JSON(404, gin.H{"error": "The question_id and answer_id are not matched"})
			logger.Log.Printf("error: the question_id and answer_id are not matched")
			return
		}

		// check authenticated or not
		if answer.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			logger.Log.Printf("error: the answer does not belong to this user")
			return
		}


		t2 := d.NewTiming()

		// delete the file if exist
		var fileAnswerArr []models.FileAnswer
		if err = dao.DB.Where("answer_id=?", answerId).Find(&fileAnswerArr).Error; err!=nil {
			c.JSON(200, gin.H{"msg": "cannot find the file for this answer"})
			logger.Log.Printf(err.Error())
			return
		}
		// delete the file in mysql
		if err = dao.DB.Where("answer_id=?", answerId).Delete(models.FileAnswer{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an answer's file in mysql"})
		}

		t2.Send("db_response_time")

		t3 := d.NewTiming()

		// delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		for _, fileAnswer := range fileAnswerArr {
			DeleteFile(S3Bucket, fileAnswer.S3ObjectName)
		}
		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})

		t3.Send("S3_resonse_time")

		// Start to delete answer
		if err = dao.DB.Where("id=?", answer.ID).Delete(models.Answer{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted a question's answer"})
		}

		t.Send("api_response_time")
		logger.Log.Printf("Delete an answer is done...")

		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		var questionUser models.User
		if err = dao.DB.Where("id=?", question.UserID).First(&questionUser).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The user_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}

		msg := "Delete an answer," + questionId + "," + *(questionUser.EmailAddress) + "," + answerId + "," + answer.AnswerText
		snsPublish(msg, "arn:aws:sns:us-east-1:931397163240:fall2020")
	})

	// Update answer
	authorized.PUT("/question/:question_id/answer/:answer_id", func(c *gin.Context) {
		logger.Log.Printf("Update an answer is starting...")
		num12++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num12)

		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email

		// get the question_id and answer_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			logger.Log.Printf("error: cannot get the question_id")
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the answer_id"})
			logger.Log.Printf("error: cannot get hte answer_id")
			return
		}

		// check the question_id and answer_id exist or not
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		if answer.QuestionID != questionId {
			c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
			logger.Log.Printf("error: the question id and answer id are not matched")
			return
		}

		// check authenticated or not
		if answer.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			logger.Log.Printf("error: the answer does not belong to this user")
			return
		}

		// update answer
		c.BindJSON(&answer)
		// check content is empty or not
		if answer.AnswerText == "" {
			c.JSON(204, gin.H{"error": "no content"})
			logger.Log.Printf("error: no content")
			return
		}

		answer.AnswerUpdated = time.Now()

		t2 := d.NewTiming()

		// send into the DB, and then response
		if err := dao.DB.Save(&answer).Error;err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg": "Updated an answer",
			})
		}

		t2.Send("db_response_time")
		t.Send("api_response_time")
		logger.Log.Printf("Update an answer is done...")

		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		var questionUser models.User
		if err = dao.DB.Where("id=?", question.UserID).First(&questionUser).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The user_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}

		msg := "Update an answer," + questionId + "," + *(questionUser.EmailAddress) + "," + answerId + "," + answer.AnswerText
		snsPublish(msg, "arn:aws:sns:us-east-1:931397163240:fall2020")
	})

	// Post answer
	authorized.POST("/question/:question_id/answer", func(c *gin.Context) {
		logger.Log.Printf("Post an answer is starting...")
		num13++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num13)

		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email
		id, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "question id is not exist"})
			logger.Log.Printf("error: question id is not exit")
			return
		}
		var answer models.Answer
		c.BindJSON(&answer)
		answer.ID = newuuid.New().String()
		answer.QuestionID = id
		answer.UserID = user.ID
		answer.AnswerCreated = time.Now()
		answer.AnswerUpdated = time.Now()

		t2 := d.NewTiming()

		// send into the DB, and then response
		if err := dao.DB.Create(&answer).Error;err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{
				"answer_id": answer.ID,
				"question_id": answer.QuestionID,
				"created_timestamp": answer.AnswerCreated,
				"updated_timestamp": answer.AnswerUpdated,
				"user_id": answer.UserID,
				"answer_text": answer.AnswerText,
			})
		}

		t2.Send("db_response_time")
		t.Send("api_response_time")
		logger.Log.Printf("Post an answer is done...")

		var question models.Question
		if err = dao.DB.Where("id=?", id).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		var questionUser models.User
		if err = dao.DB.Where("id=?", question.UserID).First(&questionUser).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The user_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}

		msg := "Create an answer," + id + "," + *(questionUser.EmailAddress) + "," + answer.ID + "," + answer.AnswerText
		snsPublish(msg, "arn:aws:sns:us-east-1:931397163240:fall2020")
	})

	// post a new question
	authorized.POST("/question/", func(c *gin.Context) {
		logger.Log.Printf("Post a question is starting...")
		num14++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num14)

		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email

		var question models.Question
		c.BindJSON(&question)
		// auto set the id, create time and update time
		question.ID = newuuid.New().String()
		question.QuestionCreated = time.Now()
		question.QuestionUpdated = time.Now()
		question.UserID = user.ID
		categories := question.CategoryArr
		val := function.CheckCategoryDuplicate(categories)
		if !val {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate categories!"})
			logger.Log.Printf("error: duplicate catefories")
			return
		}

		t2 := d.NewTiming()

		// update categories
		for i := range categories {
			var category models.Category
			dao.DB.Where("category_name=?", categories[i].CategoryName).First(&category)
			if category.ID == "" {
				category.ID = newuuid.New().String()
				category.CategoryName = categories[i].CategoryName
				if err := dao.DB.Create(&category).Error; err != nil {
					c.JSON(http.StatusOK, gin.H{"error": err.Error()})
					logger.Log.Printf(err.Error())
				}
			}
			categories[i].ID = category.ID
		}

		// send into the DB, and then response
		if err := dao.DB.Create(&question).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{
				"question_id": question.ID,
				"created_timestamp": question.QuestionCreated,
				"updated_timestamp": question.QuestionUpdated,
				"user_id": question.UserID,
				"question_text": question.QuestionText,
				"categories": categories,
				"answers": nil,
			})
		}

		for i := range categories {
			var qc models.QuestionCategory
			qc.CategoryID = categories[i].ID
			qc.QuestionID = question.ID
			if err := dao.DB.Create(&qc).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				logger.Log.Printf(err.Error())
			}
		}

		t2.Send("db_response_time")
		t.Send("api_response_time")
		logger.Log.Printf("Post a question is done...")
	})

	// update user
	authorized.PUT("/user/self", func (c *gin.Context) {
		logger.Log.Printf("Update a user is starting...")
		num15++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num15)

		/*email, valid := c.Params.Get("email_address")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "email address is not exist"})
			return
		}*/
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email
		id := user.ID
		accountCreate := user.AccountCreated
		password := user.Password
		c.BindJSON(&user)
		user.AccountUpdated = time.Now()

		if password != user.Password {
			code := function.CheckPassword(user.Password, 2)
			switch code {
			case -1: c.JSON(http.StatusOK, gin.H{"error": "the password is too short, please use at least 8 char"})
				logger.Log.Printf("error: password too short")
				return
			case 0: c.JSON(http.StatusOK, gin.H{"error": "the password is too week, please use letters, digits and special char"})
				logger.Log.Printf("error: password too week")
				return
			}

			var pass = function.HashAndSalt(function.GetPwd(user.Password))
			user.Password = pass
		}

		/*
		user.ID = id
		user.AccountCreated = accountCreate*/

		t2 := d.NewTiming()

		// if user wants to change the email
		if email != *user.EmailAddress || id != user.ID || accountCreate != user.AccountCreated{
			c.JSON(400, gin.H{"error": "The user cannot change the email address, id or create time"})
			logger.Log.Printf("error: user cannot change the email, id or create time")
			return
		}

		if err = dao.DB.Save(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
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

		t2.Send("db_response_time")
		t.Send("api_response_time")
		logger.Log.Printf("Update a user is done...")
	})

	// delete a file in a question
	authorized.DELETE("/question/:question_id/file/:file_id", func(c *gin.Context) {
		logger.Log.Printf("Delete a file in question is starting...")
		num16++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num16)

		// 1. authen the log in user is the owner of question
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email
		// get the question_id and file_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			logger.Log.Printf("error: cannot get the question id")
			return
		}
		fileId, valid := c.Params.Get("file_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the file_id"})
			logger.Log.Printf("error: cannot get the file id")
			return
		}

		t2 := d.NewTiming()

		// check the question_id exist or not
		var fileQuestion models.FileQuestion
		if err = dao.DB.Where("id=?", fileId).First(&fileQuestion).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The file_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question is not exist"})
			logger.Log.Printf(err.Error())
			return
		}

		t2.Send("db_response_time")

		// check authenticated or not
		if question.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			logger.Log.Printf("error: the answer does not belong to this user")
			return
		}
		// check file id is matched with question id
		if fileQuestion.QuestionID != questionId {
			c.JSON(401, gin.H{"error": "the question is not matched with the file"})
			logger.Log.Printf("error: the question is not mathced with the file")
			return
		}


		t3 := d.NewTiming()

		// 2. delete the file in mysql
		if err = dao.DB.Where("id=?", fileId).Delete(models.FileQuestion{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an question's file in mysql"})
		}

		t3.Send("db_response_time")
		t4 := d.NewTiming()

		// 3. delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		DeleteFile(S3Bucket, fileQuestion.S3ObjectName)

		t4.Send("S3_response_time")

		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})

		t.Send("api_response_time")
		logger.Log.Printf("Delete a file in a question is done...")
	})

	// delete a file in an answer
	authorized.DELETE("/question/:question_id/answer/:answer_id/file/:file_id", func(c *gin.Context) {
		logger.Log.Printf("Delete a file in an answer is starting...")
		num17++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num17)

		// 1. authen the log in user is the owner of answer
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email
		// get the question_id and answer_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			logger.Log.Printf("error: cannot get the question id")
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the answer_id"})
			logger.Log.Printf("error: cannot get the answer id")
			return
		}
		fileId, valid := c.Params.Get("file_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the file_id"})
			logger.Log.Printf("error: cannot get the file id")
			return
		}

		t2 := d.NewTiming()

		// check the question_id and answer_id exist or not
		var fileAnswer models.FileAnswer
		if err = dao.DB.Where("id=?", fileId).First(&fileAnswer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The file_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			logger.Log.Printf(err.Error())
			return
		}

		t2.Send("db_response_time")

		if answer.QuestionID != questionId {
			c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
			logger.Log.Printf("error: the question id and answer id are not matched")
			return
		}
		// check authenticated or not
		if answer.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			logger.Log.Printf("error: the answer does not belong to this user")
			return
		}
		// check answer id is matched with question id
		if answer.QuestionID != questionId {
			c.JSON(401, gin.H{"error": "the answer is not matched with the question"})
			logger.Log.Printf("error: the answer is not mathced with the question")
			return
		}
		// chekc file id is mathced with the answer id
		if fileAnswer.AnswerID != answerId {
			c.JSON(401, gin.H{"error": "the file is not belonging to this answer"})
			logger.Log.Printf("error: the file is not belonging to the answer")
			return
		}

		t3 := d.NewTiming()

		// 2. delete the file in mysql
		if err = dao.DB.Where("id=?", fileId).Delete(models.FileAnswer{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an answer's file in mysql"})
		}

		t3.Send("db_response_time")
		t4 := d.NewTiming()

		// 3. delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		DeleteFile(S3Bucket, fileAnswer.S3ObjectName)

		t4.Send("S3_response_time")

		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})

		t.Send("api_response_time")
		logger.Log.Printf("Delete a file in an answer is done...")
	})

	// post a file to an answer
	authorized.POST("/question/:question_id/answer/:answer_id/file", func(c *gin.Context) {
		logger.Log.Printf("Post a file to the answer is starting...")
		num18++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num18)

		//AccessKeyID := GetEnvWithKey("AWS_ACCESS_KEY_ID")
		//SecretAccessKey := GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
		//S3Region := GetEnvWithKey("AWS_REGION")
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"

		// 1. authenticate the user is the owner of the question
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(404, gin.H{"error": "question is not exist"})
			logger.Log.Printf("error: question is not exist")
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "question is not exist"})
			return
		}
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}
		//
		if answer.UserID != user.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this question id is not belong to the user"})
			logger.Log.Printf("error: question id is not belong to user")
			return
		}
		//
		if answer.QuestionID != questionId {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this answer id is not matched with the question id"})
			logger.Log.Printf("error: answer id is not matched with question id")
			return
		}

		// 2. upload the image onto the AWS S3
		fileHeader, _ := c.FormFile("photo")
		f, _ := fileHeader.Open()
		var size int64 = fileHeader.Size

		buffer := make([]byte, size)
		f.Read(buffer)
/*
		creds := credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, "")
		s, _ := session.NewSession(&aws.Config{
			Region:      aws.String(S3Region),
			Credentials: creds,
		})
*/

		t2 := d.NewTiming()

		s := initSession()
		var fileAnswer models.FileAnswer
		fileAnswer.ID = newuuid.New().String()
		fileAnswer.S3ObjectName = answer.ID + "/" + fileAnswer.ID + "/" + fileHeader.Filename
		_, _ = s3.New(s).PutObject(&s3.PutObjectInput{
			Bucket:             aws.String(S3Bucket),
			Key:                aws.String(fileAnswer.S3ObjectName),
			ACL:                aws.String("private"),
			Body:               bytes.NewReader(buffer),
			ContentLength:      aws.Int64(size),
			ContentType:        aws.String(http.DetectContentType(buffer)),
			ContentDisposition: aws.String("attachment"),
		})

		t2.Send("S3_response_time")

		// 3. get the image info from S3
		var metadata models.Metadata
		metadata = GetObjectMetaData(S3Bucket, fileAnswer.S3ObjectName)

		// 4. save the info into DB FileQ
		fileAnswer.CreateDate = time.Now()
		fileAnswer.AnswerID = answer.ID
		fileAnswer.FileName = fileHeader.Filename
		fileAnswer.AcceptRanges = *metadata.AcceptRanges
		fileAnswer.ContentLength = strconv.FormatInt(*metadata.ContentLength, 10)
		fileAnswer.ContentType = *metadata.ContentType
		fileAnswer.ETag = *metadata.ETag

		t3 := d.NewTiming()

		// send into the DB, and then response
		if err := dao.DB.Create(&fileAnswer).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{
				"file_name": fileAnswer.FileName,
				"s3_object_name": fileAnswer.S3ObjectName,
				"file_id": fileAnswer.ID,
				"created_date": fileAnswer.CreateDate,
			})
		}

		t3.Send("db_response_time")
		t.Send("api_response_time")
		logger.Log.Printf("Post a file to the answer is done...")
	})

	// post a file to a question
	authorized.POST("/question/:question_id/file", func(c *gin.Context) {
		logger.Log.Printf("Post a file to the question is starting...")
		num19++
		// Time something.
		t := d.NewTiming()

		// Increment a counter.
		//d.Increment("foo.counter")
		d.Count("foo.counter", num19)

		//AccessKeyID := GetEnvWithKey("AWS_ACCESS_KEY_ID")
		//SecretAccessKey := GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
		//S3Region := GetEnvWithKey("AWS_REGION")
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"

		// 1. authenticate the user is the owner of the question
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}	// get the user info based on email
		id, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "question is not exist"})
			logger.Log.Printf("error: question is not exist")
			return
		}
		var question models.Question
		if err = dao.DB.Where("id=?", id).First(&question).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
			return
		}
		//
		if question.UserID != user.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this question id is not belong to the user"})
			logger.Log.Printf("error: question id is not belong to the user")
			return
		}

		// 2. upload the image onto the AWS S3
		fileHeader, _ := c.FormFile("photo")
		f, _ := fileHeader.Open()
		var size int64 = fileHeader.Size

		buffer := make([]byte, size)
		f.Read(buffer)
/*
		creds := credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, "")
		s, _ := session.NewSession(&aws.Config{
			Region:      aws.String(S3Region),
			Credentials: creds,
		})
*/
		t2 := d.NewTiming()

		s := initSession()
		var fileQuestion models.FileQuestion
		fileQuestion.ID = newuuid.New().String()
		fileQuestion.S3ObjectName = question.ID + "/" + fileQuestion.ID + "/" + fileHeader.Filename
		_, _ = s3.New(s).PutObject(&s3.PutObjectInput{
			Bucket:             aws.String(S3Bucket),
			Key:                aws.String(fileQuestion.S3ObjectName),
			ACL:                aws.String("private"),
			Body:               bytes.NewReader(buffer),
			ContentLength:      aws.Int64(size),
			ContentType:        aws.String(http.DetectContentType(buffer)),
			ContentDisposition: aws.String("attachment"),
		})

		t2.Send("S3_response_time")

		// 3. get the image info from S3
		var metadata models.Metadata
		metadata = GetObjectMetaData(S3Bucket, fileQuestion.S3ObjectName)

		// 4. save the info into DB FileQ
		fileQuestion.CreateDate = time.Now()
		fileQuestion.QuestionID = question.ID
		fileQuestion.FileName = fileHeader.Filename
		fileQuestion.AcceptRanges = *metadata.AcceptRanges
		fileQuestion.ContentLength = strconv.FormatInt(*metadata.ContentLength, 10)
		fileQuestion.ContentType = *metadata.ContentType
		fileQuestion.ETag = *metadata.ETag

		t3 := d.NewTiming()

		// send into the DB, and then response
		if err := dao.DB.Create(&fileQuestion).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			logger.Log.Printf(err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{
				"file_name": fileQuestion.FileName,
				"s3_object_name": fileQuestion.S3ObjectName,
				"file_id": fileQuestion.ID,
				"created_date": fileQuestion.CreateDate,
			})
		}

		t3.Send("db_rsponse_time")
		t.Send("api_response_time")
		logger.Log.Printf("Post a file to the question is done...")
	})

	r.Run(":9090")
}

//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}
/*
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}

func ConnectAws() *session.Session {
	AccessKeyID = GetEnvWithKey("AWS_ACCESS_KEY_ID")
	SecretAccessKey = GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
	MyRegion = GetEnvWithKey("AWS_REGION")

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})

	if err != nil {
		panic(err)
	}

	return sess
}
*/

func GetObjectMetaData(bucketName, objectName string) models.Metadata{
	sess, err := session.NewSessionWithOptions(session.Options{
		// Specify profile to load for the session's config
		Profile: "dev",

		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},

		// Force enable Shared Config support
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		log.Fatalf("can't load the aws session")
	}
	svc := s3.New(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	result, err := svc.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatalf(err.Error())
		}
	}

	fmt.Println(result)

	return models.Metadata{
		AcceptRanges:  result.AcceptRanges,
		ContentLength: result.ContentLength,
		ContentType:   result.ContentType,
		ETag:          result.ETag,
	}
}

func DeleteFile(bucketName, filename string)  {
	sess, err := session.NewSessionWithOptions(session.Options{
		// Specify profile to load for the session's config
		Profile: "dev",

		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},

		// Force enable Shared Config support
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		log.Fatalf("can't load the aws session")
	}
	svc := s3.New(sess)

	if _, err := svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucketName), Key: aws.String(filename)}); err != nil {
		fmt.Printf("Unable to delete object %q from bucket %q, %v", filename, bucketName, err)
		return
	}

	_ = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})

	if err != nil {
		// Print the error and exit.
		fmt.Printf("Unable to delete %q to %q, %v", filename, bucketName, err)
		return
	}

	fmt.Printf("Successfully deleted %q to %q\n", filename, bucketName)

}

var sess *session.Session

func initSession() *session.Session {
	if sess == nil {
		newSess, err := session.NewSessionWithOptions(session.Options{
			// Specify profile to load for the session's config
			Profile: "prod",

			// Provide SDK Config options, such as Region.
			Config: aws.Config{
				Region: aws.String("us-east-1"),
			},

			// Force enable Shared Config support
			SharedConfigState: session.SharedConfigEnable,
		})

		if err != nil {
			log.Fatalf("can't load the aws session")
		}else{
			sess = newSess
		}
	}

	return sess
}

var sess1 *session.Session

func snsPublish(msg, topicARN string) {
	sess1, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	if err != nil {
		fmt.Println("NewSession error:", err)
		return
	}

	client := sns.New(sess1)
	input := &sns.PublishInput{
		Message:  aws.String(msg),
		TopicArn: aws.String(topicARN),
	}

	_, err = client.Publish(input)
	if err != nil {
		fmt.Println("Publish error:", err)
		return
	}

	//fmt.Println(result)
}