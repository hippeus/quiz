package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var csvsource = flag.String("f", "problems.csv", "path to quiz data, CSV format required")
var timeout = flag.Int("t", 30, "sets timeout in seconds per question")

func main() {
	flag.Parse()

	app := filepath.Dir(os.Args[0])
	file, err := os.Open(filepath.Join(app, *csvsource))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	csv := csv.NewReader(file)
	ss, err := csv.ReadAll()
	if err != nil {
		log.Printf("Couldn't parse %s entry: %v\n", ss, err)
	}

	scan := bufio.NewScanner(os.Stdin)
	answers := make([]int, len(ss))
	resp := make(chan int)

	for i, v := range ss {
		fmt.Printf("[%d] Question: %s\n", i+1, v[0])
		go func(<-chan int) {
			scan.Scan()
			if scan.Err() != nil {
				log.Fatal("scanner internal error")
			}
			input := scan.Text()
			value, err := strconv.Atoi(input)
			if err != nil {
				log.Println(err)
			}
			resp <- value
		}(resp)
		select {
		case ans := <-resp:
			answers[i] = ans
		case <-time.After(time.Duration(*timeout) * time.Second):
			fmt.Println("Too slow, next question")
			answers[i] = -1
		}
	}

	var score int
	for i, v := range answers {
		correctAns, err := strconv.Atoi(ss[i][1])
		if err != nil {
			log.Fatal(err)
		}
		if correctAns == v {
			score++
		}
	}
	fmt.Printf("You scored: %d of %d\n", score, len(answers))
}
