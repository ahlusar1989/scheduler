package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	opentracing_log "github.com/opentracing/opentracing-go/log"

	uuid "github.com/satori/go.uuid"

	"github.com/ahlusar1989/scheduling_service/v1"
	"github.com/ahlusar1989/scheduling_service/v1/config"
	"github.com/ahlusar1989/scheduling_service/v1/log"
	"github.com/ahlusar1989/scheduling_service/v1/tasks"
	"github.com/urfave/cli"

	exampletasks "github.com/ahlusar1989/scheduling_service/example/tasks"
	tracers "github.com/ahlusar1989/scheduling_service/example/tracers"
)

var (
	app        *cli.App
	configPath string
)

func init() {
	// Initialize a cli app
	app = cli.NewApp()
	app.Name = "scheduling_service"
	app.Usage = "Scheduling worker and send example tasks with scheduling_service send"
	app.Author = "Saran Ahluwalia"
	app.Email = "ahlusar.ahluwalia@gmail.com"
	app.Version = "0.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "c",
			Value:       "",
			Destination: &configPath,
			Usage:       "Path to a configuration file",
		},
	}
}

func main() {
	// Set the CLI app commands
	app.Commands = []cli.Command{
		{
			Name:  "worker",
			Usage: "launch scheduling_service worker",
			Action: func(c *cli.Context) error {
				if err := worker(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
		{
			Name:  "send",
			Usage: "send example tasks ",
			Action: func(c *cli.Context) error {
				if err := send(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
	}

	// Run the CLI app
	app.Run(os.Args)
}

func loadConfig() (*config.Config, error) {
	if configPath != "" {
		return config.NewFromYaml(configPath, true)
	}

	return config.NewFromEnvironment(true)
}

func startServer() (*scheduling_service.Server, error) {
	cnf, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Create server instance
	server, err := scheduling_service.NewServer(cnf)
	if err != nil {
		return nil, err
	}

	// Register tasks
	tasks := map[string]interface{}{
		"add":               exampletasks.Add,
		"multiply":          exampletasks.Multiply,
		"sum_ints":          exampletasks.SumInts,
		"sum_floats":        exampletasks.SumFloats,
		"concat":            exampletasks.Concat,
		"split":             exampletasks.Split,
		"panic_task":        exampletasks.PanicTask,
		"long_running_task": exampletasks.LongRunningTask,
		"email_task":        exampletasks.SendEmail,
		"make_request":      exampletasks.MakeRequest,
		"read_db":           exampletasks.ReadDB,
		"write_db":          exampletasks.MakeConcurrentWrites,
	}

	return server, server.RegisterTasks(tasks)
}

func worker() error {
	consumerTag := "scheduling_worker"

	cleanup, err := tracers.SetupTracer(consumerTag)
	if err != nil {
		log.FATAL.Fatalln("Unable to instantiate a tracer:", err)
	}
	defer cleanup()

	server, err := startServer()
	if err != nil {
		return err
	}

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker := server.NewWorker(consumerTag, 0)

	return worker.Launch()
}

func send() error {
	cleanup, err := tracers.SetupTracer("sender")
	if err != nil {
		log.FATAL.Fatalln("Unable to instantiate a tracer:", err)
	}
	defer cleanup()

	server, err := startServer()
	if err != nil {
		return err
	}

	var (
		addTask0, addTask1, addTask2                      tasks.Signature
		multiplyTask0, multiplyTask1                      tasks.Signature
		sumIntsTask, sumFloatsTask, concatTask, splitTask tasks.Signature
		panicTask                                         tasks.Signature
		longRunningTask                                   tasks.Signature
		emailTask                                         tasks.Signature
		makeRequest                                       tasks.Signature
		readDBMethod                                      tasks.Signature
		writeDBMethod                                     tasks.Signature
	)

	var initTasks = func() {
		addTask0 = tasks.Signature{
			Name: "add",
			Args: []tasks.Arg{
				{
					Type:  "int64",
					Value: 1,
				},
				{
					Type:  "int64",
					Value: 1,
				},
			},
		}

		addTask1 = tasks.Signature{
			Name: "add",
			Args: []tasks.Arg{
				{
					Type:  "int64",
					Value: 2,
				},
				{
					Type:  "int64",
					Value: 2,
				},
			},
		}

		addTask2 = tasks.Signature{
			Name: "add",
			Args: []tasks.Arg{
				{
					Type:  "int64",
					Value: 5,
				},
				{
					Type:  "int64",
					Value: 6,
				},
			},
		}

		multiplyTask0 = tasks.Signature{
			Name: "multiply",
			Args: []tasks.Arg{
				{
					Type:  "int64",
					Value: 4,
				},
			},
		}

		multiplyTask1 = tasks.Signature{
			Name: "multiply",
		}

		sumIntsTask = tasks.Signature{
			Name: "sum_ints",
			Args: []tasks.Arg{
				{
					Type:  "[]int64",
					Value: []int64{1, 2},
				},
			},
		}

		sumFloatsTask = tasks.Signature{
			Name: "sum_floats",
			Args: []tasks.Arg{
				{
					Type:  "[]float64",
					Value: []float64{1.5, 2.7},
				},
			},
		}

		concatTask = tasks.Signature{
			Name: "concat",
			Args: []tasks.Arg{
				{
					Type:  "[]string",
					Value: []string{"foo", "bar"},
				},
			},
		}

		splitTask = tasks.Signature{
			Name: "split",
			Args: []tasks.Arg{
				{
					Type:  "string",
					Value: "foo",
				},
			},
		}

		panicTask = tasks.Signature{
			Name: "panic_task",
		}

		longRunningTask = tasks.Signature{
			Name: "long_running_task",
		}

		emailTask = tasks.Signature{
			Name: "email_task",
		}

		makeRequest = tasks.Signature{
			Name: "make_request",
		}

		readDBMethod = tasks.Signature{
			Name: "read_db",
		}

		writeDBMethod = tasks.Signature{
			Name: "write_db",
		}
	}

	/*
	 * Lets start a span representing this run of the `send` command and
	 * set a batch id as baggage so it can travel all the way into
	 * the worker functions.
	 */
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "send")
	defer span.Finish()

	batchUUID, err := uuid.NewV4()

	if err != nil {
		return fmt.Errorf("Error generating batch id: %s", err.Error())
	}

	batchID := batchUUID.String()
	span.SetBaggageItem("batch.id", batchID)
	span.LogFields(opentracing_log.String("batch.id", batchID))

	log.INFO.Println("Starting batch:", batchID)
	/*
	 * First, let's try sending some individual tasks
	 */
	initTasks()

	log.INFO.Println("Single task:")

	asyncResult, err := server.SendTaskWithContext(ctx, &addTask0)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	results, err := asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("1 + 1 = %v\n", tasks.HumanReadableResults(results))

	asyncResult, err = server.SendTaskWithContext(ctx, &emailTask)
	if err != nil {
		return fmt.Errorf("Could not send email task: %s", err.Error())
	}

	asyncResult, err = server.SendTaskWithContext(ctx, &makeRequest)
	if err != nil {
		return fmt.Errorf("Could not send request task: %s", err.Error())
	}

	asyncResult, err = server.SendTaskWithContext(ctx, &readDBMethod)
	if err != nil {
		return fmt.Errorf("Could not send db task: %s", err.Error())
	}

	asyncResult, err = server.SendTaskWithContext(ctx, &writeDBMethod)
	if err != nil {
		return fmt.Errorf("Could not send write to db task: %s", err.Error())
	}

	/*
	 * Try couple of tasks with a slice argument and slice return value
	 */
	asyncResult, err = server.SendTaskWithContext(ctx, &sumIntsTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	results, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("sum([1, 2]) = %v\n", tasks.HumanReadableResults(results))

	asyncResult, err = server.SendTaskWithContext(ctx, &sumFloatsTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	results, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("sum([1.5, 2.7]) = %v\n", tasks.HumanReadableResults(results))

	asyncResult, err = server.SendTaskWithContext(ctx, &concatTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	results, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("concat([\"foo\", \"bar\"]) = %v\n", tasks.HumanReadableResults(results))

	asyncResult, err = server.SendTaskWithContext(ctx, &splitTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	results, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("split([\"foo\"]) = %v\n", tasks.HumanReadableResults(results))

	/*
	 * Now let's explore ways of sending multiple tasks
	 */

	// Now let's try a parallel execution
	initTasks()
	log.INFO.Println("Group of tasks (parallel execution):")

	group, err := tasks.NewGroup(&addTask0, &addTask1, &addTask2)
	if err != nil {
		return fmt.Errorf("Error creating group: %s", err.Error())
	}

	asyncResults, err := server.SendGroupWithContext(ctx, group, 10)
	if err != nil {
		return fmt.Errorf("Could not send group: %s", err.Error())
	}

	for _, asyncResult := range asyncResults {
		results, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
		if err != nil {
			return fmt.Errorf("Getting task result failed with error: %s", err.Error())
		}
		log.INFO.Printf(
			// "%v + %v = %v\n",
			// asyncResult.Signature.Args[0].Value,
			// asyncResult.Signature.Args[1].Value,
			tasks.HumanReadableResults(results),
		)
	}

	// Now let's try a group with a chord
	initTasks()
	log.INFO.Println("Group of tasks with a callback (chord):")

	group, err = tasks.NewGroup(&addTask0, &addTask1, &addTask2)
	if err != nil {
		return fmt.Errorf("Error creating group: %s", err.Error())
	}

	chord, err := tasks.NewChord(group, &multiplyTask1)
	if err != nil {
		return fmt.Errorf("Error creating chord: %s", err)
	}

	chordAsyncResult, err := server.SendChordWithContext(ctx, chord, 10)
	if err != nil {
		return fmt.Errorf("Could not send chord: %s", err.Error())
	}

	results, err = chordAsyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting chord result failed with error: %s", err.Error())
	}
	log.INFO.Printf("(1 + 1) * (2 + 2) * (5 + 6) = %v\n", tasks.HumanReadableResults(results))

	// Now let's try chaining task results
	initTasks()
	log.INFO.Println("Chain of tasks:")

	chain, err := tasks.NewChain(&addTask0, &addTask1, &addTask2, &multiplyTask0)
	if err != nil {
		return fmt.Errorf("Error creating chain: %s", err)
	}

	chainAsyncResult, err := server.SendChainWithContext(ctx, chain)
	if err != nil {
		return fmt.Errorf("Could not send chain: %s", err.Error())
	}

	results, err = chainAsyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting chain result failed with error: %s", err.Error())
	}
	log.INFO.Printf("(((1 + 1) + (2 + 2)) + (5 + 6)) * 4 = %v\n", tasks.HumanReadableResults(results))

	// Let's try a task which throws panic to make sure stack trace is not lost
	initTasks()
	asyncResult, err = server.SendTaskWithContext(ctx, &panicTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	_, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err == nil {
		return errors.New("Error should not be nil if task panicked")
	}
	log.INFO.Printf("Task panicked and returned error = %v\n", err.Error())

	// Let's try a long running task
	initTasks()
	asyncResult, err = server.SendTaskWithContext(ctx, &longRunningTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	results, err = asyncResult.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		return fmt.Errorf("Getting long running task result failed with error: %s", err.Error())
	}
	log.INFO.Printf("Long running task returned = %v\n", tasks.HumanReadableResults(results))

	return nil
}
