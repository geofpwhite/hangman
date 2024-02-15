package main

import (
	"github.com/gin-gonic/gin"
	"hangman"
)

func main() {
	go func() {
		r := gin.Default()
		gin.SetMode(gin.ReleaseMode)
		r.Static("/", "./build/")
		r.Run("localhost:4200")
	}()
	// hangman.TestRun()
	hangman.Run()
}
