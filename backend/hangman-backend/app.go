package main

import (
	"hangman"

	"github.com/gin-gonic/gin"
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
