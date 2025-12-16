package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Hello, World!")

	api := fiber.New()
	// routes

	api.Listen("localhost:9000")
}