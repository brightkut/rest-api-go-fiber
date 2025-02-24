package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/brightkut/rest-api-go-fiber/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Book data model
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// Book DB
var books []Book

func logMiddleware(c *fiber.Ctx) error {
	startTime := time.Now()

	fmt.Printf("URL = %s, Method = %s, Time = %s \n", c.OriginalURL(), c.Method(), startTime)

	return c.Next()
}

func  loginMiddleware(c *fiber.Ctx) error{
	cookie := c.Cookies("jwt")
	jwtSecretKey := os.Getenv("JWT_SECRET")

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecretKey), nil
	})

	if err != nil || !token.Valid {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claim := token.Claims.(jwt.MapClaims)

	fmt.Print(claim)

	return c.Next()
}

const (
	host = "localhost"
	port = 5432
	user = "postgres"
	password = "admin"
	dbname = "ticket"
)

var db *gorm.DB 

// @title Book API
// @description This is sample book API.
// @version 1.0
// @host localhost:8080
// @BasePath / 
// @schemes httpr
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
		  SlowThreshold:              time.Second,   // Slow SQL threshold
		  LogLevel:                   logger.Info, // Log level	
		  Colorful:                  true,
		},
	  )
	var err error
	  
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user , password , dbname)

	db, err= gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil{
		panic("failed to connect DB")
	}
	// auto create and update table but not for delete case
	db.AutoMigrate(&Ticket{}, &User{})
	fmt.Printf("Connect DB successfully")

	// create ticket
	// ticket := &Ticket{Name: "Ticket1", Price: 100}
	// createTicket(db, ticket)
	// ticket2 := getTicket(db, 1)
	// ticket2.Name = "Agoda Ticket"
	// ticket2.Price = 200
	// updateTicket(db, ticket2)
	// getTicket(db, 1)
	// deleteTicket(db, 1)
	// create app
	app := fiber.New()

	// setup swaggeer
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Initialize books DB
	books = append(books, Book{ID: 1, Title: "Book 1", Author: "Author 1"})
	books = append(books, Book{ID: 2, Title: "Book 2", Author: "Author 2"})

	// load env
	if err := godotenv.Load(); err != nil{
		log.Fatal("Load env error")
	}

	// setup middleware
	app.Use(logMiddleware)

	app.Get("/health", healthCheck)
	app.Post("/login", login)
	app.Post("/register", register)

	// setup jwt middleware
	// app.Use(jwtware.New(jwtware.Config{
	// 	SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
	// }))

	// check token middleware
	app.Use(loginMiddleware)

	// create route
	app.Get("/books", getBooks)
	app.Get("/books/:id", getBook)
	app.Post("/books", createBook)
	app.Put("/books/:id", updateBook)
	app.Delete("/books/:id", deleteBook)
	app.Post("/upload", uploadFile)
	app.Get("/config", getEnv)

	// listen port
	app.Listen(":8080")
}

func uploadFile(c *fiber.Ctx) error{
	file, err := c.FormFile("image")

	if err != nil{
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err = c.SaveFile(file, "./uploads/" + file.Filename)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("File upload completed !!")
}

func getEnv(c *fiber.Ctx) error{
	return c.JSON(fiber.Map{
		"SECRET": os.Getenv("SECRET"),
	})
}

func login(c *fiber.Ctx) error{
	user := new(User)

	if err := c.BodyParser(user); err != nil{
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	token , err := loginUser(db, user)
	
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name: "jwt",
		Value: token,
		Expires: time.Now().Add(time.Hour * 72),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"message": "Login success",
	})
}

func register(c *fiber.Ctx)error{
	user := new(User)

	if err := c.BodyParser(user); err != nil{
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}	

	err:= createUser(db, user)

	if err != nil{
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"message": "Register success",
	})
}