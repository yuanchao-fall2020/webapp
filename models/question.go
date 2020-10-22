package models

import "time"

type Question struct {
	ID string `json:"question_id"`
	QuestionCreated time.Time `json:"created_timestamp"`
	QuestionUpdated time.Time `json:"updated_timestamp"`
	User User `gorm:"ForeignKey:Id;AssociationForeignKey:UserId" json:"-"`
	UserID string `json:"user_id"`
	QuestionText string `json:"question_text"`
	CategoryArr []Category `json:"categories" sql:"-"`
	AnswerArr []Answer `json:"answers" sql:"-"`
	FileArr []FileQuestion `json:"attachments" sql:"-"`
}

/*type Question struct {
	QuestionID string `json:"question_id"`
	QuestionCreated time.Time `json:"created_timestamp"`
	QuestionUpdated time.Time `json:"updated_timestamp"`
	UserID string `json:"user_id" gorm:"foreignKey:ID"`
	QuestionText string `json:"question_text"`
}*/