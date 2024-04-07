package userAccounts

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string
	Salt         []byte
	PasswordHash string
}

type GameInstance struct {
	gorm.Model
	Players []*User
	Guesses map[*User]string
}

type Game struct {
	gorm.Model
	Instances []*GameInstance
}
