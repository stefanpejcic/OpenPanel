// Package config reads /etc/openpanel/openpanel/conf/openpanel.config, a
// plain key=value configuration file.
package config

import (
	"bufio"
	"os"
	"regexp"
)

var lineRE = regexp.MustCompile(`^(\w+)=(.*)$`)

// Config is the parsed key=value file. Missing keys fall back to a
// caller-supplied default, via Get.
type Config map[string]string

// Load reads path and parses it line by line. A missing file is not an
// error — it returns an empty Config.
func Load(path string) (Config, error) {
	cfg := Config{}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := lineRE.FindStringSubmatch(scanner.Text()); m != nil {
			cfg[m[1]] = m[2]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Get returns the value for key, or def if key is unset.
func (c Config) Get(key, def string) string {
	if v, ok := c[key]; ok {
		return v
	}
	return def
}
