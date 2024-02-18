package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"encoding/json"
	"io/ioutil"
)
const (
	ImgDir = "images"
)
type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello,Nanachand!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// Item 構造体の定義
type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// Items 構造体の定義
type Items struct {
	Items []Item `json:"items"`
}

func getItems(c echo.Context) error {
        // item.jsonを読み込む
	data, err := ioutil.ReadFile("./item.json")
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}

	var items Items
        // 読み込んだdataをItemsという型に変換してitemsに格納
	err = json.Unmarshal(data, &items)
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}

	return c.JSON(http.StatusOK, items)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
