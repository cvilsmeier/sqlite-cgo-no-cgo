package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var perfTune = `
pragma journal_mode = WAL;
pragma synchronous = normal;
pragma temp_store = memory;
pragma mmap_size = 30000000000;`

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	wordsFile, err := os.ReadFile("../words")
	if err != nil {
		panic(err)
	}
	words := strings.Split(string(wordsFile), "\n")

	rand.Seed(0)
	times := 10
	rowTests := []int{10_000, 479_827, 4_798_270}

	_, err = db.Exec(perfTune)
	if err != nil {
		panic(err)
	}

	fmt.Println("time,rows,category,version")

	for _, rows := range rowTests {
		var insertDurations []time.Duration
		var groupByDurations []time.Duration
		for i := 0; i < times; i++ {
			_, err = db.Exec("DROP TABLE IF EXISTS people")
			if err != nil {
				panic(err)
			}
			_, err = db.Exec(`
CREATE TABLE people (
  name TEXT,
  country TEXT,
  region TEXT,
  occupation TEXT,
  age INT,
  company TEXT,
  favorite_team TEXT,
  favorite_sport TEXT
)`)
			if err != nil {
				panic(err)
			}

			t1 := time.Now()
			stmt, err := db.Prepare("INSERT INTO people VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
			if err != nil {
				panic(err)
			}
			for i := 0; i < rows; i++ {
				rnd_name := words[rand.Int()%len(words)]
				rnd_country := words[rand.Int()%len(words)]
				rnd_region := words[rand.Int()%len(words)]
				rnd_occupation := words[rand.Int()%len(words)]
				rnd_company := words[rand.Int()%len(words)]
				rnd_age := rand.Int() % 110
				rnd_fav_team := words[rand.Int()%len(words)]
				rnd_fav_sport := words[rand.Int()%len(words)]
				_, err := stmt.Exec(rnd_name, rnd_country, rnd_region, rnd_occupation, rnd_age, rnd_company, rnd_fav_team, rnd_fav_sport)
				if err != nil {
					panic(err)
				}
			}
			insertDurations = append(insertDurations, time.Since(t1))
			fmt.Printf("%f,%d,insert,cgo\n", float64(time.Since(t1))/1e9, rows)

			t1 = time.Now()
			resultRows, err := db.Query("SELECT COUNT(1), age FROM people GROUP BY age ORDER BY COUNT(1) DESC")
			if err != nil {
				panic(err)
			}
			var totalCount int
			var resultCount int
			var resultAge int
			for resultRows.Next() {
				resultRows.Scan(&resultCount, &resultAge)
				totalCount += resultCount
			}
			resultRows.Close()
			if totalCount != rows {
				panic(fmt.Sprintf("totalCount %d != rows %d", totalCount, rows))
			}
			groupByDurations = append(groupByDurations, time.Since(t1))
			fmt.Printf("%f,%d,group_by,cgo\n", float64(time.Since(t1))/1e9, rows)
		}
		fmt.Printf("cgo,%d,insert-avg,%f\n", rows, float64(avg(insertDurations))/1e9)
		fmt.Printf("cgo,%d,groupBy-avg,%f\n", rows, float64(avg(groupByDurations))/1e9)
	}
}

func avg(values []time.Duration) time.Duration {
	var sum time.Duration
	for _, v := range values {
		sum += v
	}
	return sum / time.Duration(len(values))
}
