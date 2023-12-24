package internal

type Logger interface {
	Printf(fmt string, v ...any)
	Println(v ...any)
}

type NopLogger struct{}

func (log *NopLogger) Printf(fmt string, v ...any) {
	// NOP
}

func (log *NopLogger) Println(v ...any) {
	// NOP
}

var Log Logger = &NopLogger{}
