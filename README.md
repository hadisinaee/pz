# Prettier Zap

[![Build Status](https://travis-ci.org/hadisinaee/pz.svg?branch=master)](https://travis-ci.org/hadisinaee/pz)
[![Coverage Status](https://coveralls.io/repos/github/hadisinaee/pz/badge.svg?branch=master)](https://coveralls.io/github/hadisinaee/pz?branch=master)

> in simple terms, it pretty prints zap logs and let you query them!

![Prettier Logo](./logo.png)

If you are using [Zap Logger](https://github.com/uber-go/zap), you can use `Prettier Zap` cli for:

* making the raw JSON outputs of Zap more readable for developer
* making query base on `timestamp`, `level`, `caller` or a specific `(key, value)` pair

> all non-zap logs are treated as `debug` logs with the `caller` field set to the `user-code` string

[![asciicast](https://asciinema.org/a/181271.png)](https://asciinema.org/a/181271)

## How To Install?

```sh
go get -u github.com/hadisinaee/prettierzap
# or
go dep ensure --add github.com/hadisinaee/prettierzap
```

## How To Use It?

#### Quick Start

If your program logs into `stdout`, you can simply pipe it to the `pz` command:

```sh
go run main.go | pz
```

or if it's written to a file:

```sh
tail -f .log | pz
```

You can add a `-e` to all the following commands to make the logs look funny with emojis:

```sh
go run main.go | pz -e
```

#### Query Base On A Level

You can make a query for a specific log level by adding `-l log_level`:

```sh
go run main.go | pz -l info
```

#### Query Base On A Timestamp

You can make a query for all logs of today by adding a `-t today`:

```sh
go run main.go | pz -t today
```

You can make a query for all logs which are generated from now on by adding a `-t now`:

```sh
go run main.go | pz -t now
```

You can make a query for all logs which are generated after a specific timestamp by adding a `-t 123456789`:

```sh
go run main.go | pz -t 123456789
```

#### Query Base On A Caller

You can make a query for a specific caller function by adding a `-c caller`:

```sh
go run main.go| pz -c authentication
```

#### Query Base On a Key-Value Pair

If your logs have key value pairs and you are looking for logs with a specific key-value pair, you can add a `k key_1=value_1`:

```sh
go run main.go| pz -k req_id="abcdef1234="
# or
go run main.go| pz -k req_id="abcdef1234=",uid=10230212
```

## CLI Help

```
>> pz -h
NAME:
   Prettier Zap - make zap logs more beautiful and queryable

USAGE:
   pz [global options] command [command options] [arguments...]

VERSION:
   0.9.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -l log_level, --level log_level             just logs with log level of log_level
   -t timestamp, --timestamp timestamp         just logs after the timestamp(>=). it is possible to use the following keywords with `timestamp`:
                                                   now: to show all logs from the current time
                                                   today: to show all logs of the tody(start from 00:00)
   -c caller_name, --caller caller_name        just logs that its caller field contains caller_name
   -k key_1=value_1, --keyvalue key_1=value_1  just logs that have specific pairs of key_1=value_1
   -e, --emoji                                 add some funny emoji to output
   --help, -h                                  show help
   --version, -v                               print the version
```
