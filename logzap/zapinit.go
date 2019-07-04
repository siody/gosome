package logzap

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"strings"
)

var (
	corelist []zapcore.Core
	//Logger global logger
	Logger *zap.Logger
)

//AppendCore new a out put can sync when app close
func AppendCore(level zapcore.Level, stream zapcore.WriteSyncer) {

	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level
	})
	consoleStdout := zapcore.Lock(stream)
	var consoleEncoder zapcore.Encoder
	if level == zapcore.DebugLevel {
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	}
	corelist = append(corelist, zapcore.NewCore(consoleEncoder, consoleStdout, priority))

}

//AppendCoreSync new a out put sync writer
func AppendCoreSync(level zapcore.Level, stream io.Writer) {

	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level
	})
	consoleStdout := zapcore.AddSync(stream)
	var consoleEncoder zapcore.Encoder
	if level == zapcore.DebugLevel {
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		consoleEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}
	corelist = append(corelist, zapcore.NewCore(consoleEncoder, consoleStdout, priority))

}

//Build build log from corelist
func Build() {
	Logger = zap.New(zapcore.NewTee(corelist...))
}

//NewLogFile new file log
func NewLogFile(filename string) io.Writer {
	//log part
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    1000,
		MaxAge:     1,
		MaxBackups: 10,
	}
}

//NewFileCore file log
func NewFileCore(level zapcore.Level) {
	AppendCoreSync(level, NewLogFile(viper.GetString("log")))
}

//NewConsoleCore file log
func NewConsoleCore(level zapcore.Level) {
	AppendCore(level, os.Stdout)
}

//LogConfig config file of log
type LogConfig struct {
	OutPut []struct {
		Type   string
		Stream string
		Level  string
	}
}

//FromConfig build logger form config
func FromConfig() {
	cfg := new(LogConfig)
	viper.UnmarshalKey("logconfig", cfg)
	for _, endpoiont := range cfg.OutPut {
		switch strings.ToLower(endpoiont.Type) {
		case "console":
			switch strings.ToLower(endpoiont.Stream) {
			default:
				fallthrough
			case "stdout":
				AppendCore(resolvestringlevel(endpoiont.Level), os.Stdout)
			case "stderr":
				AppendCore(resolvestringlevel(endpoiont.Level), os.Stderr)
			}
		case "file":
			AppendCoreSync(resolvestringlevel(endpoiont.Level), NewLogFile(endpoiont.Stream))
		}
	}

	Build()
}

func resolvestringlevel(l string) zapcore.Level {
	switch strings.ToLower(l) {
	default:
		fallthrough
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	}
}
