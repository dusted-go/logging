package stackdriver

import (
	"log/slog"
	"strconv"
	"strings"
)

// The slog package provides four log levels by default,
// and each one is associated with an integer value:
// DEBUG (-4), INFO (0), WARN (4), and ERROR (8).
const (
	DEBUG     = slog.LevelDebug
	INFO      = slog.LevelInfo
	NOTICE    = slog.Level(2)
	WARNING   = slog.LevelWarn
	ERROR     = slog.LevelError
	CRITICAL  = slog.Level(10)
	ALERT     = slog.Level(12)
	EMERGENCY = slog.Level(14)
)

var levelNames = map[slog.Leveler]string{
	NOTICE:    "NOTICE",
	CRITICAL:  "CRITICAL",
	ALERT:     "ALERT",
	EMERGENCY: "EMERGENCY",
}

func ReplaceLogLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		value, ok := levelNames[level]
		if !ok {
			value = level.String()
		}
		a.Key = "severity"
		a.Value = slog.StringValue(value)
		return a
	}
	return a
}

func ParseLogLevel(value string) slog.Leveler {
	if len(value) > 0 {
		v := strings.ToLower(strings.Trim(value, " "))

		if v == "debug" {
			return DEBUG
		}
		if v == "info" {
			return INFO
		}
		if v == "notice" {
			return NOTICE
		}
		if v == "warning" {
			return WARNING
		}
		if v == "error" {
			return ERROR
		}
		if v == "critical" {
			return CRITICAL
		}
		if v == "alert" {
			return ALERT
		}
		if v == "emergency" {
			return EMERGENCY
		}

		if i, err := strconv.Atoi(value); err == nil {
			return slog.Level(i)
		}
	}
	return INFO
}
