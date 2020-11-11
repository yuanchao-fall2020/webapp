package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"os"
	"strings"
)

var DB *gorm.DB

type DBConfig struct {
	Host     string
	Port     int
	User     string
	DBName   string
	Password string
}

func BuildDBConfig() *DBConfig {
	dbConfig := DBConfig{
		Host:     GetHostname(),//"localhost",
		Port:     3306,
		User:     "csye6225fall2020", //"root",
		Password: "Y940519a",
		DBName:   "csye6225",//"db1",
	}
	return &dbConfig
}
func DbURL(dbConfig *DBConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,

	)
}

func GetHostname() string {
	hostname := os.Getenv("HOSTNAME")
	return strings.TrimRight(hostname, ":3306")
}