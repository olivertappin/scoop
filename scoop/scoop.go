package main

import (
    "log"
    "flag"
    "os"
    "fmt"
    "os/signal"
    "strings"
    "strconv"
    "github.com/streadway/amqp"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var arguments arrayFlags
var fromArguments arrayFlags
var toArguments arrayFlags
var terminate bool = false

var (
    username         = flag.String("username", "guest", "Username")
    password         = flag.String("password", "guest", "Password")
    hostname         = flag.String("hostname", "localhost", "Hostname")
    port             = flag.String("port", "5672", "Port")
    fromQueueName    = flag.String("from", "", "The queue name to consume messages from")
    toQueueName      = flag.String("to", "", "The queue name to deliver messages to")
    fromDurable      = flag.Bool("from-durable", false, "Define the from queue deceleration to be durable")
    toDurable        = flag.Bool("to-durable", false, "Define the to queue deceleration to be durable")
    exchange         = flag.String("exchange", "", "The exchange name to deliver messages through")
    messageCount     = flag.Int("count", 1, "The number of messages to move between queues")
    verbose          = flag.Bool("v", false, "Turn on verbose mode")
    veryVerbose      = flag.Bool("vv", false, "Turn on very verbose mode")
    extremelyVerbose = flag.Bool("vvv", false, "Turn on extremely verbose mode")
)

func init() {
    flag.Var(&arguments, "arg", "Argument(s) to pass to the queue decelerations")
    flag.Var(&fromArguments, "from-arg", "Argument(s) to pass the queue deceleration which consumes messages")
    flag.Var(&toArguments, "to-arg", "Argument(s) to pass to the queue deceleration which delivers messages")
    flag.Parse()
}

func main() {
    if *extremelyVerbose {
        log.Printf("Extremely verbose mode enabled")
    } else if *veryVerbose {
        log.Printf("Very verbose mode enabled")
    } else if *verbose {
        log.Printf("Verbose mode enabled")
    }

    // Set the verbose modes accordingly
    if *extremelyVerbose {
        *veryVerbose = true
        *verbose = true
    } else if *veryVerbose {
       *verbose = true
    }

    if *fromQueueName == "" {
        log.Printf("The from argument must be defined")
        os.Exit(2)
    }

    if *toQueueName == "" {
        log.Printf("The to argument must be defined")
        os.Exit(2)
    }

    if *fromQueueName == *toQueueName {
        log.Printf("The from queue name matches the to queue name")
        os.Exit(2)
    }

    fromArgs := make(amqp.Table)
    toArgs := make(amqp.Table)

    for _, argument := range arguments {
        fromArgs = mapQueueArguments(fromArgs, argument)
        toArgs = mapQueueArguments(toArgs, argument)
    }
    for _, fromArgument := range fromArguments {
        fromArgs = mapQueueArguments(fromArgs, fromArgument)
    }
    for _, toArgument := range toArguments {
        toArgs = mapQueueArguments(toArgs, toArgument)
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        for sig := range c {
            terminate = true
            log.Printf("Finishing up ... (%v detected)", sig)
        }
    }()

    // It's advisable to use separate connections for Channel.Publish and Channel.Consume so not to have TCP pushback
    // on publishing affect the ability to consume messages: https://godoc.org/github.com/streadway/amqp#Channel.Consume

    consumerConnection, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", *username, *password, *hostname, *port))
    failOnError(err, "Failed to create the consumer connection to RabbitMQ")
    defer consumerConnection.Close()

    publisherConnection, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", *username, *password, *hostname, *port))
    failOnError(err, "Failed to create the publisher connection to RabbitMQ")
    defer publisherConnection.Close()

    consumerChannel, err := consumerConnection.Channel()
    failOnError(err, "Failed to open the consumer channel")
    defer consumerChannel.Close()

    publisherChannel, err := publisherConnection.Channel()
    failOnError(err, "Failed to open the publisher channel")
    defer publisherChannel.Close()

    if *verbose {
        log.Printf("Moving %d messages from queue %s to %s", *messageCount, *fromQueueName, *toQueueName)
    }

    // Check if the queue exists, otherwise, fail
    // To do this, add the additional argument of passive to true. See:
    // https://github.com/streadway/amqp/blob/master/channel.go#L758
    // so if the queue does exist, the command fails (but we need the latest code for that)

    if *extremelyVerbose && len(fromArgs) != 0 {
        log.Printf("Declaring from queue with the following arguments:")
        for key, argument := range fromArgs {
            log.Println("-", key, ":", argument);
        }
    }

    fromQueue, err := consumerChannel.QueueDeclare(
        *fromQueueName, // name
        *fromDurable,   // durable
        false,          // delete when unused
        false,          // exclusive
        false,          // no-wait
        fromArgs,       // arguments
    )
    failOnError(err, "Failed to declare a queue")

    if *veryVerbose {
        log.Printf("There are %d messages in queue %s", fromQueue.Messages, fromQueue.Name)
    }

    if *extremelyVerbose && len(toArgs) != 0 {
        log.Printf("Declaring to queue with the following arguments:")
        for key, argument := range toArgs {
            log.Println("-", key, ":", argument);
        }
    }

    toQueue, err := publisherChannel.QueueDeclare(
        *toQueueName, // name
        *toDurable,   // durable
        false,        // delete when unused
        false,        // exclusive
        false,        // no-wait
        toArgs,       // arguments
    )
    failOnError(err, "Failed to declare a queue")

    if *veryVerbose {
        log.Printf("There are %d messages in queue %s", toQueue.Messages, toQueue.Name)
    }

    messages, err := consumerChannel.Consume(
        fromQueue.Name, // queue
        "",             // consumer
        false,          // auto-ack (it's very important this stays false)
        false,          // exclusive
        false,          // no-local
        false,          // no-wait
        nil,            // args
    )
    failOnError(err, "Failed to register the scoop consumer")

    log.Printf("Running scoop consumer... (press Ctl-C to cancel)")

    confirms := publisherChannel.NotifyPublish(make(chan amqp.Confirmation, 1))

    if err := publisherChannel.Confirm(false); err != nil {
        log.Fatalf("Unable to put publisher channel into confirm mode: %s", err)
    }

    i := 1

    for {
        message, ok := <-messages
        if !ok {
            log.Printf("The consumer channel was unexpectedly closed")
            break
        }

        if terminate {
            break
        }

        if i > *messageCount {
            break
        }

        err = publisherChannel.Publish(
            *exchange,    // exchange
            toQueue.Name, // routing key
            false,        // mandatory
            false,        // immediate
            amqp.Publishing{
                ContentType:     message.ContentType,
                ContentEncoding: message.ContentEncoding,
                DeliveryMode:    message.DeliveryMode,
                Priority:        message.Priority,
                CorrelationId:   message.CorrelationId,
                ReplyTo:         message.ReplyTo,
                Expiration:      message.Expiration,
                MessageId:       message.MessageId,
                Timestamp:       message.Timestamp,
                Type:            message.Type,
                UserId:          message.UserId,
                AppId:           message.AppId,
                Headers:         message.Headers,
                Body:            message.Body,
            })

        if err != nil {
            message.Nack(false, false)
            log.Printf("Failed to deliver message")
            break
        }

        if confirmed := <-confirms; confirmed.Ack {
            message.Ack(false)
            if *extremelyVerbose {
                log.Printf("Successfully delivered message (%d/%d)", i, *messageCount)
            }
            i++
            continue
        }

        message.Nack(false, false)
        if *extremelyVerbose {
            log.Printf("Failed to deliver message ... refused to acknowledge delivery")
        }
    }

    if *verbose {
        log.Printf("Complete")
    }
}

func mapQueueArguments(arguments amqp.Table, argument string) amqp.Table {
    parts := strings.Split(argument, ":")
    key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

    // Keep values as strings where the keys expect it
    if key == "x-overflow" || key == "x-queue-mode" || key == "x-queue-master-locator" || key == "x-dead-letter-exchange" || key == "x-dead-letter-routing-key" {
        arguments[key] = value
        return arguments
    }

    // Cast values to integers where the keys expect it
    i, err := strconv.Atoi(value)
    if err != nil {
        log.Fatalf("Argument with key \"%s\" does not have a valid integer value. Received: \"%s\"", key, value)
    }
    arguments[key] = i
    return arguments
}

func failOnError(err error, message string) {
    if err != nil {
        log.Fatalf("%s: %s", message, err)
    }
}
