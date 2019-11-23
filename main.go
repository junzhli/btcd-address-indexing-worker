package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/btcd-address-indexing-worker/account"
	"github.com/btcd-address-indexing-worker/mongo"
	"github.com/btcd-address-indexing-worker/utils/logger"

	"github.com/go-bongo/bongo"
	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
)

type request struct {
	Account string `json:"account"`
	Task    string `json:"task"`
}

type responseBase struct {
	Command string `json:"command"`
	Account string `json:"account"`
}

type responseBalance struct {
	responseBase
	DataBalance float64 `json:"data"`
}

type responseTransactions struct {
	responseBase
	DataTx []string `json:"data"`
}

type responseUnspents struct {
	responseBase
	DataUspt []mongo.Unspent `json:"data"`
}

type responseAll struct {
	responseBase
	DataAll account.UserData `json:"data"`
}

// commands
const (
	CommandBalance      = "balance"
	CommandTransactions = "transactions"
	CommandUnspents     = "unspents"
	CommandAll          = "all"
)

const exAccountReq = "account_req"
const exAccountRet = "account_ret"

func doTask(wg *sync.WaitGroup, id int, c chan bool, d amqp.Delivery, config *account.Config, messageChannel *amqp.Channel) {
	defer wg.Done()
	c <- true
	lg := log.New(os.Stdout, "[Task "+strconv.Itoa(id)+"] ", log.LstdFlags)
	lg2 := logger.New(lg)
	lg.Printf("Received a message: %s", d.Body)

	var req request
	err := json.Unmarshal(d.Body, &req)
	if err != nil {
		lg2.LogOnError(err, "Failed to parse request message from receiverChannel")
	}

	if err == nil {
		lg.Printf("Task is requested with parameters: addr => " + req.Account + " task => " + req.Task)
		startTime := time.Now()
		var res []byte
		switch req.Task {
		case CommandBalance:
			balance, err := account.GetAddressBalance(lg, lg2, config, req.Account)
			if err != nil {
				lg2.LogOnError(err, "Fails on the task")
				break
			}
			res, err = json.Marshal(responseBalance{
				responseBase{
					CommandBalance,
					req.Account,
				},
				balance,
			})
		case CommandTransactions:
			transactions, err := account.GetAddressTransactions(lg, lg2, config, req.Account)
			if err != nil {
				lg2.LogOnError(err, "Fails on the task")
				break
			}
			res, err = json.Marshal(responseTransactions{
				responseBase{
					CommandTransactions,
					req.Account,
				},
				transactions,
			})
		case CommandUnspents:
			unspents, err := account.GetAddressUnspentOutputs(lg, lg2, config, req.Account)
			if err != nil {
				lg2.LogOnError(err, "Fails on the task")
				break
			}

			_unspents := make([]mongo.Unspent, 0)
			for _, val := range unspents {
				_unspents = append(_unspents, *val)
			}

			res, err = json.Marshal(responseUnspents{
				responseBase{
					CommandUnspents,
					req.Account,
				},
				_unspents,
			})
		case CommandAll:
			result, err := account.GetAddressResult(lg, lg2, config, req.Account)
			if err != nil {
				lg2.LogOnError(err, "Fails on the task")
				break
			}

			res, err = json.Marshal(responseAll{
				responseBase{
					CommandAll,
					req.Account,
				},
				*result,
			})
		default:
			panic("Unsupported task")
		}

		if err != nil {
			lg2.LogOnError(err, "Failed to output the result for the task")
		} else {
			messageChannel.Publish(
				exAccountRet,
				"",
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        res,
				},
			)
		}
		elapsedTime := time.Since(startTime)
		lg.Println("The requested task takes " + elapsedTime.String())
	}
	<-c
}

func main() {
	rs := initRedis()
	defer rs.Close()
	db := initMongoDb()
	defer db.Session.Close()
	receiver, messageChannel, rabbitMqConn := initRabbitMq()
	defer messageChannel.Close()
	defer rabbitMqConn.Close()

	running := true
	var wg sync.WaitGroup
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		running = false
		log.Println("Awaiting goroutines to gracefully shutdown...")
		wg.Wait()
		db.Session.Close()
		messageChannel.Close()
		rabbitMqConn.Close()
		rs.Close()
		os.Exit(1)
	}()

	maxTasks := 10000
	forever := make(chan bool)
	taskPool := make(chan bool, maxTasks)
	tasks := 0
	go func() {
		config := &account.Config{
			Btcd: struct {
				Endpoint string
				Username string
				Password string
			}{
				Endpoint: "https://10.23.127.245:8334",
				Username: "user",
				Password: "bitcoin321",
			},
			MongoClient: db,
			RedisClient: rs,
		}

		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range receiver {
			if !running {
				continue
			}

			log.Printf("Task received")
			if tasks >= maxTasks {
				tasks = 1
			} else {
				tasks++
			}
			wg.Add(1)
			go doTask(&wg, tasks, taskPool, d, config, messageChannel)
			if err := d.Ack(false); err != nil {
				log.Printf("Ack error occurred : %s", err)
			} else {
				log.Printf("Ack successfully")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           0,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	})
}

func initMongoDb() *bongo.Connection {
	config := &bongo.Config{
		ConnectionString: "127.0.0.1",
		Database:         "bitcoinindex",
	}
	conn, err := bongo.Connect(config)
	if err != nil {
		logger.FailOnError(err, "An error has occurred when MongoDB gets connected")
		panic(err)
	}
	return conn
}

func initRabbitMq() (<-chan amqp.Delivery, *amqp.Channel, *amqp.Connection) {
	connection, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	if err != nil {
		logger.FailOnError(err, "An error has occurred when RabbitMQ gets connected")
		panic(err)
	}

	channel, err := connection.Channel()
	if err != nil {
		logger.FailOnError(err, "Failed to create a channel over RabbitMQ")
		panic(err)
	}

	queue, err := channel.QueueDeclare(
		exAccountReq,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.FailOnError(err, "Failed to declare a queue with name 'account_req'")
		panic(err)
	}

	receiverQueue, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.FailOnError(err, "Failed to register a consumer serving queue 'account_req'")
		panic(err)
	}

	err = channel.ExchangeDeclare(
		exAccountRet,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.FailOnError(err, "Failed to register a responder channel for 'account_ret'")
		panic(err)
	}

	return receiverQueue, channel, connection
}
