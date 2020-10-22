package main

import (
	"bytes"
	"fmt"
	"gin_demo/dao"
	"gin_demo/function"
	"gin_demo/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	//"github.com/joho/godotenv"
)

var err error
var AccessKeyID string
var SecretAccessKey string
var MyRegion string
var MyBucket string
var filepath string

func main() {
	// connect to the DB
	dao.DB, err = gorm.Open("mysql", dao.DbURL(dao.BuildDBConfig()))
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
/*
	LoadEnv()
	sess := ConnectAws()
	r.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})
*/
	r.GET("/hello", func (c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	v1Group := r.Group("/v1")
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
			// ???????
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
					"id": user.ID,
					"create time": user.AccountCreated,
					"update time": user.AccountUpdated,
					"email": user.EmailAddress,
					"first name": user.FirstName,
					"last name": user.LastName,
				})
			}
		})

		// view user
		v1Group.GET("/user/", func (c *gin.Context) {
			var userList []models.User
			if err = dao.DB.Find(&userList).Error;err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"msg": "success",
					"data": userList,
				})
			}
		})

		// get a user info by id
		v1Group.GET("/user/:id", func(c *gin.Context) {
			id := c.Params.ByName("id")
			var user models.User
			err := dao.DB.Where("id=?", id).First(&user).Error
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"data": user,
				})
			}
		})

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

		// get a question
		v1Group.GET("/question/:question_id", func(c *gin.Context) {
			// get the question_id
			questionId, valid := c.Params.Get("question_id")
			if !valid {
				c.JSON(204, gin.H{"error": "cannot get the question_id"})
				return
			}
			// check the question_id exist or not
			var question models.Question
			if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				return
			}

			// now, we can get a question info
			// first, get categories
			// get from qc
			var qcArr []models.QuestionCategory
			if err = dao.DB.Where("question_id=?", questionId).Find(&qcArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				return
			}
			// get from category table
			var cateArr []models.Category
			for i := range qcArr {
				var category models.Category
				if err = dao.DB.Where("id=?", qcArr[i].CategoryID).Find(&category).Error; err != nil {
					c.JSON(404, gin.H{"error": "The question_id is not exist"})
					return
				}
				cateArr = append(cateArr, category)
			}
			// second, get answers
			var answerArr []models.Answer
			if err = dao.DB.Where("question_id=?", questionId).Find(&answerArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				return
			}
			// add the file info into the answer
			var fileQuestionArr []models.FileQuestion
			if err = dao.DB.Where("question_id=?", questionId).Find(&fileQuestionArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The answer id is not exist"})
				return
			}
			question.AnswerArr = answerArr
			question.CategoryArr = cateArr
			question.FileArr = fileQuestionArr
			// then, print out
			c.JSON(http.StatusOK, gin.H{
				"question": question,
			})
		})

		// get all questions
		v1Group.GET("/question", func(c *gin.Context) {
			// get all the questions first
			var questionArr []models.Question
			if err = dao.DB.Find(&questionArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The question_id is not exist"})
				return
			}

			// get the answers and cates based on each question
			for i := range questionArr {
				// first, get categories
				// get from qc
				var qcArr []models.QuestionCategory
				if err = dao.DB.Where("question_id=?", questionArr[i].ID).Find(&qcArr).Error; err!=nil {
					c.JSON(404, gin.H{"error": "The question_id is not exist"})
					return
				}
				// get from category table
				var cateArr []models.Category
				for i := range qcArr {
					var category models.Category
					if err = dao.DB.Where("id=?", qcArr[i].CategoryID).Find(&category).Error; err != nil {
						c.JSON(404, gin.H{"error": "The question_id is not exist"})
						return
					}
					cateArr = append(cateArr, category)
				}
				// second, get answers
				var answerArr []models.Answer
				if err = dao.DB.Where("question_id=?", questionArr[i].ID).Find(&answerArr).Error; err!=nil {
					c.JSON(404, gin.H{"error": "The question_id is not exist"})
					return
				}
				// add the file info into the answer
				var fileQuestionArr []models.FileQuestion
				if err = dao.DB.Where("question_id=?", questionArr[i].ID).Find(&fileQuestionArr).Error; err!=nil {
					c.JSON(404, gin.H{"error": "The answer id is not exist"})
					return
				}
				questionArr[i].CategoryArr = cateArr
				questionArr[i].AnswerArr = answerArr
				questionArr[i].FileArr = fileQuestionArr
			}
			c.JSON(http.StatusOK, gin.H{
				"questions": questionArr,
			})
		})

		// get a question's answer
		v1Group.GET("/question/:question_id/answer/:answer_id", func(c *gin.Context) {
			// get the question_id and answer_id
			questionId, valid := c.Params.Get("question_id")
			if !valid {
				c.JSON(204, gin.H{"error": "cannot get the question_id"})
				return
			}
			answerId, valid := c.Params.Get("answer_id")
			if !valid {
				c.JSON(204, gin.H{"error": "cannot get the answer_id"})
				return
			}

			// check the question_id and answer_id exist or not
			var answer models.Answer
			if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The answer_id is not exist"})
				return
			}
			if answer.QuestionID != questionId {
				c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
				return
			}

			// add the file info into the answer
			var fileAnswerArr []models.FileAnswer
			if err = dao.DB.Where("answer_id=?", answerId).Find(&fileAnswerArr).Error; err!=nil {
				c.JSON(404, gin.H{"error": "The answer id is not exist"})
				return
			}
			answer.FileArr = fileAnswerArr

			// now, we can get the answer
			c.JSON(200, gin.H{
				"answer": answer,
			})
		})
	}

	// Group using gin.BasicAuth() middleware
	// gin.Accounts is a shortcut for map[string]string
	authorized := r.Group("/v1", function.BasicAuth())

	// basic authorized to get a user info
	authorized.GET("/user_auth/self", func(c *gin.Context) {
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

	// update a question
	authorized.PUT("/question/:question_id", func(c *gin.Context) {
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			return
		}	// get the user info based on email

		// get the question_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			return
		}

		// check the question_id exist or not
		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question_id is not exist"})
			return
		}

		// check authenticated or not
		if question.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the question does not belong to this user"})
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
			if err = dao.DB.Where("question_id=?", question.ID).First(&qc).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"msg": "The question does not have category"})
				flag2 = false // no category
			}
			// if the question has category, delete in qc table
			if flag2 {
				if err = dao.DB.Where("question_id=?", question.ID).Delete(models.QuestionCategory{}).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				}
			}

			val := function.CheckCategoryDuplicate(categories)
			if !val {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate categories!"})
				return
			}

			// update categories
			for i := range categories {
				var category models.Category
				dao.DB.Where("category_name=?", categories[i].CategoryName).First(&category)
				if category.ID == "" {
					category.ID = newuuid.New().String()
					category.CategoryName = categories[i].CategoryName
					if err := dao.DB.Create(&category).Error; err != nil {
						c.JSON(http.StatusOK, gin.H{"error": err.Error()})
					}
				}
				categories[i].ID = category.ID
			}
		}

		// send into the DB, and then response
		if err := dao.DB.Save(&question).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Updated a question"})
		}

		if !flag1 {
			for i := range categories {
				var qc models.QuestionCategory
				qc.CategoryID = categories[i].ID
				qc.QuestionID = question.ID
				if err := dao.DB.Create(&qc).Error; err != nil {
					c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				}
			}
		}
	})

	// Delete a question
	authorized.DELETE("/question/:question_id", func(c *gin.Context) {
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			return
		}	// get the user info based on email

		// get the question_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			return
		}

		// check the question_id exist or not
		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			return
		}

		// check authenticated or not
		if question.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the question does not belong to this user"})
			return
		}

		// check the question has answers or not
		var answer models.Answer
		if err = dao.DB.Where("question_id=?", question.ID).First(&answer).Error; err==nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The answer is exist, user cannot delete the question"})
			return
		}

		// check the question has categories or not
		var flag bool = true // have category
		var qc models.QuestionCategory
		if err = dao.DB.Where("question_id=?", question.ID).First(&qc).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "The question does not have category"})
			flag = false // no category
		}
		// now, the user can delete the question without any answers
		// delete in qc table
		// if the question has category, delete
		if flag {
			if err = dao.DB.Where("question_id=?", question.ID).Delete(models.QuestionCategory{}).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
		}

		// delete the file if exist
		var fileQuestionArr []models.FileQuestion
		if err = dao.DB.Where("question_id=?", questionId).Find(&fileQuestionArr).Error; err!=nil {
			c.JSON(200, gin.H{"msg": "cannot find the file for this answer"})
			return
		}
		// delete the file in mysql
		if err = dao.DB.Where("question_id=?", questionId).Delete(models.FileQuestion{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an question's file in mysql"})
		}

		// delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		for _, fileQuestion := range fileQuestionArr {
			DeleteFile(S3Bucket, fileQuestion.S3ObjectName)
		}
		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})

		// then, delete the question
		if err = dao.DB.Where("id=?", question.ID).Delete(models.Question{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted a question"})
		}
	})

	// Delete a question's answer, delete the file if exist
	authorized.DELETE("/question/:question_id/answer/:answer_id", func(c *gin.Context) {
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			return
		}	// get the user info based on email

		// get the question_id and answer_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the answer_id"})
			return
		}

		// check the question_id and answer_id exist or not
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			return
		}
		if answer.QuestionID != questionId {
			c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
			return
		}

		// check authenticated or not
		if answer.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			return
		}

		// delete the file if exist
		var fileAnswerArr []models.FileAnswer
		if err = dao.DB.Where("answer_id=?", answerId).Find(&fileAnswerArr).Error; err!=nil {
			c.JSON(200, gin.H{"msg": "cannot find the file for this answer"})
			return
		}
		// delete the file in mysql
		if err = dao.DB.Where("answer_id=?", answerId).Delete(models.FileAnswer{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an answer's file in mysql"})
		}

		// delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		for _, fileAnswer := range fileAnswerArr {
			DeleteFile(S3Bucket, fileAnswer.S3ObjectName)
		}
		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})

		// Start to delete answer
		if err = dao.DB.Where("id=?", answer.ID).Delete(models.Answer{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted a question's answer"})
		}
	})

	// Update answer
	authorized.PUT("/question/:question_id/answer/:answer_id", func(c *gin.Context) {
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			return
		}	// get the user info based on email

		// get the question_id and answer_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the answer_id"})
			return
		}

		// check the question_id and answer_id exist or not
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			return
		}
		if answer.QuestionID != questionId {
			c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
			return
		}

		// check authenticated or not
		if answer.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			return
		}

		// update answer
		c.BindJSON(&answer)
		// check content is empty or not
		if answer.AnswerText == "" {
			c.JSON(204, gin.H{"error": "no content"})
			return
		}

		answer.AnswerUpdated = time.Now()
		// send into the DB, and then response
		if err := dao.DB.Save(&answer).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg": "Updated an answer",
			})
		}

	})

	// Post answer
	authorized.POST("/question/:question_id/answer", func(c *gin.Context) {
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}	// get the user info based on email
		id, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "email address is not exist"})
			return
		}
		var answer models.Answer
		c.BindJSON(&answer)
		answer.ID = newuuid.New().String()
		answer.QuestionID = id
		answer.UserID = user.ID
		answer.AnswerCreated = time.Now()
		answer.AnswerUpdated = time.Now()

		// send into the DB, and then response
		if err := dao.DB.Create(&answer).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
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
	})

	// post a new question
	authorized.POST("/question/", func(c *gin.Context) {
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
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
			return
		}

		// update categories
		for i := range categories {
			var category models.Category
			dao.DB.Where("category_name=?", categories[i].CategoryName).First(&category)
			if category.ID == "" {
				category.ID = newuuid.New().String()
				category.CategoryName = categories[i].CategoryName
				if err := dao.DB.Create(&category).Error; err != nil {
					c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				}
			}
			categories[i].ID = category.ID
		}

		// send into the DB, and then response
		if err := dao.DB.Create(&question).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
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
			}
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

		if password != user.Password {
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

	// delete a file in a question
	authorized.DELETE("/question/:question_id/file/:file_id", func(c *gin.Context) {
		// 1. authen the log in user is the owner of question
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			return
		}	// get the user info based on email
		// get the question_id and file_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			return
		}
		fileId, valid := c.Params.Get("file_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the file_id"})
			return
		}
		// check the question_id exist or not
		var fileQuestion models.FileQuestion
		if err = dao.DB.Where("id=?", fileId).First(&fileQuestion).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The file_id is not exist"})
			return
		}
		var question models.Question
		if err = dao.DB.Where("id=?", questionId).First(&question).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The question is not exist"})
			return
		}
		// check authenticated or not
		if question.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			return
		}
		// check file id is matched with question id
		if fileQuestion.QuestionID != questionId {
			c.JSON(401, gin.H{"error": "the question is not matched with the file"})
			return
		}

		// 2. delete the file in mysql
		if err = dao.DB.Where("id=?", fileId).Delete(models.FileQuestion{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an question's file in mysql"})
		}

		// 3. delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		DeleteFile(S3Bucket, fileQuestion.S3ObjectName)
		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})
	})

	// delete a file in an answer
	authorized.DELETE("/question/:question_id/answer/:answer_id/file/:file_id", func(c *gin.Context) {
		// 1. authen the log in user is the owner of answer
		email := function.FetchUsername
		var user models.User
		if err = dao.DB.Where("email_address=?", email).First(&user).Error; err!=nil {
			c.JSON(404, gin.H{"error": "cannot find the user"})
			return
		}	// get the user info based on email
		// get the question_id and answer_id
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the question_id"})
			return
		}
		answerId, valid := c.Params.Get("answer_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the answer_id"})
			return
		}
		fileId, valid := c.Params.Get("file_id")
		if !valid {
			c.JSON(204, gin.H{"error": "cannot get the file_id"})
			return
		}
		// check the question_id and answer_id exist or not
		var fileAnswer models.FileAnswer
		if err = dao.DB.Where("id=?", fileId).First(&fileAnswer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The file_id is not exist"})
			return
		}
		var answer models.Answer
		if err = dao.DB.Where("id=?", answerId).First(&answer).Error; err!=nil {
			c.JSON(404, gin.H{"error": "The answer_id is not exist"})
			return
		}
		if answer.QuestionID != questionId {
			c.JSON(404, gin.H{"error": "The question_id and answer_id are noe matched"})
			return
		}
		// check authenticated or not
		if answer.UserID !=  user.ID{
			c.JSON(401, gin.H{"error": "the answer does not belong to this user"})
			return
		}
		// check answer id is matched with question id
		if answer.QuestionID != questionId {
			c.JSON(401, gin.H{"error": "the answer is not matched with the question"})
			return
		}
		// chekc file id is mathced with the answer id
		if fileAnswer.AnswerID != answerId {
			c.JSON(401, gin.H{"error": "the file is not belonging to this answer"})
			return
		}

		// 2. delete the file in mysql
		if err = dao.DB.Where("id=?", fileId).Delete(models.FileAnswer{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted an answer's file in mysql"})
		}

		// 3. delete the file in AWS S3
		//S3Bucket := GetEnvWithKey("BUCKET_NAME")
		S3Bucket := "webapp.chaoyi.yuan"
		DeleteFile(S3Bucket, fileAnswer.S3ObjectName)
		c.JSON(200, gin.H{"msg": "Deleted a file in AWS S3"})
	})

	// post a file to an answer
	authorized.POST("/question/:question_id/answer/:answer_id/file", func(c *gin.Context) {
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
			return
		}	// get the user info based on email
		questionId, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "question is not exist"})
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
			return
		}
		//
		if answer.UserID != user.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this question id is not belong to the user"})
			return
		}
		//
		if answer.QuestionID != questionId {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this answer id is not matched with the question if"})
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

		// send into the DB, and then response
		if err := dao.DB.Create(&fileAnswer).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"file_name": fileAnswer.FileName,
				"s3_object_name": fileAnswer.S3ObjectName,
				"file_id": fileAnswer.ID,
				"created_date": fileAnswer.CreateDate,
			})
		}

	})

	// post a file to a question
	authorized.POST("/question/:question_id/file", func(c *gin.Context) {
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
			return
		}	// get the user info based on email
		id, valid := c.Params.Get("question_id")
		if !valid {
			c.JSON(http.StatusOK, gin.H{"error": "question is not exist"})
			return
		}
		var question models.Question
		if err = dao.DB.Where("id=?", id).First(&question).Error; err!=nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}
		//
		if question.UserID != user.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this question id is not belong to the user"})
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

		// send into the DB, and then response
		if err := dao.DB.Create(&fileQuestion).Error;err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"file_name": fileQuestion.FileName,
				"s3_object_name": fileQuestion.S3ObjectName,
				"file_id": fileQuestion.ID,
				"created_date": fileQuestion.CreateDate,
			})
		}
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
		}else{
			sess = newSess
		}
	}

	return sess
}