package prettierzap

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ParsedJSON represents a parsed zap json object.
type ParsedJSON interface {
	GetLevel() string
	GetTimestamp() string
	GetCaller() string
	GetMsg() string
	GetMeta() map[string]string
}

// LogFilter represents a filter that is used for filter logs based on specific fields
type LogFilter struct {
	Level     string
	Timestamp string
	Caller    string
	Meta      map[string]*string
}

var (
	fgCyan              = color.New(color.FgCyan).SprintfFunc()
	bgYellowfgBlackBold = color.New(color.BgYellow, color.FgBlack, color.Bold).SprintfFunc()
	bgYellowfgBlack     = color.New(color.BgYellow, color.FgBlack).SprintfFunc()
	fgYellow            = color.New(color.FgYellow).SprintfFunc()
	fgRed               = color.New(color.FgRed).SprintfFunc()
)

const (
	debugLevel   = `debug`
	warningLevel = `warn`
	infoLevel    = `info`
	errorLevel   = `error`
	dPanicLevel  = `dpanic`
	panicLevel   = `panic`
	fatalLevel   = `fatal`
)

// findKey search the given data array for a key which ends with a `"`.
func findKey(data []byte) (string, int) {
	k := make([]byte, 0)
	for i, b := range data {
		if b == '"' {
			return string(k), i
		}
		k = append(k, b)
	}
	return "", 0
}

// findValue search the given data array for a value which ends with a `,` or a `}`.
func findValue(data []byte) (string, int) {
	v := make([]byte, 0)
	for i, b := range data {
		if b == ',' || b == '}' {
			return string(v), i
		}
		v = append(v, b)
	}
	return "", 0
}

// filterJSON filters jsons based on the given log filter object.
func filterJSON(pj ParsedJSON, f LogFilter) bool {
	pp := true

	if f.Level != "" && strings.Replace(pj.GetLevel(), "\"", "", -1) != f.Level {
		pp = false
	} else if f.Caller != "" && !strings.Contains(pj.GetCaller(), f.Caller) {
		pp = false
	} else if f.Timestamp != "" && pj.GetTimestamp() < f.Timestamp {
		pp = false
	}

	pMeta := pj.GetMeta()
	for kf, vf := range f.Meta {
		if vp, ok := pMeta[kf]; !ok {
			pp = false
			break
		} else if vp != *vf {
			pp = false
			break
		}
	}
	return pp
}

// GenerateOutputString generates the formatted output string for the given parsed JSON.
func GenerateOutputString(pj ParsedJSON, emoji bool) (string, error) {
	var (
		l = pj.GetLevel()
		s = ""
		e error
	)

	if pj.GetTimestamp() != "" {
		tss := strings.Split(pj.GetTimestamp(), ".")[0]

		ts, errParse := strconv.ParseInt(tss, 10, 64)
		if errParse != nil {
			return "", errParse
		}

		if emoji {
			s = fmt.Sprintf("%s %s ", "\U000023F0", bgYellowfgBlack("%-20s", time.Unix(ts, 0).Format("02/01/2006 15:04:05")))
		} else {
			s = bgYellowfgBlack("%-20s| ", time.Unix(ts, 0).Format("02/01/2006 15:04:05"))
		}
	}

	if l != "" {
		var emojiChar string

		if emoji {
			switch l {
			case `"info"`:
				// pagger
				emojiChar = "\U0001F4DF"
			case `"warn"`:
				// warning
				emojiChar = "\U000026A0 "
			case `"error"`:
				// alarm
				emojiChar = "\U0001F6A8"
			case `"panic"`, `"dpanic"`:
				// pile of poo
				emojiChar = "\U0001F4A9"
			case `"fatal"`:
				// skull
				emojiChar = "\U00002620 "
			case `"debug"`:
				// high voltage
				// emojiChar = "\U000026A1"
				// eyes
				emojiChar = "\U0001F440"
			}
			s = s + fmt.Sprintf("%s %s", emojiChar, bgYellowfgBlackBold(" %-8s", strings.Replace(strings.ToUpper(l), "\"", "", -1)))
		} else {
			s = s + bgYellowfgBlackBold(" %-8s", strings.Replace(strings.ToUpper(l), "\"", "", -1))
		}
	}

	if pj.GetCaller() != "" {
		if emoji {
			s = s + fmt.Sprintf(" %s%s", "\U0001F5E3", fgCyan(" [%s]", pj.GetCaller()))
		} else {
			s = s + fgCyan(" @[%s]", pj.GetCaller())
		}

	}

	l = strings.Replace(l, "\"", "", -1)
	if l == debugLevel || l == warningLevel {
		s = s + " " + fgYellow(pj.GetMsg())
	} else if l == fatalLevel || l == errorLevel || l == dPanicLevel || l == panicLevel {
		s = s + " " + fgRed(pj.GetMsg())
	} else {
		s = s + " " + pj.GetMsg()
	}

	s += "\n"

	if len(pj.GetMeta()) > 0 {
		var m bytes.Buffer
		var r string
		meta := pj.GetMeta()

		st, ok := meta["stacktrace"]
		if ok {
			st = strings.Replace(st, "\\n\\t", "\U0000000A\U00000009\U00000009> ", -1)
			st = strings.Replace(st, "\\n", "\U0000000A\U00000009\U00000009 ", -1)
			st = st[1 : len(st)-1]

			r = fmt.Sprintf("\t%v: \n\t\t%s\n", fgRed("%q", "stacktrace"), fgRed("> %s", st))
			m.WriteString(r)
		}
		for key := range meta {
			if key != "stacktrace" {
				r = fmt.Sprintf("   %v: %s\n", fgCyan("%q", key), meta[key])
				m.WriteString(r)
			}
		}
		s = fmt.Sprintf("%s%s\n", s, m.String())
	}
	return s, e
}

