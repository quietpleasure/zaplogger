package zaplogger

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Hook struct {
	zapcore.Core
	Hooker hooker
}

type hooker interface {
	SendLog(ctx context.Context, ts time.Time, lvl string, msg []byte) error
}

func (hc *Hook) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if hc.Enabled(entry.Level) {
		return checked.AddCore(entry, hc)
	}
	return checked
}

func (hc *Hook) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return hc.sendLog(entry, fields)
}

func (hc *Hook) sendLog(en zapcore.Entry, fields []zapcore.Field) error {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("02.01.2006 15:04:05.000000")
	cfg.CallerKey = "caller"
	cfg.StacktraceKey = "stack"
	buf, err := zapcore.NewJSONEncoder(cfg).EncodeEntry(en, fields)
	if err != nil {
		return err
	}
	type log struct {
		Timestamp string `json:"ts"`
		Message   string `json:"msg"`
		Level     string `json:"lvl"`
	}
	var l log
	if err := json.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&l); err != nil {
		return err
	}
	return hc.Hooker.SendLog(
		context.TODO(),
		en.Time,
		en.Level.String(),
		buf.Bytes(),
	)
}
