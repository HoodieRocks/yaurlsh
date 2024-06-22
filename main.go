package main

import (
	"context"
	"log"
	"net/http"

	"github.com/hoodierocks/yaurlsh/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to database
	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// test database connection
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create tables
	err = conn.CreateTables(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// start http server
	e := echo.New()

	e.GET("/:alias", func(c echo.Context) error {
		alias := c.Param("alias")
		url, err := conn.GetURL(context.Background(), alias)
		if err != nil {
			return c.String(http.StatusNotFound, "URL not found")
		}
		return c.Redirect(http.StatusMovedPermanently, url)
	})

	e.POST("/p/create", func(c echo.Context) error {
		alias := c.FormValue("alias")
		url := c.FormValue("url")
		err := conn.CreateURL(context.Background(), alias, url)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error creating URL")
		}
		return c.String(http.StatusOK, "URL created")
	})

	// add route for homepage
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to YAURLSH!")
	})

	e.Logger.Fatal(e.Start("localhost:1323"))
}