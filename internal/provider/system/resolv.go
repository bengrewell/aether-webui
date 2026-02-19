package system

import (
	"bufio"
	"os"
	"strings"
)

// parseResolvConf parses a resolv.conf-style file for nameserver and search directives.
func parseResolvConf(path string) (servers []string, search []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver":
			servers = append(servers, fields[1])
		case "search":
			search = append(search, fields[1:]...)
		}
	}
	return servers, search, sc.Err()
}
