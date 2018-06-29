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
	"github.com/ahlusar1989/scheduling_service/v1/log"

	// "gopkg.in/gomail.v2"
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
