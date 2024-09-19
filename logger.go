package zaplogger

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Option func(option *options) error

type options struct {
	level           *zapcore.Level
	timeFormat      zapcore.TimeEncoder
	filePath        *string
	pretty          *bool
	caller          *bool
	rotateAtStartup *bool
	maxSize         *int
	maxBackups      *int
	maxAge          *int
	localtime       *bool
	compress        *bool
}

func New(opts ...Option) (*zap.Logger, error) {
	var opt options
	for _, option := range opts {
		if err := option(&opt); err != nil {
			return nil, err
		}
	}

	productionCfg := zap.NewProductionEncoderConfig()
	developmentCfg := zap.NewDevelopmentEncoderConfig()

	var lvl zapcore.Level
	if opt.level != nil {
		lvl = *opt.level
	} else {
		lvl = zap.DebugLevel
	}
	if opt.timeFormat != nil {
		// productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = opt.timeFormat
		developmentCfg.EncodeTime = opt.timeFormat
	}

	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	stdout := zapcore.AddSync(os.Stdout)

	

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)
	if opt.pretty != nil && *opt.pretty {
		productionCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		fileEncoder = zapcore.NewConsoleEncoder(productionCfg)
	}

	cores := []zapcore.Core{zapcore.NewCore(consoleEncoder, stdout, lvl)}
	if opt.filePath != nil && *opt.filePath != "" {
		file := zapcore.AddSync(newRollingFile(*opt.filePath, opt))
		cores = append(cores, zapcore.NewCore(fileEncoder, file, lvl))
	}
	core := zapcore.NewTee(cores...)
	log := zap.New(core)
	if opt.caller != nil && *opt.caller {
		
		log = zap.New(core, zap.AddCaller())
	}
	return log, nil
}

func WithLevel(level string) Option {
	return func(options *options) error {
		lvl, err := zapcore.ParseLevel(level)
		if err != nil {
			return err
		}
		options.level = &lvl
		return nil
	}
}

func WithCustomTimestamp(timeformat string) Option {
	return func(options *options) error {
		options.timeFormat = zapcore.TimeEncoderOfLayout(timeformat)
		return nil
	}
}

func WithPretty(with bool) Option {
	return func(options *options) error {
		options.pretty = &with
		return nil
	}
}

func WithCaller(with bool) Option {
	return func(options *options) error {
		options.caller = &with
		return nil
	}
}


func WithFile(filepath string) Option {
	return func(options *options) error {
		options.filePath = &filepath
		return nil
	}
}

func WithRotateAtStartup(with bool) Option {
	return func(options *options) error {
		options.rotateAtStartup = &with
		return nil
	}
}

func WithCompress(with bool) Option {
	return func(options *options) error {
		options.compress = &with
		return nil
	}
}

func WithLocalTime(with bool) Option {
	return func(options *options) error {
		options.localtime = &with
		return nil
	}
}

func WithMaxSize(size int) Option {
	return func(options *options) error {
		if size < 0 {
			return fmt.Errorf("file size cannot be less than zero")
		}
		options.maxSize = &size
		return nil
	}
}

func WithMaxBackups(backups int) Option {
	return func(options *options) error {
		if backups < 0 {
			return fmt.Errorf("number of files cannot be less than zero")
		}
		options.maxBackups = &backups
		return nil
	}
}

func WithMaxAge(age int) Option {
	return func(options *options) error {
		if age < 0 {
			return fmt.Errorf("number of days cannot be less than zero")
		}
		options.maxAge = &age
		return nil
	}
}

func newRollingFile(pathfile string, opts options) io.Writer {
	var (
		localtime, compress         bool
		maxsize, maxbackups, maxage int
	)

	if opts.localtime != nil {
		localtime = *opts.localtime
	}
	if opts.compress != nil {
		compress = *opts.compress
	}
	if opts.maxSize != nil {
		maxsize = *opts.maxSize
	}
	if opts.maxBackups != nil {
		maxbackups = *opts.maxBackups
	}
	if opts.maxAge != nil {
		maxage = *opts.maxAge
	}

	rotator := &lumberjack.Logger{
		Filename:   pathfile,
		LocalTime:  localtime,
		Compress:   compress,
		MaxSize:    maxsize,
		MaxBackups: maxbackups,
		MaxAge:     maxage,
	}
	if opts.rotateAtStartup != nil && *opts.rotateAtStartup {
		rotator.Rotate()
	}
	return rotator
}
