package models

import "time"

type User struct {
	//gorm.Model
	ID string `json:"id"`
	AccountCreated time.Time `json:"account_created"`
	AccountUpdated time.Time `json:"account_updated"`
	EmailAddress *string `json:"email_address" gorm:"unique;not null;"`
	Password string `json:"password" gorm:"<-"`
	FirstName string `json:"first_name" gorm:"<-"`
	LastName string `json:"last_name" gorm:"<-"`

}