package models

type QuestionCategory struct {
	/*QID string    `gorm:"primary_key" json:"question_id"`
	CID string    `gorm:"primary_key" json:"category_id"`*/
	QuestionID string `json:"question_id" gorm:"primary_key"`
	Question   Question `gorm:"ForeignKey:ID;AssociationForeignKey:QuestionID"`
	CategoryID string `json:"category_id" gorm:"primary_key"`
	Category   Question `gorm:"ForeignKey:ID;AssociationForeignKey:CategoryID"`
}