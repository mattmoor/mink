// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: envoy/type/token_bucket.proto

package envoy_type

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on TokenBucket with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *TokenBucket) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TokenBucket with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in TokenBucketMultiError, or
// nil if none found.
func (m *TokenBucket) ValidateAll() error {
	return m.validate(true)
}

func (m *TokenBucket) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetMaxTokens() <= 0 {
		err := TokenBucketValidationError{
			field:  "MaxTokens",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if wrapper := m.GetTokensPerFill(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := TokenBucketValidationError{
				field:  "TokensPerFill",
				reason: "value must be greater than 0",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	if m.GetFillInterval() == nil {
		err := TokenBucketValidationError{
			field:  "FillInterval",
			reason: "value is required",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if d := m.GetFillInterval(); d != nil {
		dur, err := d.AsDuration(), d.CheckValid()
		if err != nil {
			err = TokenBucketValidationError{
				field:  "FillInterval",
				reason: "value is not a valid duration",
				cause:  err,
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		} else {

			gt := time.Duration(0*time.Second + 0*time.Nanosecond)

			if dur <= gt {
				err := TokenBucketValidationError{
					field:  "FillInterval",
					reason: "value must be greater than 0s",
				}
				if !all {
					return err
				}
				errors = append(errors, err)
			}

		}
	}

	if len(errors) > 0 {
		return TokenBucketMultiError(errors)
	}
	return nil
}

// TokenBucketMultiError is an error wrapping multiple validation errors
// returned by TokenBucket.ValidateAll() if the designated constraints aren't met.
type TokenBucketMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TokenBucketMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TokenBucketMultiError) AllErrors() []error { return m }

// TokenBucketValidationError is the validation error returned by
// TokenBucket.Validate if the designated constraints aren't met.
type TokenBucketValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TokenBucketValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TokenBucketValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TokenBucketValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TokenBucketValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TokenBucketValidationError) ErrorName() string { return "TokenBucketValidationError" }

// Error satisfies the builtin error interface
func (e TokenBucketValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTokenBucket.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TokenBucketValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TokenBucketValidationError{}
