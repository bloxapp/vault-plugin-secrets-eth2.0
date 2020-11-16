// Package logex configures a new logger for an application.
package logex

import (
	"os"

	"github.com/makasim/sentryhook"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bloxapp/key-vault/utils/sentry"
)

// Options contains required options to initialize a new logger.
type Options struct {
	// Format specifies the output log format.
	// Accepted values are: json, logfmt
	Format string

	// NoColor makes sure that no log output gets colorized.
	NoColor bool

	// This is the list of log levels.
	Levels []string

	// DSN is the DSN of a external logs store.
	DSN string
}

// Init creates a new logger and sets up its options.
func Init(opts Options) (*logrus.Logger, error) {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = &logrus.TextFormatter{
		DisableColors: opts.NoColor,
		ForceColors:   true,
	}

	// Define format.
	switch format := opts.Format; format {
	case "logfmt", "":
		// Already the default
		break
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
		break
	default:
		return nil, errors.Errorf("undefined logs format: %s", format)
	}

	// Prepare external logs store configuration.
	if opts.DSN != "" {
		// Init Sentry first
		if err := sentry.Init(opts.DSN); err != nil {
			return nil, errors.Wrap(err, "failed to init Sentry")
		}

		// Set default log levels
		logLevels := []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}

		// Parse log levels if exists
		if lvls := opts.Levels; len(lvls) > 0 {
			logLevels = []logrus.Level{}
			for _, lvl := range lvls {
				level, err := logrus.ParseLevel(lvl)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse external log level '%s'", lvl)
				}

				logLevels = append(logLevels, level)
			}
		}

		// Add Sentry log hook
		logger.Hooks.Add(sentryhook.New(logLevels))
	}

	return logger, nil
}
