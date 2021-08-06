package logrusjournal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/journal"
	"github.com/sirupsen/logrus"
)

// Hook A logrus hook for logging to systemd-journal
type Hook struct{}

// User Journal Fields as specified at: https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
const (
	JournalMessageID        = "MESSAGE_ID"
	JournalCodeFile         = "CODE_FILE"
	JournalCodeLine         = "CODE_LINE"
	JournalCodeFunc         = "CODE_FUNC"
	JournalErrno            = "ERRNO"
	JournalInvocationID     = "INVOCATION_ID"
	JournalUserInvocationID = "USER_INVOCATION_ID"
	JournalDocumentation    = "DOCUMENTATION"
	JournalThreadID         = "TID"
)

var (
	levelMap = map[logrus.Level]journal.Priority{
		logrus.TraceLevel: journal.PriDebug, // No separate trace level in journald
		logrus.DebugLevel: journal.PriDebug,
		logrus.InfoLevel:  journal.PriInfo,
		logrus.WarnLevel:  journal.PriWarning,
		logrus.ErrorLevel: journal.PriErr,
		logrus.FatalLevel: journal.PriCrit,
		logrus.PanicLevel: journal.PriAlert,
	}
)

func (j Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (j Hook) Fire(entry *logrus.Entry) error {
	fields := convertFields(entry.Data)

	// Annotate with code location information if available
	if entry.Caller.File != "" {
		fields[JournalCodeFile] = entry.Caller.File
	}
	if entry.Caller.Line != 0 {
		fields[JournalCodeLine] = strconv.Itoa(entry.Caller.Line)
	}
	if entry.Caller.Function != "" {
		fields[JournalCodeFunc] = entry.Caller.Function
	}

	return journal.Send(entry.Message, levelMap[entry.Level], fields)
}

func convertFields(fields logrus.Fields) map[string]string {
	entries := make(map[string]string, len(fields))

	for k, v := range fields {
		key := convertFieldName(k)
		entries[key] = fmt.Sprint(v)
	}

	return entries
}

func convertFieldName(s string) string {
	return strings.TrimLeft(strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r >= 'a' && r <= 'z':
			return r - 32
		default:
			return '_'
		}
	}, s), "_")
}
