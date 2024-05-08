// Package appctx
package appctx

type ValidationMessage map[string][]string

// NewValidationMessage create new ValidationMessage
func NewValidationMessage() ValidationMessage {
	return ValidationMessage{}
}

// AddMessage setter
func (v ValidationMessage) AddMessage(field string, msg ...string) ValidationMessage {
	_, ok := v[field]
	if !ok {
		v[field] = msg
	}

	return v
}
