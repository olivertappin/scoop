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

## Install

To download Scoop via the command line, run:

```bash
curl -LOs https://raw.githubusercontent.com/olivertappin/scoop/2.0.0/bin/scoop
sudo chmod 755 scoop
```

Optionally, move this to your `/usr/bin` directory to allow this to be available globally:

```bash
sudo mv scoop /usr/bin/scoop
```

## Usage

To use Scoop via the binary, run:

```bash
./scoop \
  -from <from-queue-name> \
  -to <to-queue-name> \
  -count <number-of-messages-to-move> \
  -vvv
```

To run Scoop directly via the source code, run:

```bash
cd path/to/scoop
go run scoop/scoop.go \
  -from <from-queue-name> \
  -to <to-queue-name> \
  -count <number-of-messages-to-move> \
  -vvv
```

## Arguments
- `username` - The username used to connect (default: `guest`)
- `password` - The password used to connect (default: `guest`)
- `hostname` - The hostname to connect to (default: `localhost`)
- `port` - The port to connect to (default: `5672`)
- `from` - The queue name to consume messages from (required)
- `to` - The queue name to deliver messages to (required)
- `from-durable` - Define the from queue deceleration to be durable (default: `false`)
- `to-durable` - Define the to queue deceleration to be durable (default: `false`)
- `arg` - Argument(s) to pass to the from queue deceleration (default: `""`)
- `from-arg` - Argument(s) to pass the queue deceleration which consumes messages; overrides values set by `arg` (default: `""`)
- `to-arg` - Argument(s) to pass to the queue deceleration which delivers messages; overrides values set by `arg` (default: `""`)
- `exchange` - The exchange name to deliver messages through (default: `""`)
- `count` - The _maximum_ number of messages to move (default: `1`)
- `v` - Turn on verbose mode (default: `false`)
- `vv` - Turn on very verbose mode (default: `false`)
- `vvv` - Turn on extremely verbose mode (default: `false`)

### Queue deceleration arguments

#### `x-message-ttl` (number)<br>
How long a message published to a queue can live before it is discarded (milliseconds).
<br>
https://www.rabbitmq.com/ttl.html

#### `x-expires` (number)<br>
How long a queue can be unused for before it is automatically deleted (milliseconds).<br>
https://www.rabbitmq.com/ttl.html#queue-ttl

#### `x-max-length` (number)<br>
How many (ready) messages a queue can contain before it starts to drop them from its head.<br>
https://www.rabbitmq.com/maxlength.html

#### `x-max-length-bytes` (number)<br>
Total body size for ready messages a queue can contain before it starts to drop them from its head.<br>
https://www.rabbitmq.com/maxlength.html#definition-using-x-args

#### `x-overflow` (string)<br>
Sets the queue overflow behaviour. This determines what happens to messages when the maximum length of a queue is reached. Valid values are `drop-head` or `reject-publish`.<br>
https://www.rabbitmq.com/maxlength.html#overflow-behaviour

#### `x-dead-letter-exchange` (string)<br>
Optional name of an exchange to which messages will be republished if they are rejected or expire.<br>
https://www.rabbitmq.com/dlx.html

#### `x-dead-letter-routing-key` (string)<br>
Optional replacement routing key to use when a message is dead-lettered. If this is not set, the message's original routing key will be used.<br>
https://www.rabbitmq.com/dlx.html#routing

#### `x-max-priority` (number)<br>
Maximum number of priority levels for the queue to support; if not set, the queue will not support message priorities.<br>
https://www.rabbitmq.com/priority.html

#### `x-queue-mode` (string)<br>
Set the queue into lazy mode, keeping as many messages as possible on disk to reduce RAM usage; if not set, the queue will keep an in-memory cache to deliver messages as fast as possible.<br>
https://www.rabbitmq.com/lazy-queues.html

#### `x-queue-master-locator` (string)<br>
Set the queue into master location mode, determining the rule by which the queue master is located when declared on a cluster of nodes.<br>
https://www.rabbitmq.com/ha.html#queue-master-location

Please note, this is not an exhaustive list. For up-to-date documentation for your correct version of RabbitMQ, please consult the manual.<br>

### Example of `-arg` usage

The above values can be passed into the `arg` argument. An example of how this could be used is shown below:

```bash
./scoop \
  -from <from-queue-name> \
  -to <to-queue-name> \
  -count <number-of-messages-to-move> \
  -arg "x-message-ttl: 60000" \
  -arg "x-queue-mode: lazy" \
  -vvv
```

You can use the `-from-arg` and `-to-arg` in exactly the same way, but these definitions will be separated between the to and from queues.

Any values defined by the `-from-arg` and `-to-arg` will override values set by the `-arg` parameter.

## Notes

Scoop attempts to achieve shovelling with the important option of being able to control how many messages are shovelled into another queue.

Another way of achieving this (suggested by [@EagleEyeJohn](https://github.com/EagleEyeJohn)) would be to create a new queue with the following arguments:
```
x-overflow: reject-publish
x-max-length: <count>
```

You can then, optionally, create another queue. You would then move your messages out of your bad queue, into this new queue.

Finally, moving the messages into the first queue you created (with the arguments shown above), you can shovel messages into that queue, and they will stop when they get to the `x-max-length` limit.

## Uninstall

To remove scoop from your system globally, simply remove the binary from `/usr/bin`:

```bash
sudo rm -f /usr/bin/scoop
```

## Contributing

Please open a GitHub issue for discussion before opening a pull-request.

Once created, create your branch in the following format: `{feature|hotfix}/{github-issue-number}` and open the relevant
pull-request into the master branch for review.


Before pushing your changes, ensure you have compiled the binary file using:

```bash
go build -o bin/scoop scoop/scoop.go
```

All contributions, no matter how large or small, are welcome.
