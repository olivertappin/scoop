# Scoop

## Introduction

Scoop was created to solve the problem of shovelling messages from one queue to another, with the option of choosing how
many messages you want to send.

### Why?

This might be necessary for a number of reasons. The more likely scenario is a disaster situation, where you have a
multiple node RabbitMQ cluster where one or more node has failed. You have messages on a queue which are not HA, and
they are building up.

After many '/' vhost errors, and your cluster looking more unhealthy by the minute, you may want
to rearrange where some of your messages are being distributed to. With the shovel plugin, you have no choice but to
shovel all your messages. Scoop solves this by giving you the option of how many you want to move around at a time. 

### Tell me more!

It's written in Golang and is distributed as a binary package for all supported platforms and architectures.

## Usage

To use Scoop via the binary, run:

```bash
bin/scoop -from <from-queue-name> -to <to-queue-name> -count <number-of-messages-to-move>
```

To run Scoop directly via the source code, run:

```bash
go run src/scoop.go -from <from-queue-name> -to <to-queue-name> -count <number-of-messages-to-move>
```

## Arguments
- `username` - The username used to connect (default: `guest`)
- `password` - The password used to connect (default: `guest`)
- `hostname` - The hostname to connect to (default: `localhost`)
- `port` - The port to connect to (default: `5672`)
- `from` - The queue name to consume messages from (required)
- `to` - The queue name to deliver messages to (required)
- `durable` - Define the queue decleration to be durable (default: `false`)
- `exchange` - The exchange name to deliver messages through (default: `""`)
- `count` - The _maximum_ number of messages to move (default: `1`)
- `v` - Turn on verbose mode (default: `false`)
- `vv` - Turn on very verbose mode (default: `false`)
- `vvv` - Turn on extremely verbose mode (default: `false`)

## Notes

Scoop attempts to achieve shovelling with the important option of being able to control how many messages are shovelled into another queue.

Another way of achieving this (suggested by [@EagleEyeJohn](https://github.com/EagleEyeJohn)) would be to create a new queue with the following arguments:
```
x-overflow: reject-publish
x-max-length: <count>
```

You can then, optionally, create another queue. You would then move your messages out of your bad queue, into this new queue.

Finally, moving the messages into the first queue you created (with the arguments shown above), you can shovel messages into that queue, and they will stop when they get to the `x-max-length` limit.

## Contributing

Please open a GitHub issue for discussion before opening a pull-request.

Once created, create your branch in the following format: `{feature|hotfix}/{github-issue-number}` and open the relevant
pull-request into the master branch for review.


Before pushing your changes, ensure you have compiled the binary file using:

```bash
go build -o bin/scoop src/scoop.go
```

All contributions, no matter how large or small, are welcome.
