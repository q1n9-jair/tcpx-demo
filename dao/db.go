package dao

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DbEngine *gorm.DB

func initd() {
	dsn := "root:root@tcp(127.0.0.1:3306)/dbo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(err.Error())
	}
	DbEngine = db
	fmt.Println(DbEngine)
}
