package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type oneRec struct {
	Board       string  `json:"board"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}

type oneBoard struct {
	Board       string `json:"board"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type allRecords []oneRec

var database *sql.DB

func main() {

	database, _ = sql.Open("sqlite3", "./fathenda.db")

	prepareDatabase()

	http.HandleFunc("/get_board_data", jsonResponse)
	http.HandleFunc("/get_board_data_count", jsonResponseCount)
	http.HandleFunc("/", htmlHelpResponse)
	http.HandleFunc("/set_board_data", setSensorData)

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatal(err)
	}

}

func checkIfBoardExist(BoardObj oneBoard) {
	// database, _ :=
	// 	sql.Open("sqlite3", "./fathenda.db")

	rows, _ := database.Query("SELECT id, board FROM sensors WHERE board='" + BoardObj.Board + "'")
	defer rows.Close()

	var count int = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		statement, _ :=
			database.Prepare("INSERT INTO sensors (board, name, description, added) VALUES (?, ?, ?, ?)")
		statement.Exec(BoardObj.Board, "", "", time.Now().Unix())
	}
}

func prepareDatabase() {

	// database, _ :=
	// 	sql.Open("sqlite3", "./fathenda.db")

	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS sensorsdata (id INTEGER PRIMARY KEY, board TEXT, timestamp NUMERIC, temperature NUMERIC, humidity NUMERIC, pressure NUMERIC)")
	statement.Exec()

	statement1, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS sensors (id INTEGER PRIMARY KEY, board TEXT, name TEXT, description TEXT, added NUMERIC)")
	statement1.Exec()

}

func setSensorData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		var rec oneRec

		err := json.NewDecoder(r.Body).Decode(&rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(rec.Board)

		var Board oneBoard

		Board.Board = rec.Board
		Board.Description = ""
		Board.Name = ""

		checkIfBoardExist(Board)

		database, _ :=
			sql.Open("sqlite3", "./fathenda.db")
		statement, _ :=
			database.Prepare("CREATE TABLE IF NOT EXISTS sensorsdata (id INTEGER PRIMARY KEY, board TEXT, timestamp NUMERIC, temperature NUMERIC, humidity NUMERIC, pressure NUMERIC)")
		statement.Exec()
		statement, _ =
			database.Prepare("INSERT INTO sensorsdata (board, timestamp, temperature, humidity, pressure) VALUES (?, ?, ?, ?, ?)")
		statement.Exec(rec.Board, time.Now().Unix(), rec.Temperature, rec.Humidity, rec.Pressure)

		w.WriteHeader(http.StatusCreated)
		addedrecord, _ := getJSON("SELECT * FROM sensorsdata ORDER BY id DESC LIMIT 1")
		fmt.Fprintf(w, addedrecord)

		fmt.Println(rec)

	default:
		fmt.Fprintf(w, "Sorry, only POST methods are supported.")
	}
}

func getJSON(sqlString string) (string, error) {

	// database, _ :=
	// 	sql.Open("sqlite3", "./fathenda.db")

	rows, err := database.Query(sqlString)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	// database.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	var rowCount int = 0
	for rows.Next() {

		rowCount++

		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}

	fmt.Println("Row count : " + strconv.Itoa(rowCount))

	jsonData, err := json.Marshal(tableData)
	if err != nil {
		return "", err
	}

	fmt.Println(string(jsonData))
	return string(jsonData), nil
}

func getCount(sqlString string) (string, error) {

	// database, _ :=
	// 	sql.Open("sqlite3", "./fathenda.db")

	rows, err := database.Query(sqlString)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var count int = 0
	for rows.Next() {
		count++
	}
	return string(count), nil
}

func jsonResponseCount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		keys := r.URL.Query()
		fmt.Println(keys)
		var sqlString string
		sqlString = "SELECT id, board, timestamp, temperature, humidity, pressure FROM sensorsdata"
		if r.URL.Query().Get("board") != "" {
			sqlString = "SELECT id, board, timestamp, temperature, humidity, pressure FROM sensorsdata WHERE board='" + r.URL.Query().Get("board") + "'"
		}
		fmt.Println(sqlString)
		w.Header().Set("Content-Type", "application/json")
		response, _ := getJSON(sqlString)
		fmt.Fprintf(w, response)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func jsonResponse(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

		keys := r.URL.Query()
		fmt.Println(keys)
		var sqlString string

		sqlString = "SELECT id, board, timestamp, temperature, humidity, pressure FROM sensorsdata"

		if r.URL.Query().Get("board") != "" {
			sqlString = sqlString + " WHERE board='" + r.URL.Query().Get("board") + "'"
		}

		// SORTING
		var sortBy string = "timestamp"
		if r.URL.Query().Get("sort_by") != "" {
			sortBy = r.URL.Query().Get("sort_by")
		}

		var sortDir string = "DESC"
		if r.URL.Query().Get("sort") != "" {
			sortDir = r.URL.Query().Get("sort")
		}

		sqlString = sqlString + " ORDER BY " + sortBy + " " + sortDir

		fmt.Println(sqlString)
		w.Header().Set("Content-Type", "application/json")
		response, _ := getJSON(sqlString)
		fmt.Fprintf(w, response)

	default:
		fmt.Fprintf(w, "Sorry, only GET")
	}
}

func htmlHelpResponse(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/" {
	// 	http.Error(w, "404 not found.", http.StatusNotFound)
	// 	return
	// }

	switch r.Method {
	case "GET":

		w.Header().Set("Content-Type", "application/json")
		response, _ := getJSON("SELECT id, board, timestamp, temperature, humidity, pressure FROM sensorsdata")
		fmt.Fprintf(w, response)

	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		name := r.FormValue("name")
		address := r.FormValue("address")
		fmt.Fprintf(w, "Name = %s\n", name)
		fmt.Fprintf(w, "Address = %s\n", address)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}
