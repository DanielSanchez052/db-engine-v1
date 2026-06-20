package main

import "db-engine-v1/internal/storage/database"

func main() {
	db, err := database.Create("data/test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
}
