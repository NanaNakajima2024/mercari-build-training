package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

// 読み込んだ画像ファイルを受け取り、ハッシュ化して返す
func sha_256(target []byte) string {
	h := sha256.New()
	h.Write(target)
	hashBytes := h.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func addItem(c echo.Context) error {
	db, err := sql.Open("sqlite3", "/Users/nakajimanana/mercari-build-training/go/mercari.sqlite3")
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}

	defer db.Close()
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image := c.FormValue("image")
	id := c.FormValue("id")
	c.Logger().Infof("Receive item: %s", name)

	target, _ := ioutil.ReadFile(image)
	hash := sha_256(target)
	hashImage := hash + ".jpg"

	insertQuery := "INSERT INTO items(name, category,image_name,id) VALUES(?, ?, ?, ?)"
	_, err = db.Exec(insertQuery, name, category, hashImage, id)
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}

	return c.JSON(http.StatusOK, "success")
}

// Item 構造体の定義
type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
	Id       string `json:"id"`
}

// Items 構造体の定義
type Items struct {
	Items []Item `json:"items"`
}

func getItems(c echo.Context) error {
	db, err := sql.Open("sqlite3", "/Users/nakajimanana/mercari-build-training/go/mercari.sqlite3")
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}

	defer db.Close()

	rows, err := db.Query("SELECT * FROM items")
	if err != nil {
		c.Logger().Debugf("Image not found: %s", "aaa")
		return err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.Image)
		if err != nil {
			res := Response{Message: err.Error()}
			return c.JSON(http.StatusBadRequest, res)
		}
		items = append(items, item)
	}

	return c.JSON(http.StatusOK, Items{Items: items})
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

func getItemByID(c echo.Context) error {
	id := c.Param("id")

	data, err := ioutil.ReadFile("./items.json")
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
	// 商品IDに一致する商品を検索
	var item Item
	for _, i := range items.Items {
		if i.Id == id {
			item = i
			break
		}
	}

	// 商品が見つからなかった場合は404エラーを返す
	if item.Id == "" {
		return c.String(http.StatusNotFound, "Item not found")
	}

	return c.JSON(http.StatusOK, item)
}

func searchItems(c echo.Context) error {

	keyword := c.QueryParam("keyword")

	// データベースに接続
	db, err := sql.Open("sqlite3", "/Users/nakajimanana/mercari-build-training/go/mercari.sqlite3")
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}
	defer db.Close()

	//検索クエリを作成
	query := "SELECT id, name, category, image_name FROM items WHERE name LIKE '%' || ? || '%'"

	rows, err := db.Query(query, keyword)
	if err != nil {
		res := Response{Message: err.Error()}
		return c.JSON(http.StatusBadRequest, res)
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.Image)
		if err != nil {
			res := Response{Message: err.Error()}
			return c.JSON(http.StatusBadRequest, res)
		}
		items = append(items, item)
	}

	return c.JSON(http.StatusOK, Items{Items: items})
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
	e.GET("/search", searchItems)
	e.GET("/items/:id", getItemByID)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
