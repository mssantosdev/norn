package logger

import (
	"os"

	clog "github.com/charmbracelet/log"
	"github.com/mssantosdev/norn/internal/ui/themes"
)

type Options struct {
	Level string
	Theme string
}

var instance = clog.NewWithOptions(os.Stderr, clog.Options{ReportTimestamp: false})

func Init(opts Options) {
	if opts.Theme != "" {
		themes.Set(opts.Theme)
	}
	instance.SetStyles(defaultStyles())
	if opts.Level != "" {
		if level, err := clog.ParseLevel(opts.Level); err == nil {
			instance.SetLevel(level)
		}
	}
}

func defaultStyles() *clog.Styles {
	styles := clog.DefaultStyles()
	styles.Levels[clog.InfoLevel] = styles.Levels[clog.InfoLevel].Foreground(themes.Current.Primary)
	styles.Levels[clog.WarnLevel] = styles.Levels[clog.WarnLevel].Foreground(themes.Current.Warning)
	styles.Levels[clog.ErrorLevel] = styles.Levels[clog.ErrorLevel].Foreground(themes.Current.Error)
	styles.Levels[clog.DebugLevel] = styles.Levels[clog.DebugLevel].Foreground(themes.Current.Muted)
	return styles
}

func Logger() *clog.Logger { return instance }

func Debug(msg string, keyvals ...any) { instance.Debug(msg, keyvals...) }
func Info(msg string, keyvals ...any)  { instance.Info(msg, keyvals...) }
func Warn(msg string, keyvals ...any)  { instance.Warn(msg, keyvals...) }
func Error(msg string, keyvals ...any) { instance.Error(msg, keyvals...) }
func Print(msg string, keyvals ...any) { instance.Info(msg, keyvals...) }
