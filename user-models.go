package main

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


type User struct{
	Email string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}

func createUser(db *gorm.DB, user *User) error{
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)

	if err != nil {
		return err
	}

	user.Password = string(hashPassword)

	result := db.Create(user)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func loginUser(db *gorm.DB, user *User)(string, error){
	// get user from email 
	selectedUser := new(User)

	result := db.Where("email = ?", user.Email).First(selectedUser)

	if result.Error != nil {
		return "",result.Error
	}

	// compare password
	err := bcrypt.CompareHashAndPassword([]byte(selectedUser.Password), []byte(user.Password))

	if err != nil{
		return "", err
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"email":  user.Email,
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil{
		return "", err
	}

	return t, nil
}