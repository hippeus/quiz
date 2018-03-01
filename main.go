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

var csvFilename = flag.String("f", "problems.csv", "csv file in the format of 'problem, correct answer'")
var timeLimit = flag.Int("t", 30, "the time limit for the quiz in seconds")

type Problems = []problem

func main() {
	flag.Parse()

	app := filepath.Dir(os.Args[0])
	file, err := os.Open(filepath.Join(app, *csvFilename))
	if err != nil {
		log.Fatalf("Failed to open the CSV file: %s\n", *csvFilename)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	csv := csv.NewReader(file)
	records, err := csv.ReadAll()
	if err != nil {
		log.Printf("Couldn't parse CSV file: %v\n", err)
	}
	problems := questionsPool(records)

	scan := bufio.NewScanner(os.Stdin)
	answers := make([]int, len(problems))
	resp := make(chan int)
	timeout := time.Tick(time.Duration(*timeLimit) * time.Second)

quizloop:
	for i, p := range problems {
		fmt.Printf("Question #%d: %s\n", i+1, p.q)
		go func(<-chan int) {
			scan.Scan()
			if scan.Err() != nil {
				log.Fatal("scanner internal error")
			}
			ans := scan.Text()
			value, err := strconv.Atoi(ans)
			if err != nil {
				log.Println(err)
			}
			resp <- value
		}(resp)

		select {
		case ans := <-resp:
			answers[i] = ans
		case <-timeout:
			fmt.Println("Time ran out, sorry!")
			break quizloop
		}
	}

	var score int
	for i, a := range answers {
		problem := problems[i]
		if a == problem.ans {
			score++
		}
	}
	fmt.Printf("You scored: %d out of %d\n", score, len(problems))
}

type problem struct {
	q   string
	ans int
}


func questionsPool(r [][]string) Problems {
	quiz := make(Problems, len(r))	
	for i, v := range r {
		var p problem
		p.q = v[0]
		num, err := strconv.Atoi(v[1])
		if err != nil {
			log.Fatalln("CSV answer field is not convertable to integer")
		}
		p.ans = num
		quiz[i] = p
	}
	return quiz
}