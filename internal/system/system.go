package system

import (
	"os"
	"regexp"
	"strings"
)

func GetVersion() string {
	data, err := os.ReadFile("/etc/version")
	if err != nil {
		return "-"
	}

	version := regexp.MustCompile(`\d+\.\d+\.\d+`).FindString(string(data))
	if version == "" {
		return "-"
	}

	return version
}

func GetSerialNumber() string {
	data, err := os.ReadFile("/etc/serialnum")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}
