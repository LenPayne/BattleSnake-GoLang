package main

import (
	"github.com/LenPayne/BattleSnake-GoLang/pkg/snake"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"apiversion": "1",
			"author":     "lenpayne",
			"color":      getEnv("SNAKE_COLOUR", "#F64A91"),
			"head":       getEnv("SNAKE_HEAD", "missile"),
			"tail":       getEnv("SNAKE_TAIL", "missile"),
			"version":    "0.0.1-alpha",
		})
	})
	r.POST("/move", func(c *gin.Context) {
		var json snake.Payload
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"move":  snake.Move(json),
			"shout": "From hell’s heart I stab at thee",
		})
	})
	r.POST("/start", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})
	r.POST("/end", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})
	r.Run()
}

func getEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return fallback
}
