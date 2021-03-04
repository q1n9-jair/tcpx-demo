package dao

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
	"log"
)

var DbEngine *gorm.DB

func init() {
	dsn := "root:root@tcp(127.0.0.1:3306)/dbo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(err.Error())
	}
	DbEngine = db
	fmt.Println(DbEngine)
}