package exampletasks

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ahlusar1989/scheduling_service/v1/log"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"
)

var (
	DBW *sql.DB

	DSN      = "test:test@tcp(192.168.99.100:3306)/session"
	SQLQuery = "INSERT INTO `tasks`(`task_id`, `subject`, `description`)  VALUES(?, ?, ?)"
	StmtMain *sql.Stmt
	wg       sync.WaitGroup
)

// Add ...
func Add(args ...int64) (int64, error) {
	sum := int64(0)
	for _, arg := range args {
		sum += arg
	}
	return sum, nil
}

// Multiply ...
func Multiply(args ...int64) (int64, error) {
	sum := int64(1)
	for _, arg := range args {
		sum *= arg
	}
	return sum, nil
}

// SumInts ...
func SumInts(numbers []int64) (int64, error) {
	var sum int64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}

// SumFloats ...
func SumFloats(numbers []float64) (float64, error) {
	var sum float64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}

// Concat ...
func Concat(strs []string) (string, error) {
	var res string
	for _, s := range strs {
		res += s
	}
	return res, nil
}

// Split ...
func Split(str string) ([]string, error) {
	return strings.Split(str, ""), nil
}

// PanicTask ...
func PanicTask() (string, error) {
	panic(errors.New("oops"))
}

// LongRunningTask ...
func LongRunningTask() error {
	log.INFO.Print("Long running task started")
	for i := 0; i < 10; i++ {
		log.INFO.Print(10 - i)
		time.Sleep(1 * time.Second)
	}
	log.INFO.Print("Long running task finished")
	return nil
}

func SendEmail() error {
	log.INFO.Print("Start email task")
	from := os.Getenv("GMAIL_EMAIL")
	pass := os.Getenv("GMAIL_PASS")
	to := "ztc@mailinator.com"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		"Hello"

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.FATAL.Printf("smtp error: %s", err)
		return err
	}

	log.INFO.Print("sent, visit http://ztc.mailinator.com")
	log.INFO.Print("Finished email")
	return nil
}

func MakeRequest() error {
	response, err := http.Get("https://httpbin.org/get")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", string(contents))
	}
	return nil
}

type Task struct {
	ID          int    `json:"task_id"`
	Name        string `json:"subject"`
	Description string `json:"description"`
}

func ReadDB() error {
	// Open up our database connection.
	db, err := sql.Open("mysql", DSN)

	// if there is an error opening the connection, handle it
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		log.INFO.Print(err.Error())
		return err
	}
	defer db.Close()

	// Execute the query
	results, err := db.Query("SELECT task_id, subject, description FROM tasks")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error()) // proper error handling instead of panic in your app
		return err
	}

	for results.Next() {
		var task Task
		// for each row, scan the result into our tasks composite object
		err = results.Scan(&task.ID, &task.Name, &task.Description)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			panic(err.Error()) // proper error handling instead of panic in your app
			return err
		}
		// and then print out the tasks's Name attribute
		fmt.Printf("%s\n", task.Name)
		fmt.Printf("%d\n", task.ID)
		fmt.Printf("%s\n", task.Description)
	}
	return nil
}

func generateRandomString(n int) string {

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remainder := n-1, rand.Int63(), letterIdxMax; i >= 0; {

		if remainder == 0 {
			cache, remainder = rand.Int63(), letterIdxMax
		}

		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}

		cache >>= letterIdxBits
		remainder--
	}

	return string(b)
}

type Stub struct {
	ID          int
	Subject     string
	Description string
}

func Store(d Stub) {
	defer wg.Done()
	_, err := StmtMain.Exec(d.ID, d.Subject, d.Description)
	if err != nil {
		log.FATAL.Println(err)
	}
}

func MakeConcurrentWrites() error {

	var (
		errDbw  error
		errStmt error
	)
	concurrencyLevel := runtime.NumCPU() * 8

	DBW, errDbw = sql.Open("mysql", DSN)
	if errDbw != nil {
		log.FATAL.Println(errDbw)
		return errDbw
	}

	DBW.SetMaxIdleConns(concurrencyLevel)
	defer DBW.Close()

	StmtMain, errStmt = DBW.Prepare(SQLQuery)
	if errStmt != nil {
		log.FATAL.Println(errStmt)
		return errStmt
	}
	defer StmtMain.Close()
	//populate data
	data := Stub{
		ID:          2,
		Subject:     generateRandomString(45),
		Description: generateRandomString(45),
	}

	f, err := os.Create("cpuprofile")
	if err != nil {
		log.FATAL.Println(err)
		return err
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	t0 := time.Now()
	for i := 0; i < 1000000; {
		for k := 0; k < concurrencyLevel; k++ {
			i++
			if i > 1000000 {
				break
			}
			data.ID = i
			data.Subject = generateRandomString(45)
			data.Description = generateRandomString(45)
			wg.Add(1)
			go Store(data)
		}
		wg.Wait()
		if i > 1000000 {
			break
		}
	}
	t1 := time.Now()
	fmt.Printf("%v per second.\n", 1000000.0/t1.Sub(t0).Seconds())
	return nil
}
