package formatter

var Default = newDefault()

type defaultFormatter struct{}

func newDefault() Formatter {
	return defaultFormatter{}
}

func (df defaultFormatter) Format(s string) string {
	return s
}
