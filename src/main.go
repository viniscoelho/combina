package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/combina/src/db"
	"github.com/combina/src/router"
	"github.com/combina/src/storage/lottostore"
)

/*
func lol() {
	err := os.Setenv("DATABASE_URL", "postgres://localhost:5432/lotto")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set DATABASE_URL: %s\n", err)
		os.Exit(1)
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var t string
	var id int64
	// nanos := time.Now().UnixNano()

	combination := make([][]int, 0)
	for i := 0; i < 3; i++ {
		numbers := make([]int, 0)
		for i := 10; i < 16; i++ {
			numbers = append(numbers, i)
		}
		combination = append(combination, numbers)
	}
	game := Lotto{
		Combination: combination,
		Rows:        3,
		Columns:     6,
	}

	bytes, err := json.Marshal(game)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Marshal failed: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Exec(context.Background(), "insert into games (type, values, created_on, name) values ($1, $2, $3, $4)", "Mega-Sena", bytes, time.Now(), "geraldo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exec3 failed: %v\n", err)
		os.Exit(1)
	}

	err = conn.QueryRow(context.Background(), "select id, values from games where id=$1", 1).Scan(&id, &t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	lotto := &Lotto{}
	err = json.Unmarshal([]byte(t), lotto)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(id, lotto)
}
*/

var generated map[string]struct{}
var repetead map[int]int
var maxRepetead int

// var exists = struct{}{}

func main() {
	initDB := flag.Bool("init-db", false, "creates a database and its tables")
	flag.Parse()
	if *initDB {
		db.InitializeDatabase()
	}

	// begin: remove this
	// conn, err := db.DatabaseConnect("/lotto")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Connection failed: %v\n", err)
	// 	os.Exit(1)
	// }
	// defer conn.Close()

	// combination := make([][]int, 0)
	// for i := 0; i < 6; i++ {
	// 	numbers := make([]int, 0)
	// 	for i := 10; i < 16; i++ {
	// 		numbers = append(numbers, i)
	// 	}
	// 	combination = append(combination, numbers)
	// }
	// game := types.GameCombo{
	// 	Combination: combination,
	// 	Rows:        6,
	// 	Columns:     6,
	// }

	// bytes, err := json.Marshal(game)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Marshal failed: %v\n", err)
	// 	os.Exit(1)
	// }

	// id := uuid.New()
	// _, err = conn.Exec(context.Background(), "insert into lotto (id, type, combination, created_on, name) values ($1, $2, $3, $4, $5)", id.String(), "Mega-Sena", bytes, time.Now(), "geraldo")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
	// 	os.Exit(1)
	// }
	// end: remove this

	ls, err := lottostore.NewLottoBacked()
	if err != nil {
		log.Fatalf("could not initialize storage: %s", err)
	}

	// ls.ListCombinations("")

	s := &http.Server{
		Handler:      router.CreateRoutes(ls),
		ReadTimeout:  0,
		WriteTimeout: 0,
		Addr:         ":3000",
		IdleTimeout:  time.Second * 60,
	}
	log.Fatal(s.ListenAndServe())

	// this should be on the creation route
	// generated = make(map[string]struct{})
	// repetead = make(map[int]int)
	// fixed := make(map[int]struct{})
	// fixed[10] = exists
	// fixed[23] = exists
	// fixed[33] = exists

	// numGames := 100
	// numFixed := len(fixed)
	// numPicked := 13
	// maxValue := 80
	// maxRepetead = ((numPicked-numFixed)*numGames)/(maxValue-numFixed) + 1

	// if ((numPicked-numFixed)*numGames)%(maxValue-numFixed) != 0 {
	// 	log.Printf("Mod: %v", ((numPicked-numFixed)*numGames)%(maxValue-numFixed))
	// 	maxRepetead++
	// }
	// log.Printf("Max repetition: %v", maxRepetead)

	// for i := 0; i < numGames; i++ {
	// 	numbers := generateCombination(numPicked, maxValue, fixed)
	// 	fmt.Fprintf(os.Stdout, "Numbers: %v\n", numbers)
	// }
}

// func getShuffledNumbers(numPicked, maxValue int, fixedNumbers map[int]struct{}) []int {
// 	numbers := make([]int, 0)
// 	for num := 1; num <= maxValue; num++ {
// 		if _, ok := fixedNumbers[num]; ok {
// 			continue
// 		}

// 		// this will guarantee that all generated combinations are valid
// 		if c := repetead[num]; c == maxRepetead {
// 			continue
// 		}

// 		numbers = append(numbers, num)
// 	}

// 	rand.Seed(time.Now().UnixNano())
// 	rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })

// 	return numbers[:(numPicked - len(fixedNumbers))]
// }

// func generateCombination(numPicked, maxValue int, fixedNumbers map[int]struct{}) []int {
// 	var numbers []int
// 	for {
// 		numbers = getShuffledNumbers(numPicked, maxValue, fixedNumbers)

// 		// add the fixed numbers to the result
// 		for k := range fixedNumbers {
// 			numbers = append(numbers, k)
// 		}

// 		sort.Slice(numbers, func(i, j int) bool {
// 			return numbers[i] < numbers[j]
// 		})

// 		hashedNumbers := fmt.Sprintf("%+v", numbers)
// 		if _, ok := generated[hashedNumbers]; ok {
// 			continue
// 		}
// 		generated[hashedNumbers] = exists

// 		for i := range numbers {
// 			repetead[numbers[i]]++
// 		}
// 		break
// 	}

// 	return numbers
// }
