package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/logidoc/logidoc-server/internal/config"
)

func New(cfg config.LoggerConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch cfg.Format {
	case "text":
		handler = &prettyHandler{w: os.Stdout, level: level}
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

const (
	reset   = "\033[0m"
	dim     = "\033[2m"
	bold    = "\033[1m"
	red     = "\033[31m"
	yellow  = "\033[33m"
	green   = "\033[32m"
	blue    = "\033[34m"
	cyan    = "\033[36m"
	magenta = "\033[35m"
)

type prettyHandler struct {
	w     io.Writer
	level slog.Level
	group string
	attrs []slog.Attr
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format("15:04:05")

	lvl, color := levelStyle(r.Level)

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s%s%s %s%s%-5s%s %s%s%s",
		dim, ts, reset,
		color, bold, lvl, reset,
		bold, r.Message, reset,
	)

	// Pre-set attrs
	for _, a := range h.attrs {
		writeAttr(&sb, a)
	}

	// Record attrs
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(&sb, a)
		return true
	})

	sb.WriteString("\n")
	_, err := fmt.Fprint(h.w, sb.String())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{w: h.w, level: h.level, group: h.group, attrs: append(h.attrs, attrs...)}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{w: h.w, level: h.level, group: name, attrs: h.attrs}
}

func levelStyle(l slog.Level) (string, string) {
	switch {
	case l >= slog.LevelError:
		return "ERROR", red
	case l >= slog.LevelWarn:
		return "WARN ", yellow
	case l >= slog.LevelInfo:
		return "INFO ", green
	default:
		return "DEBUG", blue
	}
}

func writeAttr(sb *strings.Builder, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}

	key := a.Key
	val := a.Value.Resolve()

	keyColor := cyan
	valStr := formatValue(val)

	fmt.Fprintf(sb, " %s%s%s%s=%s%s", keyColor, key, reset, dim, valStr, reset)
}

func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindDuration:
		d := v.Duration()
		if d < time.Second {
			return fmt.Sprintf("%s%dms%s", magenta, d.Milliseconds(), reset)
		}
		return fmt.Sprintf("%s%s%s", magenta, d.Round(time.Millisecond), reset)
	case slog.KindInt64:
		return fmt.Sprintf("%s%d%s", magenta, v.Int64(), reset)
	case slog.KindFloat64:
		return fmt.Sprintf("%s%.2f%s", magenta, v.Float64(), reset)
	case slog.KindBool:
		if v.Bool() {
			return fmt.Sprintf("%strue%s", green, reset)
		}
		return fmt.Sprintf("%sfalse%s", red, reset)
	default:
		s := v.String()
		if strings.HasPrefix(s, "http") {
			return fmt.Sprintf("%s%s%s", blue, s, reset)
		}
		return s
	}
}
