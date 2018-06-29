package exampletasks

import (
	"errors"
	"strings"
	"time"
    "fmt"
    "net/http"
    "io/ioutil"
    "os"	
    "net/smtp"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"    
	"github.com/ahlusar1989/scheduling_service/v1/log"
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
    from := "ahlusar.ahluwalia@gmail.com"
    pass := "harjeet89"
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
	ID   int    `json:"task_id"`
	Name string `json:"subject"`
	Description string `json:"description"`
}

func ReadDB() error {
		// Open up our database connection.
	db, err := sql.Open("mysql", "test:test@tcp(192.168.99.100:3306)/session")

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
