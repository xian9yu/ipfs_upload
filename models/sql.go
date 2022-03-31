package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
)

var (
	DB  *gorm.DB
	err error
)

// InitSQL 初始化sql
func InitSQL() *gorm.DB {

	DB, err = gorm.Open(mysql.Open("ipfs_upload:zheshimima@tcp(192.168.1.245:3306)/ipfs_upload?charset=utf8"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//TablePrefix:   "db_", // 表名前缀，`User` 的表名应该是 `db_users`
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `we_user`
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel: logger.Info, // Log level
				Colorful: true,        // 彩色打印
			},
		),
	})
	if err != nil {
		log.Fatalln("MySQL启动异常")
	}

	if err = DB.AutoMigrate(&File{}); err != nil {
		log.Println("同步数据库表失败:", err.Error())

	}
	return DB
}
