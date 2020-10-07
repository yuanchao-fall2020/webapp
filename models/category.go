package models

/*type Category struct {
	CategoryID string `json:"category_id"`
	CategoryName string `json:"category"`
	Question Question `json:"question_id" gorm:"foreignKey:QuestionRefer"`
}*/

type Category struct {
	ID string `json:"category_id"`
	CategoryName string `json:"category"` /*gorm:"unique;not null;"*/
	/*Question Question `gorm:"ForeignKey:Id;AssociationForeignKey:QuestionId"`
	QuestionID string `json:"question_id"`*/
}