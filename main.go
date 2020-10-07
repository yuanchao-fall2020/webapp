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
	dao.DB.AutoMigrate(&models.Question{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	dao.DB.AutoMigrate(&models.Answer{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT").AddForeignKey("question_id", "questions(id)", "RESTRICT", "RESTRICT")
	dao.DB.AutoMigrate(&models.Category{})
	dao.DB.AutoMigrate(&models.QuestionCategory{}).AddForeignKey("question_id", "questions(id)", "RESTRICT", "RESTRICT").AddForeignKey("category_id", "categories(id)", "RESTRICT", "RESTRICT")

	// create a default router
	r := gin.Default()

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
			id := c.Params.ByName("email_address")
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
			// then, print out
			c.JSON(http.StatusOK, gin.H{
				"question_id": question.ID,
				"created_timestamp": question.QuestionCreated,
				"updated_timestamp": question.QuestionUpdated,
				"user_id": question.UserID,
				"question_text": question.QuestionText,
				"categories": cateArr,
				"answers": answerArr,
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
				questionArr[i].CategoryArr = cateArr
				questionArr[i].AnswerArr = answerArr
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

			// now, we can get the answer
			c.JSON(http.StatusOK, gin.H{
				"answer_id": answer.ID,
				"question_id": answer.QuestionID,
				"created_timestamp": answer.AnswerCreated,
				"updated_timestamp": answer.AnswerUpdated,
				"user_id": answer.UserID,
				"answer_text": answer.AnswerText,
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
		// then, delete the question
		if err = dao.DB.Where("id=?", question.ID).Delete(models.Question{}).Error; err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "Deleted a question"})
		}
	})

	// Delete a question's answer
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

	r.Run(":9090")
}





