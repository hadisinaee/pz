package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hadisinaee/pz/cmd"
	"github.com/hadisinaee/pz/prettierzap"
)

func main() {
	var (
		level         string
		timestamp     string
		caller        string
		keyValuePairs map[string]*string
		emoji         bool
	)

	keyValuePairs = make(map[string]*string, 0)
	cmd.InitCLI(cmd.CLIConfig{
		Name:    "Prettier Zap",
		Usage:   "make zap logs more beautiful and queryable",
		Version: "0.9.2",
	}, &level, &timestamp, &caller, keyValuePairs, &emoji)
	cmd.Run(os.Args)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		l := scanner.Bytes()
		if len(l) < 1 {
			continue
		}

		pj, ok := prettierzap.ParseJSONByteArray(l)
		if !ok {
			fmt.Printf("[(PZ) cannot parse line]: %v", string(l))
		}

		prettierzap.PrettyPrint(os.Stdout, pj, prettierzap.LogFilter{
			Level:     level,
			Timestamp: timestamp,
			Caller:    caller,
			Meta:      keyValuePairs,
		}, emoji)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[(PZ) Scanner Error]= %+v", err)
	}
}
