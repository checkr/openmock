package openmock

import "github.com/sirupsen/logrus"

type omLogger struct {
	*logrus.Entry
	ctx Context
}

func newOmLogger(ctx Context) *omLogger {
	currentMock := &Mock{}
	if ctx.currentMock != nil {
		currentMock = ctx.currentMock
	}
	entry := logrus.NewEntry(logrus.StandardLogger()).WithFields(logrus.Fields{
		"current_mock_key": currentMock.Key,
	})
	return &omLogger{
		Entry: entry,
		ctx:   ctx,
	}
}

func (l *omLogger) SetPrefix(prefix string) {
	l.Entry = l.Entry.WithFields(logrus.Fields{"logger_prefix": prefix})
}
