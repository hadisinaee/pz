package prettierzap

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseJSON(t *testing.T) {
	type checkFunc func(ParsedJSON, bool) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	checkTruth := func(wanted bool) checkFunc {
		return func(_ ParsedJSON, ok bool) error {
			if wanted != ok {
				return fmt.Errorf("checkTruth: expected truth: %v received: %v", wanted, ok)
			}
			return nil
		}
	}

	checkLevel := func(wanted string) checkFunc {
		return func(pj ParsedJSON, _ bool) error {
			if wanted != pj.GetLevel() {
				return fmt.Errorf("checkLevel: expected level: %+v received: %+v", wanted, pj.GetLevel())
			}
			return nil
		}
	}

	checkTS := func(wanted string) checkFunc {
		return func(pj ParsedJSON, _ bool) error {
			if wanted != pj.GetTimestamp() {
				return fmt.Errorf("checkTS: expected ts: %+v received: %+v", wanted, pj.GetTimestamp())
			}
			return nil
		}
	}

	checkCaller := func(wanted string) checkFunc {
		return func(pj ParsedJSON, _ bool) error {
			if wanted != pj.GetCaller() {
				return fmt.Errorf("checkCaller: expected caller: %+v received: %+v", wanted, pj.GetCaller())
			}
			return nil
		}
	}

	checkMsg := func(wanted string) checkFunc {
		return func(pj ParsedJSON, _ bool) error {
			if wanted != pj.GetMsg() {
				return fmt.Errorf("checkMsg: expected msg: %+v received: %+v", wanted, pj.GetMsg())
			}
			return nil
		}
	}

	checkMeta := func(wantedList ...string) checkFunc {
		wanted := make(map[string]string, 0)
		for k := 0; k < len(wantedList); k = k + 2 {
			wanted[wantedList[k]] = wantedList[k+1]
		}

		return func(pj ParsedJSON, _ bool) error {
			if !reflect.DeepEqual(wanted, pj.GetMeta()) {
				return fmt.Errorf("checkMeta: expected meta: %+v(%d) received: %+v(%d) (%v)", wanted, len(wanted), pj.GetMeta(), len(pj.GetMeta()), reflect.DeepEqual(wanted, pj.GetMeta()))
			}
			return nil
		}
	}

	testScenarios := []struct {
		Name         string
		FreezeTest   bool
		GenerateJSON func() []byte
		Checks       []checkFunc
	}{
		{
			"fails due to empty json byte array",
			false,
			func() []byte {
				return make([]byte, 0)
			},
			checks(
				checkTruth(false),
			),
		},
		{
			"pass - parse one line non-json",
			false,
			func() []byte {
				return []byte(" \t this is a raw lin with \"some\": \"quoutes\" \t\t")
			},
			checks(
				checkTruth(true),
				checkLevel(`"debug"`),
				checkMsg(`this is a raw lin with "some": "quoutes"`),
			),
		},
		{
			"pass - parse one line non-json starting with '{'",
			false,
			func() []byte {
				return []byte(`  {this is a raw lin with "some": "quoutes" 	`)
			},
			checks(
				checkTruth(true),
				checkLevel(`"debug"`),
				checkMsg(`{this is a raw lin with "some": "quoutes"`),
			),
		},
		{
			"pass - parse one line non-json ending with '}'",
			false,
			func() []byte {
				return []byte(`  this is a raw lin with "some": "quoutes"}  `)
			},
			checks(
				checkTruth(true),
				checkLevel(`"debug"`),
				checkMsg(`this is a raw lin with "some": "quoutes"}`),
			),
		},
		{
			"pass - parse one line json without ts(field)",
			false,
			func() []byte {
				return []byte(`{"level":"info","ts":1522426145.1872783,"caller":"authentication/authentication.go:271","msg":"connected to the event queue","address":"localhost","port":"4222","user":"test"}`)
			},
			checks(
				checkTruth(true),
				checkLevel(`"info"`),
				checkTS("1522426145.1872783"),
				checkCaller(`"authentication/authentication.go:271"`),
				checkMsg(`"connected to the event queue"`),
				checkMeta("address", `"localhost"`, "port", `"4222"`, "user", `"test"`),
			),
		},
		{
			"pass - parse one line json",
			false,
			func() []byte {
				return []byte(`{"level":"debug","message":"reading directory for keys","foo":"bar","folder_path":"./keys/"}`)
			},
			checks(
				checkTruth(true),
				checkLevel(`"debug"`),
				checkMsg(`"reading directory for keys"`),
				checkMeta("folder_path", `"./keys/"`, "foo", `"bar"`),
			),
		},
	}

	for _, tc := range testScenarios {
		t.Run(tc.Name, func(t *testing.T) {
			j, ok := ParseJSONByteArray(tc.GenerateJSON())
			for _, check := range tc.Checks {
				if errCheck := check(j, ok); errCheck != nil {
					t.Error(errCheck)
				}
			}
		})
	}
}

