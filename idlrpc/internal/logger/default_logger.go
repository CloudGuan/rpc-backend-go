package logger

import "fmt"

// DefaultLogger rpc-backend-cpp logger, using fmt package
type DefaultLogger struct{}

func (s *DefaultLogger) Debug(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (s *DefaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (s *DefaultLogger) Warn(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (s *DefaultLogger) Error(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (s *DefaultLogger) Fatal(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("\n")
}

func (s *DefaultLogger) LogLevel() {

}
