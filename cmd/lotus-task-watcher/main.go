package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/notify_task", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
			})
			return
		}
		data, _ := json.MarshalIndent(req, "", " ")
		fmt.Println(string(data))
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