// ParseJSONByteArray parses the given byte array and creates a ParsedJSON object.
func ParseJSONByteArray(jsonByte []byte) (ParsedJSON, bool) {
	if len(jsonByte) == 0 {
		return nil, false
	}

	var (
		pl         = parsedLog{}
		offset     = 0
		nextOffset = 0
		byteLength = len(jsonByte)
		hasStart   = false
		hasEnd     = false
		foundKey   = ""
	)

	// check for a valid JSON string(encapsulated between {})
	// search in the byte array from both ends looking for `{` and `}`
	// if it has both `{` and `}`, it is a JSON byte array
	for i := 0; i < byteLength; i++ {
		if !hasStart {
			switch jsonByte[i] {
			case ' ', '\t':
				continue
			case '{':
				hasStart = true
				offset = i
			default:
				hasStart = false
			}
		}

		if !hasEnd {
			switch jsonByte[byteLength-i-1] {
			case ' ', '\t':
				continue
			case '}':
				hasEnd = true
				byteLength = byteLength - i - 1
			default:
				hasEnd = false
			}
		}

		if (hasEnd && hasStart) || (i >= byteLength-i-1) {
			break
		}
	}

	// if it isn't a valid JSON, treat it as a debug level message
	if !(hasEnd && hasStart) {
		pl["level"] = fmt.Sprintf("%q", debugLevel)
		pl["ts"] = fmt.Sprintf("%v", time.Now().Unix())
		pl["caller"] = `"user-code"`
		pl["msg"] = strings.TrimSpace(string(jsonByte))
		return pl, true
	}

	// search for key-value pairs inside the byte array
	for offset < byteLength {
		switch jsonByte[offset] {
		case '"':
			k, no := findKey(jsonByte[offset+1:])
			foundKey = strings.TrimSpace(string(k))
			nextOffset = no + 1
		case ':':
			v, no := findValue(jsonByte[offset+1:])
			pl[foundKey] = strings.TrimSpace(string(v))
			nextOffset = no + 1
		}

		offset = offset + nextOffset + 1
		nextOffset = 0
	}
	return pl, true
}

// PrettyPrint writes the pretty version of the parsed JSON in the given writer.
func PrettyPrint(w io.Writer, pj ParsedJSON, f LogFilter, emoji bool) error {
	if filterJSON(pj, f) {
		t, err := GenerateOutputString(pj, emoji)
		if err != nil {
			return err
		}
		w.Write([]byte(t))
	}
	return nil
}
