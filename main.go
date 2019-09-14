package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

//DB pointer
var db *sql.DB

//Customer struct for keeping information
type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")

	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Error!!! Unauthorization"})
		c.Abort()
		return
	}

	c.Next()

}

func postCustomer(c *gin.Context) {
	cust := Customer{}

	err := c.ShouldBindJSON(&cust)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "JSON parsing on insertion error!!! " + err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1,$2,$3) RETURNING id", cust.Name, cust.Email, cust.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Insertion error!!! " + err.Error()})
		return
	}
	cust.ID = id

	c.JSON(http.StatusCreated, cust)
}

func getAllCustomer(c *gin.Context) {

	custs := []Customer{}
	cust := Customer{}

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Prepare SQL for selection error!!! " + err.Error()})
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Query to select row error!!! " + err.Error()})
		return
	}

	for rows.Next() {
		err := rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Fetch next row error!!! " + err.Error()})
			return
		}
		custs = append(custs, cust)
	}

	c.JSON(http.StatusOK, custs)
}

func getOneCustomer(c *gin.Context) {
	id := c.Param("id")

	var cust Customer

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers WHERE id = $1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Prepare SQL select error!!! " + err.Error()})
		return
	}

	idnum, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Convert id to num error!!!" + err.Error()})
		return
	}

	row := stmt.QueryRow(idnum)

	err = row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
	if err != nil {
		log.Println("Select id = " + id)
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Select id %d error!!!: %s", idnum, err.Error())})
		return
	}

	c.JSON(http.StatusOK, cust)
}

func deleteCustomer(c *gin.Context) {

	id := c.Param("id")
	//Pre-select.
	var cust Customer

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers WHERE id = $1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Prepare SQL select error!!! " + err.Error()})
		return
	}

	idnum, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Convert id to num error!!!" + err.Error()})
		return
	}

	row := stmt.QueryRow(idnum)

	err = row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
	if err != nil {
		log.Println("Select id = " + id)
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Select id %d error!!!: %s", idnum, err.Error())})
		return
	}

	//Actual delete.
	stmt, err = db.Prepare("DELETE FROM customers WHERE id = $1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Prepare SQL for delete row error!!! " + err.Error()})
		return
	}

	_, err = stmt.Exec(idnum)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Execute deletion error!!! " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func updateCustomer(c *gin.Context) {
	id := c.Param("id")

	//Pre-select.
	var cust Customer

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers WHERE id = $1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Prepare SQL select error!!! " + err.Error()})
		return
	}

	idnum, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Convert id to num error!!! " + err.Error()})
		return
	}

	row := stmt.QueryRow(idnum)

	err = row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
	if err != nil {
		log.Println("Select id = " + id)
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Select id %d error!!!: %s", idnum, err.Error())})
		return
	}

	//Actual update.
	custUpdate := Customer{}

	err = c.ShouldBindJSON(&custUpdate)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "JSON parsing Error!!! " + err.Error()})
		return
	}

	stmt, err = db.Prepare("UPDATE customers SET name=$2,email=$3,status=$4 WHERE id=$1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Prepare SQL for update error!!! " + err.Error()})
		return
	}

	_, err = stmt.Exec(id, custUpdate.Name, custUpdate.Email, custUpdate.Status)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Execute update error!!! " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, custUpdate)
}

func main() {
	createTable()

	r := gin.Default()

	r.Use(authMiddleware)

	r.GET("/customers", getAllCustomer)

	r.GET("customers/:id", getOneCustomer)

	r.POST("/customers", postCustomer)

	r.DELETE("/customers/:id", deleteCustomer)

	r.PUT("/customers/:id", updateCustomer)

	r.Run(":2019")

	defer db.Close()
}

func createTable() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect to database error", err)
		return
	}

	createTb := `
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`

	_, err = db.Exec(createTb)
	if err != nil {
		log.Println("Cannot create table.", err)
		return
	}

	fmt.Println("Successfully create table.")
}
