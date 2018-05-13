package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

// CLIConfig represents cli configuration
type CLIConfig struct {
	Name    string // name of this cli
	Usage   string // usage of this cli
	Version string // app version
}

var app *cli.App

// InitCLI initialize the cli with the given config object
func InitCLI(cfg CLIConfig, level, timestamp, caller *string, keyValuePairs map[string]*string, emoji *bool) error {
	app = cli.NewApp()
	app.Name = cfg.Name
	app.Usage = cfg.Usage
	app.Version = cfg.Version

	var tempKVs string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "l, level",
			Usage:       "just logs with log level of `log_level`",
			Destination: level,
		},
		cli.StringFlag{
			Name:        "t, timestamp",
			Usage:       "just logs after the `timestamp`(>=). it is possible to use the following keywords with `timestamp`:\n\t\t\tnow: to show all logs from the current time\n\t\t\ttoday: to show all logs of the tody(start from 00:00)",
			Destination: timestamp,
		},
		cli.StringFlag{
			Name:        "c, caller",
			Usage:       "just logs that its caller field contains `caller_name`",
			Destination: caller,
		},
		cli.StringFlag{
			Name:        "k, keyvalue",
			Usage:       "just logs that have specific pairs of `key_1=value_1`",
			Destination: &tempKVs,
		},
		cli.BoolFlag{
			Name:        "e, emoji",
			Usage:       "add some funny emoji to output",
			Destination: emoji,
		},
	}

	app.Action = func(c *cli.Context) error {
		if strings.Compare(*timestamp, "now") == 0 {
			*timestamp = fmt.Sprintf("%v", time.Now().Unix())
		} else if strings.Compare(*timestamp, "today") == 0 {
			y, m, d := time.Now().Date()
			l, _ := time.LoadLocation("Local")
			t := time.Date(y, m, d, 0, 0, 0, 0, l)
			*timestamp = fmt.Sprintf("%v", t.Truncate(24*time.Hour).Unix())
		}

		pairs := strings.Split(tempKVs, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}

			_, errParse := strconv.ParseFloat(kv[1], 64)
			if errParse != nil {
				kv[1] = fmt.Sprintf("\"%s\"", kv[1])
				keyValuePairs[kv[0]] = &kv[1]
			} else {
				keyValuePairs[kv[0]] = &kv[1]
			}

		}

		title := fmt.Sprintf("\n[PRITTIER ZAP] Level: '%v' Timestamp: '%v' Caller: '%v' Emoji: '%v'", *level, *timestamp, *caller, *emoji)
		if len(keyValuePairs) > 0 {
			title += " Key-Value:"
		}
		for k, v := range keyValuePairs {
			title += fmt.Sprintf(" %s:%s", k, *v)
		}
		title += "\n"
		fmt.Println(title)
		return nil
	}
	return nil
}

// Run runs the cli application with given os arguments.
func Run(osArgs []string) error {
	errRun := app.Run(osArgs)
	if errRun != nil {
		return errRun
	}
	return nil
}
