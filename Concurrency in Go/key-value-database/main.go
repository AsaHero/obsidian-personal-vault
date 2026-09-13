package main

import (
	"bufio"
	"fmt"
	"key-value-database/compute"
	"key-value-database/storage"
	"log/slog"
	"os"
)

func main() {
	stdin := bufio.NewScanner(os.Stdin)
	db := storage.NewNegine()

	for stdin.Scan() {
		line := stdin.Text()
		query, err := compute.NewParser(line).Parse()
		if err != nil {
			slog.Error("failed to parse query", "error", err)
			continue
		}

		switch query.Cmd {
		case compute.GET:
			val, err := db.Get(query.Key)
			if err != nil {
				slog.Error("failed to get", "error", err)
				continue
			}

			fmt.Println(val)
		case compute.SET:
			db.Set(query.Key, query.Value)
			fmt.Println("set successfully")
		case compute.DEL:
			db.Del(query.Key)
			fmt.Println("deleted successfully")
		}
	}
}
