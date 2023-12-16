package src

import (
	"reflect"
	"testing"
	"unsafe"
)

type logger struct {
	t *testing.T
}

func (l *logger) Printf(fmt string, v ...any) {
	l.t.Logf(fmt, v...)
}

func (l *logger) Println(v ...any) {
	l.t.Log(v...)
}

func SetLogger(cfg interface{}, t *testing.T) {
	logger := &logger{t: t}
	pcfgType := reflect.ValueOf(cfg)
	field := pcfgType.Elem().FieldByName("log")
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	field.Set(reflect.ValueOf(logger))
}
