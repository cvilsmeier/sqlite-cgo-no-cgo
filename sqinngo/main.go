package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

var perfTune = `
pragma journal_mode = WAL;
pragma synchronous = normal;
pragma temp_store = memory;
pragma mmap_size = 30000000000;`

func main() {
	sq := sqinn.MustLaunch(sqinn.Options{
		SqinnPath: "/home/cv/tmp/sqinn",
	})
	defer sq.Terminate()
	sq.Open(":memory:")
	defer sq.Close()

	wordsFile, err := os.ReadFile("../words")
	if err != nil {
		panic(err)
	}
	words := strings.Split(string(wordsFile), "\n")

	rand.Seed(0)
	times := 10
	rowTests := []int{10_000, 479_827, 4_798_270}

	sq.MustExecOne(perfTune)

	fmt.Println("time,rows,category,version")

	for _, rows := range rowTests {
		var insertDurations []time.Duration
		var groupByDurations []time.Duration
		for i := 0; i < times; i++ {
			sq.MustExecOne("DROP TABLE IF EXISTS people")
			sq.MustExecOne(`
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

			t1 := time.Now()
			insertValues := make([]any, 0, rows*8)
			for i := 0; i < rows; i++ {
				rnd_name := words[rand.Int()%len(words)]
				rnd_country := words[rand.Int()%len(words)]
				rnd_region := words[rand.Int()%len(words)]
				rnd_occupation := words[rand.Int()%len(words)]
				rnd_company := words[rand.Int()%len(words)]
				rnd_age := rand.Int() % 110
				rnd_fav_team := words[rand.Int()%len(words)]
				rnd_fav_sport := words[rand.Int()%len(words)]
				insertValues = append(insertValues,
					rnd_name,
					rnd_country,
					rnd_region,
					rnd_occupation,
					rnd_company,
					rnd_age,
					rnd_fav_team,
					rnd_fav_sport,
				)
			}
			sq.MustExec("INSERT INTO people VALUES (?, ?, ?, ?, ?, ?, ?, ?)", rows, 8, insertValues)
			insertDurations = append(insertDurations, time.Since(t1))
			fmt.Printf("%f,%d,insert,sqinngo\n", float64(time.Since(t1))/1e9, rows)

			t1 = time.Now()
			resultRows := sq.MustQuery("SELECT COUNT(1), age FROM people GROUP BY age ORDER BY COUNT(1) DESC", nil, []byte{sqinn.ValInt, sqinn.ValInt})
			var totalCount int
			var resultCount int
			var resultAge int
			for _, row := range resultRows {
				resultCount = row.Values[0].Int.Value
				resultAge = row.Values[1].Int.Value
				totalCount += resultCount
				_ = resultAge
			}
			groupByDurations = append(groupByDurations, time.Since(t1))
			if totalCount != rows {
				panic(fmt.Sprintf("totalCount %d != rows %d", totalCount, rows))
			}
			fmt.Printf("%f,%d,group_by,sqinngo\n", float64(time.Since(t1))/1e9, rows)
		}
		fmt.Printf("sqinngo,%d,insert-avg,%f\n", rows, float64(avg(insertDurations))/1e9)
		fmt.Printf("sqinngo,%d,groupBy-avg,%f\n", rows, float64(avg(groupByDurations))/1e9)
	}
}

func avg(values []time.Duration) time.Duration {
	var sum time.Duration
	for _, v := range values {
		sum += v
	}
	return sum / time.Duration(len(values))
}
