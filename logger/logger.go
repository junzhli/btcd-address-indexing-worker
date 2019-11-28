package logger

import "log"

// FailOnError prints out message if fail occurs and exits with code 1
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\n", msg, err)
	}
}

// LogOnError prints out message if fail occurs
func LogOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s\n", msg, err)
	}
}

type customLogger struct {
	logger *log.Logger
}

// FailOnError prints out message if fail occurs and exits with code 1
func (c customLogger) FailOnError(err error, msg string) {
	if err != nil {
		c.logger.Fatalf("%s: %s\n", msg, err)
	}
}

// LogOnError prints out message if fail occurs
func (c customLogger) LogOnError(err error, msg string) {
	if err != nil {
		c.logger.Printf("%s: %s\n", msg, err)
	}
}

// CustomLogger replaces itself logger with user-provided one
type CustomLogger interface {
	FailOnError(err error, msg string)
	LogOnError(err error, msg string)
}

// New creates a instance of CustomLogger and returns
func New(lg *log.Logger) CustomLogger {
	return &customLogger{
		logger: lg,
	}
}
