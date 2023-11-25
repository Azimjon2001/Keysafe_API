package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Response struct {
	Message string `json:"message"`
}

type User struct {
	ID            int    `json:"id"`
	Password_name string `json:"password_name"`
	Login         string `json:"login"`
	Password      string `json:"password"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "user =  password =  dbname= sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Соединение с базой данных PostgreSql успешно установлено!")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		//log.Println("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("Insert into keysafe (password_name, login, password) values ($1, $2, $3)", user.Password_name, user.Login, user.Password)
	if err != nil {
		http.Error(w, "Error inserting data into the database", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error getting rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 1 {
		w.Header().Set("Content-Type", "application/json")
		jsonEncoder := json.NewEncoder(w)
		w.WriteHeader(http.StatusCreated)
		jsonEncoder.SetIndent("", "    ")
		json.NewEncoder(w).Encode(Response{Message: "User created syccessfully"})
	} else {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID parameter", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow("select id, password_name, login, password from keysafe where id = $1", userID).Scan(&user.ID, &user.Password_name, &user.Login, &user.Password)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		//log.Println("Error querying the database:", err)
		http.Error(w, "Error querying the database-1", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent("", "    ")
	jsonEncoder.Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID parameter", http.StatusBadRequest)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("Update keysafe set password_name = $1, login = $2, password = $3 where id = $4", &user.Password_name, &user.Login, &user.Password, &userID)
	if err != nil {
		http.Error(w, "Error updating user in the database", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent("", "    ")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "User updated syccessfully"})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID parameter", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("delete from keysafe where id = $1", userID)
	rowsAffected, _ := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error deleting user from the database", http.StatusInternalServerError)
		return
	} else if rowsAffected == 0 {
		http.Error(w, "We cannot find this user", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "User deleted syccessfully"})
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, password_name, login, password FROM keysafe order by id")
	if err != nil {
		http.Error(w, "Error querying the database"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var users []User

	for rows.Next() {
		var user User

		if err := rows.Scan(&user.ID, &user.Password_name, &user.Login, &user.Password); err != nil {
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent("", "    ")
	jsonEncoder.Encode(users)
}

func main() {
	http.HandleFunc("/api/users/create", createUser)
	http.HandleFunc("/api/users/get", getUser)
	http.HandleFunc("/api/users/update", updateUser)
	http.HandleFunc("/api/users/delete", deleteUser)
	http.HandleFunc("/api/users/getallusers", getAllUsers)
	port := 8080

	fmt.Printf("Сервер успешно запущен на порту %d...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Error starting the server: ", err)
	}
}
