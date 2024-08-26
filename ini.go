package ini

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	fileName = "config.ini"
)

var (
	config     map[string]string
	lastLoaded time.Time
	mutex      sync.Mutex
)

func init() {
	lastLoaded = time.Unix(0, 0)
	reload()
}

func reload() {
	config = make(map[string]string)
	if file, err := os.Open(fileName); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if key, value, found := strings.Cut(line, "="); found {
				key := strings.ToUpper(strings.TrimSpace(key))
				if len(key) > 0 {
					if strings.HasPrefix(key, ";") || strings.HasPrefix(key, "#") {
						// Comment
					} else {
						value := strings.TrimSpace(value)
						if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
							value = value[1 : len(value)-1]
						}
						config[key] = value
					}
				}
			}
		}
		file.Close()
	}
	file, err := os.Stat(fileName)
	if err == nil {
		lastLoaded = file.ModTime()
	}
}

func SetFile(name string) {
	fileName = name
	lastLoaded = time.Unix(0, 0)
	reload()
}

func GetLastLoaded() time.Time {
	mutex.Lock()
	defer mutex.Unlock()
	return lastLoaded
}

func GetString(key, def string) string {
	mutex.Lock()
	defer mutex.Unlock()
	fi, err := os.Stat(fileName)
	if err == nil {
		if fi.ModTime().After(lastLoaded) {
			reload()
		}
	}
	if text, ok := config[strings.ToUpper(key)]; ok {
		return text
	}
	return def
}

func GetInt(key string, def int) int {
	s := GetString(key, strconv.Itoa(def))
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func GetInt64(key string, def int64) int64 {
	s := GetString(key, strconv.FormatInt(def, 10))
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return i
}

func GetStrings(key string, callback func(string) string) []string {
	out := []string{}
	if s := GetString(key, ""); s != "" {
		a := strings.Split(s, ",")
		for _, t := range a {
			if t = strings.TrimSpace(t); t != "" {
				if callback != nil {
					t = callback(t)
				}
				out = append(out, t)
			}
		}
	}
	return out
}

func RequireSet(keys []string) bool {
	for _, s := range keys {
		if GetString(s, "") == "" {
			return false
		}
	}
	return true
}

func GetDuration(key string, def time.Duration) time.Duration {
	s := GetString(key, "")
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}
