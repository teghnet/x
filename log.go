package x

import (
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
)

// NoticeLevel sits between [log.DebugLevel] and [log.InfoLevel].
const NoticeLevel log.Level = -1

func init() {
	styles := log.DefaultStyles()
	styles.Levels[NoticeLevel] = lipgloss.NewStyle().
		SetString(strings.ToUpper("notice")).
		Bold(true).
		MaxWidth(4).
		Foreground(lipgloss.Color("212"))
	log.SetStyles(styles)

	log.SetLevel(NoticeLevel)
}

// Debug logs a debug message using the default logger.
func Debug(msg any, keyvals ...any) { log.Debug(msg, keyvals...) }

// Notice logs a message at [NoticeLevel] using the default logger.
func Notice(msg any, keyvals ...any) { log.Log(NoticeLevel, msg, keyvals...) }

// Info logs an info message using the default logger.
func Info(msg any, keyvals ...any) { log.Info(msg, keyvals...) }

// Warn logs a warning message using the default logger.
func Warn(msg any, keyvals ...any) { log.Warn(msg, keyvals...) }

// Error logs an error message using the default logger.
func Error(msg any, keyvals ...any) { log.Error(msg, keyvals...) }

// Fatal logs a fatal message using the default logger and exits.
func Fatal(msg any, keyvals ...any) { log.Fatal(msg, keyvals...) }

// Print logs a message without a level using the default logger.
func Print(msg any, keyvals ...any) { log.Print(msg, keyvals...) }

// Debugf logs a formatted debug message using the default logger.
func Debugf(format string, args ...any) { log.Debugf(format, args...) }

// Noticef logs a formatted message at [NoticeLevel] using the default logger.
func Noticef(format string, args ...any) { log.Logf(NoticeLevel, format, args...) }

// Infof logs a formatted info message using the default logger.
func Infof(format string, args ...any) { log.Infof(format, args...) }

// Warnf logs a formatted warning message using the default logger.
func Warnf(format string, args ...any) { log.Warnf(format, args...) }

// Errorf logs a formatted error message using the default logger.
func Errorf(format string, args ...any) { log.Errorf(format, args...) }

// Fatalf logs a formatted fatal message using the default logger and exits.
func Fatalf(format string, args ...any) { log.Fatalf(format, args...) }

// Printf logs a formatted message without a level using the default logger.
func Printf(format string, args ...any) { log.Printf(format, args...) }

func PrintErr(err error) {
	if err != nil {
		Printf("err: %v", err)
	}
}
func WarnErr(err error) {
	if err != nil {
		Warnf("err: %v", err)
	}
}
