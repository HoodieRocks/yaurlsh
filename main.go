package main

import (
	"context"
	"log"
	"net/http"
	"regexp"

	"github.com/hoodierocks/yaurlsh/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var urlRegexp = `^(?!mailto:)(?:(?:http|https|ftp)://)(?:\\S+(?::\\S*)?@)?(?:(?:(?:[1-9]\\d?|1\\d\\d|2[01]\\d|22[0-3])(?:\\.(?:1?\\d{1,2}|2[0-4]\\d|25[0-5])){2}(?:\\.(?:[0-9]\\d?|1\\d\\d|2[0-4]\\d|25[0-4]))|(?:(?:[a-z\\u00a1-\\uffff0-9]+-?)*[a-z\\u00a1-\\uffff0-9]+)(?:\\.(?:[a-z\\u00a1-\\uffff0-9]+-?)*[a-z\\u00a1-\\uffff0-9]+)*(?:\\.(?:[a-z\\u00a1-\\uffff]{2,})))|localhost)(?::\\d{2,5})?(?:(/|\\?|#)[^\\s]*)?$`

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
	
	RegisterRoutes(e)
	
	e.Logger.Fatal(e.Start("localhost:1323"))
}

func RegisterRoutes(e *echo.Echo) {
	
	e.GET("/:alias", func(c echo.Context) error {
		alias := c.Param("alias")
		conn, err := db.Connect()
		
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error connecting to database")
		}
		defer conn.Close()
		
		url, err := conn.GetURL(context.Background(), alias)
		
		if err != nil {
			return c.String(http.StatusNotFound, "URL not found")
		}
		
		return c.Redirect(http.StatusMovedPermanently, url)
	})

	e.Static("/", "./client/")
	e.File("/index.css", "./client/index.css")
	
	e.POST("/api/shorten", func(c echo.Context) error {
		alias := c.FormValue("alias")
		url := c.FormValue("url")

		conn, err := db.Connect()

		if err != nil {
			return c.String(http.StatusInternalServerError, "Error connecting to database")
		}
		defer conn.Close()

		if url == "" {
			return c.String(http.StatusBadRequest, "URL is required")
		}

		if len(url) > 2048 {
			return c.String(http.StatusBadRequest, "URL is too long")
		}

		if valid, err := regexp.MatchString(urlRegexp, url); !valid || err != nil {
			return c.String(http.StatusBadRequest, "URL is invalid")
		}

		if alias == "" {
			alias, err = gonanoid.New(8)

			if err != nil {
				return c.String(http.StatusInternalServerError, "Error generating alias")
			}
		}

		if len(alias) > 32 {
			return c.String(http.StatusBadRequest, "Alias is too long")
		}

		err = conn.CreateURL(context.Background(), alias, url)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error creating URL")
		}
		return c.String(http.StatusOK, "URL created")
	})

	// add route for homepage
	e.GET("/api/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to the YAURLSH API!")
	})

}