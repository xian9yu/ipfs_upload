package main

import (
	"github.com/gin-gonic/gin"
	"ipfs_upload/ctrls"
	"ipfs_upload/models"
	"log"
)

func InitRouter(r *gin.Engine) {
	r.POST("/add", ctrls.Add)
	r.GET("/")
}
func main() {
	// 初始化router
	r := gin.Default()
	InitRouter(r)
	// 初始化sql
	models.InitSQL()

	// 使用gin自带的异常恢复中间件，避免出现异常时程序退出
	r.Use(gin.Recovery())

	err := r.Run(":8000")
	if err != nil {
		log.Fatalln("服务启动失败 ：", err)
	}
}
