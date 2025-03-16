package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Mannah struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	FROM_USER   int    `json:"from_user"`
	TO_USER     int    `json:"to_user"`
}

func main() {
	// Connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// create table users id should be primary key and auto increment
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS mannah (id SERIAL PRIMARY KEY, description TEXT, from_user INT, to_user INT)")
	db.Exec("ALTER TABLE mannah ADD CONSTRAINT fk_from FOREIGN KEY (from_user) REFERENCES users(id)")
	db.Exec("ALTER TABLE mannah ADD CONSTRAINT fk_to FOREIGN KEY (to_user) REFERENCES users(id)")

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/testdata", getTestData()).Methods("GET")
	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")
	router.HandleFunc("/users/{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/users/{id}", deleteUser(db)).Methods("DELETE")

	router.HandleFunc("/mannah/{userId}", getUserMannah(db)).Methods("GET")
	router.HandleFunc("/mannah/{userId}", createUserMannah(db)).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", jsonMiddleware(router)))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getTestData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Test Data")
	}
}

func getUserMannah(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userId := params["userId"]

		// get mannah for user
		rows, err := db.Query("SELECT * FROM mannah WHERE to_user = $1", userId)
		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()

		mannahs := []Mannah{}

		for rows.Next() {
			mannah := Mannah{}
			err := rows.Scan(&mannah.ID, &mannah.Description, &mannah.FROM_USER, &mannah.TO_USER)
			if err != nil {
				log.Fatal(err)
			}
			mannahs = append(mannahs, mannah)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mannahs)
	}
}

func createUserMannah(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m Mannah
		json.NewDecoder(r.Body).Decode(&m)

		db.QueryRow("INSERT INTO mannah (description, from, to) VALUES ($1, $2, $3, $4) RETURNING id", m.Description, m.FROM_USER, m.TO_USER)

		json.NewEncoder(w).Encode(m)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		//  insert user into database users
		err := db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", u.Name, u.Email).Scan(&u.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", u.Name, u.Email, id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("DELETE users WHERE id = $1", id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode("User Deleted")
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var u User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}

// get users
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()

		users := []User{}

		for rows.Next() {
			user := User{}
			err := rows.Scan(&user.ID, &user.Name, &user.Email)
			if err != nil {
				log.Fatal(err)
			}
			users = append(users, user)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}
}
