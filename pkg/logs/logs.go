// 2026/5/15 Bin Liu <bin.liu@enmotech.com>

package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/toolkits/file"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var logLevels = map[string]logrus.Level{
	"PANIC": logrus.PanicLevel,
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	"FATAL": logrus.FatalLevel,
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	"ERROR": logrus.ErrorLevel,
	// WarnLevel level. Non-critical entries that deserve eyes.
	"WARN": logrus.WarnLevel,
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	"INFO": logrus.InfoLevel,
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	"DEBUG": logrus.DebugLevel,
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	"TRACE": logrus.TraceLevel,
}

func NewLogRusDir(logDir, logfile, logLevel string) (*logrus.Logger, error) {
	if logDir == "" {
		return NewLogRus("", logLevel)
	}
	return NewLogRus(filepath.Join(logDir, logfile), logLevel)
}

func NewLogRus(logfile, logLevel string) (*logrus.Logger, error) {
	logger := newLogRus()
	if err := setLogLevel(logger, logLevel); err != nil {
		return nil, err
	}
	if logfile == "" {
		return logger, nil
	}
	if err := InitLogRusLogFile(logger, logfile, true); err != nil {
		return nil, err
	}
	return logger, nil
}

func InitLogRusLogFile(logger *logrus.Logger, logfile string, rotate bool) error {
	var (
		writer io.Writer
		err    error
	)
	writer, err = newFileRotateLog(logfile, rotate)
	if err != nil {
		return err
	}
	logger.SetOutput(io.MultiWriter(os.Stdout, writer))
	return nil
}

func MkdirAll(dir string) error {
	return afero.NewOsFs().MkdirAll(dir, 0755)
}

func newFileRotateLog(logfile string, rotate bool) (io.Writer, error) {
	if err := MkdirAll(filepath.Dir(logfile)); err != nil {
		return nil, err
	}
	if rotate {
		if !strings.HasPrefix(logfile, "/") {
			pwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			logfile = filepath.Join(pwd, logfile)
		}
		return rotatelogs.New(
			logfile+".%Y%m%d",
			rotatelogs.WithLinkName(logfile),          // 生成软链，指向最新日志文件
			rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
			rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
			rotatelogs.WithLocation(time.Local),
		)
	}
	return file.MustOpenLogFile(logfile), nil
}

func setLogLevel(logger *logrus.Logger, logLevel string) error {
	if logLevel == "" {
		return nil
	}
	logLevel = strings.ToUpper(logLevel)
	level, ok := logLevels[logLevel]
	if !ok {
		return fmt.Errorf("log level not support. please set INFO,WARN,ERROR,DEBUG,TRACE")
	}
	logger.SetLevel(level)
	return nil
}

func newLogRus() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = newLogRusFormat(true)
	return logger
}

func newTextFormatter() *prefixed.TextFormatter {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableColors = true
	formatter.DisableSorting = true
	formatter.FullTimestamp = true                           // 显示完整时间
	formatter.TimestampFormat = "2006-01-02 15:04:05.000000" // 时间格式
	formatter.DisableTimestamp = false
	return formatter
}

func newLogRusFormat(Caller bool) logrus.Formatter {
	formatter := newTextFormatter()
	if !Caller {
		return formatter
	}
	f := &runtime.Formatter{
		ChildFormatter: formatter,
		Line:           true,
		File:           true,
	}
	return f
}
