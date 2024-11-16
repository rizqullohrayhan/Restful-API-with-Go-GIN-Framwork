package main

import (
	"net/http"
    "database/sql"
    "fmt"
	"log"

    "github.com/go-sql-driver/mysql"
	"github.com/gin-gonic/gin"
)

var db *sql.DB

// album represents data about a record album
type album struct {
	ID		string	`form:"id,omitempty"`
	Title	string	`form:"title" binding:"required"`
	Artist	string	`form:"artist" binding:"required"`
	Price	float32	`form:"price"`
}

func init()  {
	// Capture connection properties.
	cfg := mysql.Config{
		User:					"root",
		Passwd:					"",
		Net:					"tcp",
		Addr: 					"127.0.0.1:3306",
		DBName: 				"recordings",
		AllowNativePasswords: 	true,
	}
	// Get a dataabse handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	if pingErr := db.Ping(); pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Database Connected!")
}

func main()  {
	router := gin.Default()

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", addAlbum)
	router.PUT("/albums/:id", editAlbum)
	router.DELETE("/albums/:id", destroyAlbum)

	router.Run("localhost:8080")
}

// check database connection
// return true if connected
// return false if not connected
func checkDBConnection(c *gin.Context) bool {
	if err := db.Ping(); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return false
	}
	return true
}

// getAlbums get all album data from database
func getAlbums(c *gin.Context)  {
	if !checkDBConnection(c) {
		return
	}

	var albums []album

	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var album album
		if err := rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
		}
		albums = append(albums, album)
	}

	c.JSON(http.StatusOK, albums)
}

// getAlbumByID get specified album data by ID
func getAlbumByID(c *gin.Context)  {
	if !checkDBConnection(c) {
		return
	}

	var album album
	id := c.Param("id")
	result := db.QueryRow("SELECT * FROM album WHERE id=?", id)
	if err := result.Scan(&album.ID, &album.Title, &album.Artist, &album.Price); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, album)
}

// make validation for album manajemens
// return true if pass
// return false if invalid
func dataAlbumValidation(album album, c *gin.Context) bool {
	if album.Artist == "" || album.Title == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Title and Artist are required"})
        return false
	}
	if album.Price < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "price must be a positive value"})
        return false
	}
	return true
}

// addAlbum to insert data into album table
func addAlbum(c *gin.Context)  {
	if !checkDBConnection(c) {
		return
	}

	var postAlbum album

	if err := c.ShouldBind(&postAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Data validation
	if !dataAlbumValidation(postAlbum, c) {
		return
	}

	result, err := db.Exec(
		"INSERT INTO album (title, artist, price) VALUES (?, ?, ?)",
		postAlbum.Title, postAlbum.Artist, postAlbum.Price,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Album berhasil ditambahkan", "id": lastInsertID})
}

// modify data album
func editAlbum(c *gin.Context)  {
	if !checkDBConnection(c) {
		return
	}

	var album album
	id := c.Param("id")
	if err := c.ShouldBind(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if !dataAlbumValidation(album, c) {
		return
	}

	result, err := db.Exec(
		"UPDATE album SET title=?, artist=?, price=? WHERE id=?",
		album.Title, album.Artist, album.Price, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffacted, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Album succesfully edited!", "rows affected": rowsAffacted})
}

// delete data album
func destroyAlbum(c *gin.Context) {
	if !checkDBConnection(c) {
		return
	}

	id := c.Param("id")
	result, err := db.Exec("DELETE FROM album WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Album successfully deleted!", "rowsAffected": rowsAffected})
}