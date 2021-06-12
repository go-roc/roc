package log

import (
	"path"
	"runtime"
	"strconv"
	"time"

	"roc/rlog/common"
	"roc/rlog/format"
	"roc/rlog/output"
)

func init() {
	// Call(4) is the actual line where used
	//Overload(Call(-1))
	Overload(Call(4))
}

var defaultLogger *log

type log struct {
	opts   Option
	detail *common.Detail
}

func Overload(opts ...Options) {
	if defaultLogger != nil {
		defaultLogger = nil
	}

	defaultLogger = &log{opts: newOpts(opts...)}

	defaultLogger.detail = &common.Detail{
		Name:   defaultLogger.Name(),
		Prefix: defaultLogger.Prefix(),
	}
}

func (l *log) Fire(level, msg string) *common.Detail {
	d := *l.detail
	if l.opts.call >= 0 {
		d.Line = l.caller()
	}
	d.Timestamp = time.Now().Format(l.Formatter().Layout())
	d.Level = level
	d.Content = msg
	return &d
}

func (l *log) Name() string {
	return l.opts.name
}

func (l *log) Prefix() string {
	return l.opts.prefix
}

func (l *log) Formatter() format.Formatter {
	return l.opts.format
}

func (l *log) Output() output.Outputor {
	return l.opts.out
}

func (l *log) caller() string {
	pc, file, line, _ := runtime.Caller(l.opts.call)
	funcName := runtime.FuncForPC(pc).Name()
	return path.Base(file) + "/" + path.Base(funcName) + "." + strconv.Itoa(line)
}

func Close() {
	defaultLogger.opts.out.Close()
}

func Debug(content string) {
	defaultLogger.
		Output().
		Out(common.DEBUG, defaultLogger.
			Formatter().
			Format(defaultLogger.Fire("DBUG", content)),
		)
}

func Info(content string) {
	defaultLogger.
		Output().
		Out(common.INFO, defaultLogger.
			Formatter().
			Format(defaultLogger.Fire("INFO", content)),
		)
}

func Warn(content string) {
	defaultLogger.
		Output().
		Out(common.WARN, defaultLogger.
			Formatter().
			Format(defaultLogger.Fire("WARN", content)),
		)
}

func Error(content string) {
	defaultLogger.
		Output().
		Out(common.ERR, defaultLogger.
			Formatter().
			Format(defaultLogger.Fire("ERRO", content)),
		)
}

func Fatal(content string) {
	defaultLogger.
		Output().
		Out(common.FATAL, defaultLogger.
			Formatter().
			Format(defaultLogger.Fire("FATA", content)),
		)
}

func Stack(content string) {

	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	content += string(buf[:n]) + "\n"

	defaultLogger.
		Output().
		Out(common.STACK, defaultLogger.
			Formatter().
			Format(defaultLogger.Fire("STAK", content)),
		)
}