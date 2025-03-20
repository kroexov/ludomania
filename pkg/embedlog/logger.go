package embedlog

import (
	"fmt"
	"log"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

var statLogEvents *prometheus.CounterVec

// SetStatLogEvents sets prometheus counter for error and debug log events.
func SetStatLogEvents(stat *prometheus.CounterVec) {
	statLogEvents = stat
}

// Logger is a struct for embedding std loggers.
type Logger struct {
	warn, log *log.Logger
}

// Printf prints message to Stdout (app.log variable) if a.verbose is set.
func (l Logger) Printf(format string, v ...interface{}) {
	if l.log != nil {
		if err := l.log.Output(2, fmt.Sprintf(format, v...)); err != nil {

			if statLogEvents != nil {
				statLogEvents.WithLabelValues("debug").Inc()
			}
		}
	}
}

// Errorf prints message to Stderr (l.warn variable).
func (l Logger) Errorf(format string, v ...interface{}) {
	if l.warn != nil {
		if err := l.warn.Output(2, fmt.Sprintf(format, v...)); err != nil {
			if statLogEvents != nil {
				statLogEvents.WithLabelValues("error").Inc()
			}
		}
	}
}

func (l *Logger) SetStdLoggers(verbose bool) {
	if verbose {
		l.log = log.New(os.Stdout, "D", log.LstdFlags|log.Lshortfile)
	}

	l.warn = log.New(os.Stderr, "E", log.LstdFlags|log.Lshortfile)
}

func (l Logger) Warn() *log.Logger                 { return l.warn }
func (l Logger) Log() *log.Logger                  { return l.log }
func (l Logger) Loggers() (warn, log *log.Logger)  { return l.Warn(), l.Log() }
func (l *Logger) SetLoggers(warn, log *log.Logger) { l.warn, l.log = warn, log }
