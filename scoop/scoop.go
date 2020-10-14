package main

import (
    "log"
    "flag"
    "os"
    "fmt"
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

var (
    username         = flag.String("username", "guest", "Username")
    password         = flag.String("password", "guest", "Password")
    hostname         = flag.String("hostname", "localhost", "Hostname")
    port             = flag.String("port", "5672", "Port")
    fromQueueName    = flag.String("from", "", "The queue name to consume messages from")
    toQueueName      = flag.String("to", "", "The queue name to deliver messages to")
    contentType      = flag.string("content-type", "text/plain", "The content_type to use when publishing messages")
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

    if *contentType == "" {
        log.Printf("The content type must not be empty")
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

    conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", *username, *password, *hostname, *port))
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    if *verbose {
        log.Printf("Moving %d messages from queue %s to %s", *messageCount, *fromQueueName, *toQueueName)
    }

    // Check if the queue exists, otherwise, fail
    // To do this, add the additional argument of passive to true: https://github.com/streadway/amqp/blob/master/channel.go#L758
    // so if the queue does exist, the command fails (but we need the latest code for that)

    if *extremelyVerbose && len(fromArgs) != 0 {
        log.Printf("Declaring from queue with the following arguments:")
        for key, argument := range fromArgs {
            log.Println("-", key, ":", argument);
        }
    }

    fromQueue, err := ch.QueueDeclare(
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

    toQueue, err := ch.QueueDeclare(
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

    msgs, err := ch.Consume(
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

    i := 1

    for d := range msgs {
        if i > *messageCount {
            if *extremelyVerbose {
                log.Printf("Complete")
            }
            break
        }

        err = ch.Publish(
            *exchange,    // exchange
            toQueue.Name, // routing key
            false,        // mandatory
            false,        // immediate
            amqp.Publishing{
                ContentType: *contentType,
                Body:        []byte(d.Body),
            })

        failOnError(err, "Failed to deliver message")

        if *extremelyVerbose {
            log.Printf("Successfully delivered message (%d/%d)", i, *messageCount)
        }

        d.Ack(true)
        i++
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

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}
