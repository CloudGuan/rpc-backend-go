package logger

//NullLogger null log anymore
type NullLogger struct{}

func (n *NullLogger) Debug(string, ...interface{}) {}
func (n *NullLogger) Info(string, ...interface{})  {}
func (n *NullLogger) Warn(string, ...interface{})  {}
func (n *NullLogger) Error(string, ...interface{}) {}
func (n *NullLogger) Fatal(string, ...interface{}) {}
func (n *NullLogger) LogLevel()                    {}
