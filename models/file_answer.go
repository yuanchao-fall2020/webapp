package models

import "time"

type FileAnswer struct {
	ID string `json:"file_id"`
	// this file name is the original name for the file
	FileName string `json:"file_name"`
	// this name is defined by myself, the purpose is to set it unique in AWS S3
	S3ObjectName string `json:"s3_object_name"`
	CreateDate time.Time `json:"create_date"`
	Answer Answer `gorm:"ForeignKey:ID;AssociationForeignKey:AnswerID" json:"-"`
	AnswerID string `json:"-"`
	AcceptRanges string `json:"-"`
	ContentLength string `json:"-"`
	ContentType string `json:"-"`
	ETag string `json:"-"`
}