package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Routine struct {
	RoutineName string `json:"routineName"`
	Day         string `json:"day"`
	Time        string `json:"time"`
}

func addRoutineHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var routine Routine
		err := json.NewDecoder(r.Body).Decode(&routine)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// SQL Query to insert data
		_, err = db.Exec("INSERT INTO routines (routine_name, day, time) VALUES ($1, $2, $3)",
			routine.RoutineName, routine.Day, routine.Time)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func getRoutinesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM routines")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		routines := make([]Routine, 0)
		for rows.Next() {
			var routine Routine
			err := rows.Scan(&routine.RoutineName, &routine.Day, &routine.Time)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			routines = append(routines, routine)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(routines)
	}
}

func main() {
	connStr := "postgres://postgres:Clever-Rival9@localhost/dbname?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/add-routine", addRoutineHandler(db))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
