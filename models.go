package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Ticket struct{
	// create default field created_at, updated_at, deleted_at
	// gorm.Model
	TicketId int64 `gorm:"primaryKey"`
	Name string `gorm:"index"`
	Price int64
	CreatedAt time.Time      // Manually define timestamp fields
	UpdatedAt time.Time      
	DeletedAt gorm.DeletedAt  // Enables soft delete
}

func createTicket(db *gorm.DB, ticket *Ticket){
	result := db.Create(ticket)

	if result.Error != nil {
		log.Fatal("Error occur when create ticket")
	}

	fmt.Printf("Create ticket success")
}

func getTicket(db *gorm.DB, ticketId int64) *Ticket{
	var ticket Ticket

	result := db.First(&Ticket{}, ticketId)

	if result.Error != nil {
		log.Fatalf("Error get ticket %v", result.Error)
	}

	return &ticket
}

func updateTicket(db *gorm.DB, ticket *Ticket){
	
	result := db.Save(ticket)

	if result.Error != nil {
		log.Fatalf("Error update ticket %v", result.Error)
	}

	fmt.Printf("Update ticket success")
}

func deleteTicket(db *gorm.DB, ticketId int64){

	// Soft delete because model has field `deleted_at`
	// if don't want soft delete remove this field
	result := db.Delete(&Ticket{}, ticketId)

	if result.Error != nil {
		log.Fatalf("Error delete ticket %v", result.Error)
	}

	fmt.Printf("Delete ticket success")
}
