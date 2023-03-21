package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"path"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Exchange struct {
	Bid string `json:"bid"`
}

var (
	apiTimeOut = 250 * time.Millisecond
	dbTimeOut  = 10 * time.Millisecond
	apiUrl     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

func main() {
	db, err := sql.Open("sqlite3", path.Join("database", "database.db"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exchange (id INTEGER PRIMARY KEY AUTOINCREMENT, bid TEXT NOT NULL)")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, apiTimeOut)
		defer cancel()

		res, err := getCurrentExchange(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// prepare
		stmt, err := db.Prepare("INSERT INTO exchange (bid) VALUES (?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		ctx, cancel = context.WithTimeout(ctx, dbTimeOut)
		defer cancel()

		// execute
		_, err = stmt.ExecContext(ctx, res.Bid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	http.ListenAndServe(":8080", nil)
}

func getCurrentExchange(ctx context.Context) (*Exchange, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var exchange map[string]*Exchange
	err = json.NewDecoder(res.Body).Decode(&exchange)
	if err != nil {
		return nil, err
	}

	return exchange["USDBRL"], nil
}
