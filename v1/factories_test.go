package scheduling_service_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ahlusar1989/scheduling_service/v1"
	"github.com/ahlusar1989/scheduling_service/v1/config"
	"github.com/stretchr/testify/assert"

	amqpbroker "github.com/ahlusar1989/scheduling_service/v1/brokers/amqp"
	redisbroker "github.com/ahlusar1989/scheduling_service/v1/brokers/redis"
	sqsbroker "github.com/ahlusar1989/scheduling_service/v1/brokers/sqs"

	amqpbackend "github.com/ahlusar1989/scheduling_service/v1/backends/amqp"
	redisbackend "github.com/ahlusar1989/scheduling_service/v1/backends/redis"
)

var (
	redisSchemeTestCases = []struct {
		desc      string
		url       string
		host, pwd string
		db        int
		err       error
	}{
		{
			desc: "invalid redis scheme",
			url:  "non_redis://127.0.0.1:5672",
			err:  errors.New("invalid redis scheme"),
		},
		{
			desc: "empty redis scheme",
			url:  "redis:/",
		},
		{
			desc: "redis host",
			url:  "redis://127.0.0.1:5672",
			host: "127.0.0.1:5672",
		},
		{
			desc: "redis password and host",
			url:  "redis://pwd@127.0.0.1:5672",
			host: "127.0.0.1:5672",
			pwd:  "pwd",
		},
		{
			desc: "redis password, host and db",
			url:  "redis://pwd@127.0.0.1:5672/2",
			host: "127.0.0.1:5672",
			pwd:  "pwd",
			db:   2,
		},
		{
			desc: "redis user, password host",
			url:  "redis://user:pwd@127.0.0.1:5672",
			host: "127.0.0.1:5672",
			pwd:  "pwd",
		},
		{
			desc: "redis user, password with colon host",
			url:  "redis://user:pwd:with:colon@127.0.0.1:5672",
			host: "127.0.0.1:5672",
			pwd:  "pwd:with:colon",
		},
		{
			desc: "redis user, empty password and host",
			url:  "redis://user:@127.0.0.1:5672",
			host: "127.0.0.1:5672",
			pwd:  "",
		},
		{
			desc: "redis empty user, password and host",
			url:  "redis://:pwd@127.0.0.1:5672",
			host: "127.0.0.1:5672",
			pwd:  "pwd",
		},
	}
)

