package models

import "time"

/*type Answer struct {
	AnswerID string `json:"answer_id"`
	AnswerCreated time.Time `json:"created_timestamp"`
	AnswerUpdated time.Time `json:"updated_timestamp"`
	User User `json:"user_id" gorm:"foreignKey:UserRefer"`
	Question Question `json:"question_id" gorm:"foreignKey:QuestionRefer"`
	AnswerText string `json:"answer_text"`
}*/

type Answer struct {
	ID string `json:"answer_id"`
	AnswerCreated time.Time `json:"created_timestamp"`
	AnswerUpdated time.Time `json:"updated_timestamp"`
	User User `gorm:"ForeignKey:ID;AssociationForeignKey:UserID" json:"-"`
	UserID string `json:"user_id"`
	Question Question `gorm:"ForeignKey:ID;AssociationForeignKey:QuestionID" json:"-"`
	QuestionID string `json:"question_id"`
	AnswerText string `json:"answer_text"`
	FileArr []FileAnswer `json:"attachments" sql:"-"`
}