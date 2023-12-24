package application

func TrimQuote(value string) string {

	if len(value) > 0 && value[0] == '"' {
		value = value[1:]
	}
	if l := len(value); l > 0 && value[l-1] == '"' {
		value = value[:l-1]
	}
	return value
}