func TestPrettyPrint(t *testing.T) {
	type checkFunc func(string, bool, error) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	checkError := func(wanted error) checkFunc {
		return func(_ string, _ bool, err error) error {
			if wanted != err {
				return fmt.Errorf("checkError: expected error: %v received: %v", wanted, err)
			}
			return nil
		}
	}

	checkString := func(wantedList ...string) checkFunc {
		var p ParsedJSON

		pl := parsedLog{}
		for i := 0; i < len(wantedList); i = i + 2 {
			pl[wantedList[i]] = wantedList[i+1]
		}

		p = pl
		return func(out string, emoji bool, _ error) error {
			wanted, err := GenerateOutputString(p, emoji)
			if err != nil {
				return fmt.Errorf("checkString: expected no error while generating string output received: %v", err)
			}
			if strings.Split(wanted, "\n")[0] != strings.Split(out, "\n")[0] {
				return fmt.Errorf("checkString: expected to be equal strings: %v received: %v", wanted, out)
			}

			a := strings.Split(wanted, "\n")[1:]
			b := strings.Split(out, "\n")[1:]

			if len(a) != len(b) {
				return fmt.Errorf("checkString: expected to have the same number of meta: %v received: %v", a, b)
			}
			for _, ia := range a {
				f := false
				for _, ib := range b {
					if ia == ib {
						f = true
						break
					}
				}
				if !f {
					return fmt.Errorf("checkString: expected to have the same meta field: '%v' received: %v", ia, b)
				}
			}
			return nil
		}
	}

	testScenarios := []struct {
		Name               string
		FreezeTest         bool
		GenerateParsedJSON func() ParsedJSON
		Checks             []checkFunc
	}{
		{
			"pass pretty print debug",
			false,
			func() ParsedJSON {
				p := parsedLog{
					"level":   `"debug"`,
					"ts":      "1234567.8901234",
					"caller":  `"authentication/authentication.go:271"`,
					"msg":     `"connecting to the database"`,
					"user":    `"test"`,
					"address": `"localhost"`,
					"port":    `"4222"`,
				}
				return p
			},
			checks(
				checkError(nil),
				checkString(
					"level", `"debug"`,
					"ts", "1234567.8901234",
					"caller", `"authentication/authentication.go:271"`,
					"msg", `"connecting to the database"`,
					"user", `"test"`,
					"address", `"localhost"`,
					"port", `"4222"`,
				),
			),
		},
		{
			"pass pretty print info",
			false,
			func() ParsedJSON {
				p := parsedLog{
					"level":   `"info"`,
					"ts":      "1234567.8901234",
					"caller":  `"authentication/authentication.go:271"`,
					"msg":     `"connecting to the database"`,
					"user":    `"test"`,
					"address": `"localhost"`,
					"port":    `"4222"`,
				}
				return p
			},
			checks(
				checkError(nil),
				checkString(
					"level", `"info"`,
					"ts", "1234567.8901234",
					"caller", `"authentication/authentication.go:271"`,
					"msg", `"connecting to the database"`,
					"user", `"test"`,
					"address", `"localhost"`,
					"port", `"4222"`,
				),
			),
		},
		{
			"pass pretty print error",
			false,
			func() ParsedJSON {
				p := parsedLog{
					"level":      `"error"`,
					"ts":         "1234567.8901234",
					"caller":     `"authentication/authentication.go:271"`,
					"msg":        `"connecting to the database"`,
					"stacktrace": `"gitlab.com/sinaee-hadi/users-management-service/server/authentication.Run\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/server/authentication/authentication.go:238\nmain.main\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/main.go:41\nruntime.main\n\t/usr/local/opt/go/libexec/src/runtime/proc.go:198"`,
					"user":       `"test"`,
					"address":    `"localhost"`,
					"port":       `"4222"`,
				}
				return p
			},
			checks(
				checkError(nil),
				checkString(
					"level", `"error"`,
					"ts", "1234567.8901234",
					"caller", `"authentication/authentication.go:271"`,
					"msg", `"connecting to the database"`,
					"stacktrace", `"gitlab.com/sinaee-hadi/users-management-service/server/authentication.Run\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/server/authentication/authentication.go:238\nmain.main\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/main.go:41\nruntime.main\n\t/usr/local/opt/go/libexec/src/runtime/proc.go:198"`,
					"user", `"test"`,
					"address", `"localhost"`,
					"port", `"4222"`,
				),
			),
		},
		{
			"pass pretty print dpanic",
			false,
			func() ParsedJSON {
				p := parsedLog{
					"level":   `"dpanic"`,
					"ts":      "1234567.8901234",
					"caller":  `"authentication/authentication.go:271"`,
					"msg":     `"connecting to the database"`,
					"user":    `"test"`,
					"address": `"localhost"`,
					"port":    `"4222"`,
				}
				return p
			},
			checks(
				checkError(nil),
				checkString(
					"level", `"dpanic"`,
					"ts", "1234567.8901234",
					"caller", `"authentication/authentication.go:271"`,
					"msg", `"connecting to the database"`,
					"user", `"test"`,
					"address", `"localhost"`,
					"port", `"4222"`,
				),
			),
		},
		{
			"pass pretty print panic",
			false,
			func() ParsedJSON {
				p := parsedLog{
					"level":   `"panic"`,
					"ts":      "1234567.8901234",
					"caller":  `"authentication/authentication.go:271"`,
					"msg":     `"connecting to the database"`,
					"user":    `"test"`,
					"address": `"localhost"`,
					"port":    `"4222"`,
				}
				return p
			},
			checks(
				checkError(nil),
				checkString(
					"level", `"panic"`,
					"ts", "1234567.8901234",
					"caller", `"authentication/authentication.go:271"`,
					"msg", `"connecting to the database"`,
					"user", `"test"`,
					"address", `"localhost"`,
					"port", `"4222"`,
				),
			),
		},
		{
			"pass pretty print fatal",
			false,
			func() ParsedJSON {
				p := parsedLog{
					"level":   `"fatal"`,
					"ts":      "1234567.8901234",
					"caller":  `"authentication/authentication.go:271"`,
					"msg":     `"connecting to the database"`,
					"user":    `"test"`,
					"address": `"localhost"`,
					"port":    `"4222"`,
				}
				return p
			},
			checks(
				checkError(nil),
				checkString(
					"level", `"fatal"`,
					"ts", "1234567.8901234",
					"caller", `"authentication/authentication.go:271"`,
					"msg", `"connecting to the database"`,
					"user", `"test"`,
					"address", `"localhost"`,
					"port", `"4222"`,
				),
			),
		},
	}

	for _, tc := range testScenarios {
		t.Run(tc.Name, func(t *testing.T) {
			pj := tc.GenerateParsedJSON()

			var b strings.Builder
			errPP := PrettyPrint(&b, pj, LogFilter{}, false)

			for _, check := range tc.Checks {
				if errCheck := check(b.String(), false, errPP); errCheck != nil {
					t.Error(errCheck)
				}
			}
			var bn strings.Builder
			errPP = PrettyPrint(&bn, pj, LogFilter{}, true)

			for _, check := range tc.Checks {
				if errCheck := check(bn.String(), true, errPP); errCheck != nil {
					t.Error(errCheck)
				}
			}
		})
	}
}

