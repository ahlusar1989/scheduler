package scheduling_service_test

import (
	"testing"

	"github.com/ahlusar1989/scheduling_service/v1"
	"github.com/ahlusar1989/scheduling_service/v1/config"
	"github.com/stretchr/testify/assert"
)

func TestRegisterTasks(t *testing.T) {
	t.Parallel()

	server := getTestServer(t)
	err := server.RegisterTasks(map[string]interface{}{
		"test_task": func() error { return nil },
	})
	assert.NoError(t, err)

	_, err = server.GetRegisteredTask("test_task")
	assert.NoError(t, err, "test_task is not registered but it should be")
}

func TestRegisterTask(t *testing.T) {
	t.Parallel()

	server := getTestServer(t)
	err := server.RegisterTask("test_task", func() error { return nil })
	assert.NoError(t, err)

	_, err = server.GetRegisteredTask("test_task")
	assert.NoError(t, err, "test_task is not registered but it should be")
}

func TestGetRegisteredTask(t *testing.T) {
	t.Parallel()

	server := getTestServer(t)
	_, err := server.GetRegisteredTask("test_task")
	assert.Error(t, err, "test_task is registered but it should not be")
}

func TestGetRegisteredTaskNames(t *testing.T) {
	t.Parallel()

	server := getTestServer(t)

	taskName := "test_task"
	err := server.RegisterTask(taskName, func() error { return nil })
	assert.NoError(t, err)

	taskNames := server.GetRegisteredTaskNames()
	assert.Equal(t, 1, len(taskNames))
	assert.Equal(t, taskName, taskNames[0])
}

func getTestServer(t *testing.T) *scheduling_service.Server {
	server, err := scheduling_service.NewServer(&config.Config{
		Broker:        "amqp://guest:guest@localhost:5672/",
		DefaultQueue:  "scheduled_tasks",
		ResultBackend: "redis://127.0.0.1:6379",
		AMQP: &config.AMQPConfig{
			Exchange:      "scheduler_exchange",
			ExchangeType:  "direct",
			BindingKey:    "scheduled_task",
			PrefetchCount: 1,
		},
	})
	if err != nil {
		t.Error(err)
	}
	return server
}
