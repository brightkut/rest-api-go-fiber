package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/brightkut/rest-api-go-fiber/docs"
	jwtware "github.com/gofiber/contrib/jwt"
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

type User struct{
	Email string `json:"email"`
	Password string `json:"password"`
}

var user1 = User{
	Email: "test@gmail.com",
	Password: "1234",
}

func logMiddleware(c *fiber.Ctx) error {
	startTime := time.Now()

	fmt.Printf("URL = %s, Method = %s, Time = %s \n", c.OriginalURL(), c.Method(), startTime)

	return c.Next()
}

func loginMiddleware(c *fiber.Ctx) error{
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	if email != "test@gmail.com"{
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

const (
	host = "localhost"
	port = 5432
	user = "postgres"
	password = "admin"
	dbname = "ticket"
)

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

	  
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user , password , dbname)

	db, err:= gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil{
		panic("failed to connect DB")
	}
	// auto create and update table but not for delete case
	db.AutoMigrate(&Ticket{})
	fmt.Printf("Connect DB successfully")

	// create ticket
	ticket := &Ticket{Name: "Ticket1", Price: 100}
	
	createTicket(db, ticket)

	ticket2 := getTicket(db, 1)
	ticket2.Name = "Agoda Ticket"
	ticket2.Price = 200

	updateTicket(db, ticket2)

	getTicket(db, 1)

	deleteTicket(db, 1)
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

	// setup jwt middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
	}))

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

	fmt.Print(user)

	if user.Email != user1.Email || user.Password != user1.Password{
		return fiber.ErrUnauthorized
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
	
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "Login success",
		"token": t,
	})
}