func TestFilterJSON(t *testing.T) {
	type checkFunc func([]bool) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	checkFiltered := func(wanted []bool) checkFunc {
		return func(given []bool) error {
			if len(wanted) != len(given) {
				return fmt.Errorf("checkFiltered: expected lengths [%v] received: [%v]", len(wanted), len(given))
			}
			for i := 0; i < len(wanted); i++ {
				if wanted[i] != given[i] {
					return fmt.Errorf("checkFiltered: expected filtered @(%d) [%v] received: [%v]", i, wanted, given)
				}
			}
			return nil
		}
	}

	testScenarios := []struct {
		Name               string
		FreezeTest         bool
		GenerateParsedJSON func() ([]ParsedJSON, LogFilter)
		Checks             []checkFunc
	}{
		{
			"pass pretty print - filters just debug messages",
			false,
			func() ([]ParsedJSON, LogFilter) {
				p := []ParsedJSON{
					parsedLog{
						"level":   `"debug"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
					},
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
					},
				}

				f := LogFilter{
					Level: "debug",
				}
				return p, f
			},
			checks(checkFiltered([]bool{true, false})),
		},
		{
			"pass pretty print - filters caller='authentication' ",
			false,
			func() ([]ParsedJSON, LogFilter) {
				p := []ParsedJSON{
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
					},
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"hello/run.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
					},
				}

				f := LogFilter{
					Caller: "authentication",
				}
				return p, f
			},
			checks(checkFiltered([]bool{true, false}))},
		{
			"pass pretty print - filters just ts >= 1234567.8901234",
			false,
			func() ([]ParsedJSON, LogFilter) {
				p := []ParsedJSON{
					parsedLog{
						"level":      `"error"`,
						"ts":         "1234567.8901220",
						"caller":     `"authentication/authentication.go:271"`,
						"msg":        `"connecting to the database"`,
						"stacktrace": `"gitlab.com/sinaee-hadi/users-management-service/server/authentication.Run\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/server/authentication/authentication.go:238\nmain.main\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/main.go:41\nruntime.main\n\t/usr/local/opt/go/libexec/src/runtime/proc.go:198"`,
						"user":       `"test"`,
						"address":    `"localhost"`,
						"port":       `"4222"`,
					},
					parsedLog{
						"level":      `"error"`,
						"ts":         "1234567.8901234",
						"caller":     `"authentication/authentication.go:271"`,
						"msg":        `"connecting to the database"`,
						"stacktrace": `"gitlab.com/sinaee-hadi/users-management-service/server/authentication.Run\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/server/authentication/authentication.go:238\nmain.main\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/main.go:41\nruntime.main\n\t/usr/local/opt/go/libexec/src/runtime/proc.go:198"`,
						"user":       `"test"`,
						"address":    `"localhost"`,
						"port":       `"4222"`,
					},
					parsedLog{
						"level":      `"error"`,
						"ts":         "1234567.8901234",
						"caller":     `"authentication/authentication.go:271"`,
						"msg":        `"connecting to the database"`,
						"stacktrace": `"gitlab.com/sinaee-hadi/users-management-service/server/authentication.Run\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/server/authentication/authentication.go:238\nmain.main\n\t/Users/hadi/Programmings/golang/src/gitlab.com/sinaee-hadi/users-management-service/main.go:41\nruntime.main\n\t/usr/local/opt/go/libexec/src/runtime/proc.go:198"`,
						"user":       `"test"`,
						"address":    `"localhost"`,
						"port":       `"4222"`,
					},
				}

				f := LogFilter{
					Timestamp: "1234567.8901234",
				}

				return p, f
			},
			checks(checkFiltered([]bool{false, true, true})),
		},
		{
			"pass pretty print - filters just token=1234 messages",
			false,
			func() ([]ParsedJSON, LogFilter) {
				p := []ParsedJSON{
					parsedLog{
						"level":   `"debug"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
						"token":   "1234",
					},
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
					},
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
						"token":   "4321",
					},
				}

				m := make(map[string]*string, 1)
				t := "1234"
				m["token"] = &t
				f := LogFilter{
					Meta: m,
				}
				return p, f
			},
			checks(checkFiltered([]bool{true, false, false})),
		},
		{
			"pass pretty print - filters just token=1234 & auth=\"badef123==\" messages",
			false,
			func() ([]ParsedJSON, LogFilter) {
				p := []ParsedJSON{
					parsedLog{
						"level":   `"debug"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
						"token":   "1234",
						"auth":    `"badef123=="`,
					},
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
						"token":   "1234",
					},
					parsedLog{
						"level":   `"info"`,
						"ts":      "1234567.8901234",
						"caller":  `"authentication/authentication.go:271"`,
						"msg":     `"connecting to the database"`,
						"user":    `"test"`,
						"address": `"localhost"`,
						"port":    `"4222"`,
						"token":   "4321",
					},
				}

				m := make(map[string]*string, 2)
				t := "1234"
				a := `"badef123=="`
				m["token"] = &t
				m["auth"] = &a

				f := LogFilter{
					Meta: m,
				}
				return p, f
			},
			checks(checkFiltered([]bool{true, false, false})),
		},
	}

	for _, tc := range testScenarios {
		t.Run(tc.Name, func(t *testing.T) {
			pjs, filter := tc.GenerateParsedJSON()

			results := make([]bool, 0)
			for _, pj := range pjs {
				filtered := filterJSON(pj, filter)
				results = append(results, filtered)
			}

			for _, check := range tc.Checks {
				if errCheck := check(results); errCheck != nil {
					t.Error(errCheck)
				}
			}
		})
	}
}