func TestBrokerFactory(t *testing.T) {
	t.Parallel()

	var cnf config.Config

	// 1) AMQP broker test

	cnf = config.Config{
		Broker:       "amqp://guest:guest@localhost:5672/",
		DefaultQueue: "scheduled_tasks",
		AMQP: &config.AMQPConfig{
			Exchange:      "machinery_exchange",
			ExchangeType:  "direct",
			BindingKey:    "scheduled_task",
			PrefetchCount: 1,
		},
	}

	actual, err := scheduling_service.BrokerFactory(&cnf)
	if assert.NoError(t, err) {
		_, isAMQPBroker := actual.(*amqpbroker.Broker)
		assert.True(
			t,
			isAMQPBroker,
			"Broker should be instance of *brokers.AMQPBroker",
		)
		expected := amqpbroker.New(&cnf)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// 2) Redis broker test

	// with password
	cnf = config.Config{
		Broker:       "redis://password@localhost:6379",
		DefaultQueue: "scheduled_tasks",
	}

	actual, err = scheduling_service.BrokerFactory(&cnf)
	if assert.NoError(t, err) {
		_, isRedisBroker := actual.(*redisbroker.Broker)
		assert.True(
			t,
			isRedisBroker,
			"Broker should be instance of *brokers.RedisBroker",
		)
		expected := redisbroker.New(&cnf, "localhost:6379", "password", "", 0)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// without password
	cnf = config.Config{
		Broker:       "redis://localhost:6379",
		DefaultQueue: "scheduled_tasks",
	}

	actual, err = scheduling_service.BrokerFactory(&cnf)
	if assert.NoError(t, err) {
		_, isRedisBroker := actual.(*redisbroker.Broker)
		assert.True(
			t,
			isRedisBroker,
			"Broker should be instance of *brokers.RedisBroker",
		)
		expected := redisbroker.New(&cnf, "localhost:6379", "", "", 0)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// using a socket file
	cnf = config.Config{
		Broker:       "redis+socket:///tmp/redis.sock",
		DefaultQueue: "scheduled_tasks",
	}

	actual, err = scheduling_service.BrokerFactory(&cnf)
	if assert.NoError(t, err) {
		_, isRedisBroker := actual.(*redisbroker.Broker)
		assert.True(
			t,
			isRedisBroker,
			"Broker should be instance of *brokers.RedisBroker",
		)
		expected := redisbroker.New(&cnf, "", "", "/tmp/redis.sock", 0)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// 3) AWS SQS
	cnf = config.Config{
		Broker:       "https://sqs.us-east-2.amazonaws.com/123456789012",
		DefaultQueue: "scheduled_tasks",
	}

	actual, err = scheduling_service.BrokerFactory(&cnf)
	if assert.NoError(t, err) {
		_, isAWSSQSBroker := actual.(*sqsbroker.Broker)
		assert.True(
			t,
			isAWSSQSBroker,
			"Broker should be instance of *brokers.AWSSQSBroker",
		)
	}
}

func TestBrokerFactoryError(t *testing.T) {
	t.Parallel()

	cnf := config.Config{
		Broker: "BOGUS",
	}

	conn, err := scheduling_service.BrokerFactory(&cnf)
	if assert.Error(t, err) {
		assert.Nil(t, conn)
		assert.Equal(t, "Factory failed with broker URL: BOGUS", err.Error())
	}
}

func TestBackendFactory(t *testing.T) {
	t.Parallel()

	var cnf config.Config

	// 1) AMQP backend test

	cnf = config.Config{ResultBackend: "amqp://guest:guest@localhost:5672/"}

	actual, err := scheduling_service.BackendFactory(&cnf)
	if assert.NoError(t, err) {
		expected := amqpbackend.New(&cnf)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// 2) Redis backend test

	// with password
	cnf = config.Config{
		ResultBackend: "redis://password@localhost:6379",
	}

	actual, err = scheduling_service.BackendFactory(&cnf)
	if assert.NoError(t, err) {
		expected := redisbackend.New(&cnf, "localhost:6379", "password", "", 0)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// without password
	cnf = config.Config{
		ResultBackend: "redis://localhost:6379",
	}

	actual, err = scheduling_service.BackendFactory(&cnf)
	if assert.NoError(t, err) {
		expected := redisbackend.New(&cnf, "localhost:6379", "", "", 0)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

	// using a socket file
	cnf = config.Config{
		ResultBackend: "redis+socket:///tmp/redis.sock",
	}

	actual, err = scheduling_service.BackendFactory(&cnf)
	if assert.NoError(t, err) {
		expected := redisbackend.New(&cnf, "", "", "/tmp/redis.sock", 0)
		assert.True(
			t,
			reflect.DeepEqual(actual, expected),
			fmt.Sprintf("conn = %v, want %v", actual, expected),
		)
	}

}

func TestBackendFactoryError(t *testing.T) {
	t.Parallel()

	cnf := config.Config{
		ResultBackend: "BOGUS",
	}

	conn, err := scheduling_service.BackendFactory(&cnf)
	if assert.Error(t, err) {
		assert.Nil(t, conn)
		assert.Equal(t, "Factory failed with result backend: BOGUS", err.Error())
	}

	if conn != nil {
		t.Errorf("conn = %v, should be nil", conn)
	}
}

func TestParseRedisURL(t *testing.T) {
	t.Parallel()

	for _, tc := range redisSchemeTestCases {
		tc := tc // capture range variable
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			host, pwd, db, err := scheduling_service.ParseRedisURL(tc.url)
			if tc.err != nil {
				assert.Error(t, err, tc.err)
				return
			}

			if assert.NoError(t, err) {
				assert.Equal(t, tc.host, host)
				assert.Equal(t, tc.pwd, pwd)
				assert.Equal(t, tc.db, db)
			}
		})
	}
}

func TestParseRedisSocketURL(t *testing.T) {
	t.Parallel()

	var (
		path, pwd, url string
		db             int
		err            error
	)

	url = "non_redissock:///tmp/redis.sock"
	_, _, _, err = scheduling_service.ParseRedisSocketURL(url)
	assert.Error(t, err, "invalid redis scheme")

	url = "redis+socket:/"
	_, _, _, err = scheduling_service.ParseRedisSocketURL(url)
	assert.Error(t, err, "invalid redis url scheme")

	url = "redis+socket:///tmp/redis.sock"
	path, pwd, db, err = scheduling_service.ParseRedisSocketURL(url)
	if assert.NoError(t, err) {
		assert.Equal(t, "/tmp/redis.sock", path)
		assert.Equal(t, "", pwd)
		assert.Equal(t, 0, db)
	}

	url = "redis+socket://pwd@/tmp/redis.sock"
	path, pwd, db, _ = scheduling_service.ParseRedisSocketURL(url)
	if assert.NoError(t, err) {
		assert.Equal(t, "/tmp/redis.sock", path)
		assert.Equal(t, "pwd", pwd)
		assert.Equal(t, 0, db)
	}

	url = "redis+socket://pwd@/tmp/redis.sock:/2"
	path, pwd, db, err = scheduling_service.ParseRedisSocketURL(url)
	if assert.NoError(t, err) {
		assert.Equal(t, "/tmp/redis.sock", path)
		assert.Equal(t, "pwd", pwd)
		assert.Equal(t, 2, db)
	}
}
