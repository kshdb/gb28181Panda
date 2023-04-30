package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GinSRes(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  msg,
		"data": data,
	})
}
func GinFRes(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  msg,
		"data": data,
	})
}
