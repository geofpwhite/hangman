package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

func main() {
	db, _ := sql.Open("sqlite3", "words.db")
	content, _ := os.ReadFile("words.txt")
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line != "" {
			line = strings.ToLower(line)
			db.Exec("insert into words (word) values ('" + line + "')")
		}
	}
}
