package prettierzap

type parsedLog map[string]string

// GetLevel returns the level of log
func (pl parsedLog) GetLevel() string {
	return pl["level"]
}

// GetTimestamp returns the timestamp of the log
func (pl parsedLog) GetTimestamp() string {
	return pl["ts"]
}

// GetCaller returns the caller field of the log
func (pl parsedLog) GetCaller() string {
	return pl["caller"]
}

// GetMsg returns the message of the log
func (pl parsedLog) GetMsg() string {
	keys := []string{"msg", "message"}

	for _, k := range keys {
		if v, ok := pl[k]; ok {
			return v
		}
	}
	return ""
}

// GetMeta returns meta data of the log
// meta data is every key-value pair that its key is not in ("msg", "message", "caller", "ts", "level")
func (pl parsedLog) GetMeta() map[string]string {
	m := make(map[string]string, 0)
	for key := range pl {
		switch key {
		case "level", "ts", "caller", "msg", "message":
			continue
		default:
			m[key] = pl[key]
		}
	}
	return m
}
