// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: envoy/api/v2/core/grpc_service.proto

package envoy_api_v2_core

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"
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
	_ = ptypes.DynamicAny{}
)

// Validate checks the field values on GrpcService with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *GrpcService) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetTimeout()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GrpcServiceValidationError{
				field:  "Timeout",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	for idx, item := range m.GetInitialMetadata() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcServiceValidationError{
					field:  fmt.Sprintf("InitialMetadata[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	switch m.TargetSpecifier.(type) {

	case *GrpcService_EnvoyGrpc_:

		if v, ok := interface{}(m.GetEnvoyGrpc()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcServiceValidationError{
					field:  "EnvoyGrpc",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_:

		if v, ok := interface{}(m.GetGoogleGrpc()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcServiceValidationError{
					field:  "GoogleGrpc",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		return GrpcServiceValidationError{
			field:  "TargetSpecifier",
			reason: "value is required",
		}

	}

	return nil
}

// GrpcServiceValidationError is the validation error returned by
// GrpcService.Validate if the designated constraints aren't met.
type GrpcServiceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcServiceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcServiceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GrpcServiceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcServiceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcServiceValidationError) ErrorName() string { return "GrpcServiceValidationError" }

// Error satisfies the builtin error interface
func (e GrpcServiceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcServiceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcServiceValidationError{}

// Validate checks the field values on GrpcService_EnvoyGrpc with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *GrpcService_EnvoyGrpc) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetClusterName()) < 1 {
		return GrpcService_EnvoyGrpcValidationError{
			field:  "ClusterName",
			reason: "value length must be at least 1 bytes",
		}
	}

	return nil
}

// GrpcService_EnvoyGrpcValidationError is the validation error returned by
// GrpcService_EnvoyGrpc.Validate if the designated constraints aren't met.
type GrpcService_EnvoyGrpcValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_EnvoyGrpcValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcService_EnvoyGrpcValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GrpcService_EnvoyGrpcValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcService_EnvoyGrpcValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_EnvoyGrpcValidationError) ErrorName() string {
	return "GrpcService_EnvoyGrpcValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_EnvoyGrpcValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_EnvoyGrpc.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_EnvoyGrpcValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_EnvoyGrpcValidationError{}

// Validate checks the field values on GrpcService_GoogleGrpc with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *GrpcService_GoogleGrpc) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetTargetUri()) < 1 {
		return GrpcService_GoogleGrpcValidationError{
			field:  "TargetUri",
			reason: "value length must be at least 1 bytes",
		}
	}

	if v, ok := interface{}(m.GetChannelCredentials()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GrpcService_GoogleGrpcValidationError{
				field:  "ChannelCredentials",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	for idx, item := range m.GetCallCredentials() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpcValidationError{
					field:  fmt.Sprintf("CallCredentials[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(m.GetStatPrefix()) < 1 {
		return GrpcService_GoogleGrpcValidationError{
			field:  "StatPrefix",
			reason: "value length must be at least 1 bytes",
		}
	}

	// no validation rules for CredentialsFactoryName

	if v, ok := interface{}(m.GetConfig()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GrpcService_GoogleGrpcValidationError{
				field:  "Config",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// GrpcService_GoogleGrpcValidationError is the validation error returned by
// GrpcService_GoogleGrpc.Validate if the designated constraints aren't met.
type GrpcService_GoogleGrpcValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpcValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcService_GoogleGrpcValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GrpcService_GoogleGrpcValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcService_GoogleGrpcValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_GoogleGrpcValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpcValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpcValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpcValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpcValidationError{}

// Validate checks the field values on GrpcService_GoogleGrpc_SslCredentials
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *GrpcService_GoogleGrpc_SslCredentials) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetRootCerts()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GrpcService_GoogleGrpc_SslCredentialsValidationError{
				field:  "RootCerts",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetPrivateKey()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GrpcService_GoogleGrpc_SslCredentialsValidationError{
				field:  "PrivateKey",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetCertChain()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GrpcService_GoogleGrpc_SslCredentialsValidationError{
				field:  "CertChain",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// GrpcService_GoogleGrpc_SslCredentialsValidationError is the validation error
// returned by GrpcService_GoogleGrpc_SslCredentials.Validate if the
// designated constraints aren't met.
type GrpcService_GoogleGrpc_SslCredentialsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_SslCredentialsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_SslCredentialsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_SslCredentialsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcService_GoogleGrpc_SslCredentialsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_SslCredentialsValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_SslCredentialsValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_SslCredentialsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_SslCredentials.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_SslCredentialsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_SslCredentialsValidationError{}

// Validate checks the field values on
// GrpcService_GoogleGrpc_GoogleLocalCredentials with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *GrpcService_GoogleGrpc_GoogleLocalCredentials) Validate() error {
	if m == nil {
		return nil
	}

	return nil
}

// GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError is the
// validation error returned by
// GrpcService_GoogleGrpc_GoogleLocalCredentials.Validate if the designated
// constraints aren't met.
type GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_GoogleLocalCredentials.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_GoogleLocalCredentialsValidationError{}

// Validate checks the field values on
// GrpcService_GoogleGrpc_ChannelCredentials with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *GrpcService_GoogleGrpc_ChannelCredentials) Validate() error {
	if m == nil {
		return nil
	}

	switch m.CredentialSpecifier.(type) {

	case *GrpcService_GoogleGrpc_ChannelCredentials_SslCredentials:

		if v, ok := interface{}(m.GetSslCredentials()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_ChannelCredentialsValidationError{
					field:  "SslCredentials",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_ChannelCredentials_GoogleDefault:

		if v, ok := interface{}(m.GetGoogleDefault()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_ChannelCredentialsValidationError{
					field:  "GoogleDefault",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_ChannelCredentials_LocalCredentials:

		if v, ok := interface{}(m.GetLocalCredentials()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_ChannelCredentialsValidationError{
					field:  "LocalCredentials",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		return GrpcService_GoogleGrpc_ChannelCredentialsValidationError{
			field:  "CredentialSpecifier",
			reason: "value is required",
		}

	}

	return nil
}

// GrpcService_GoogleGrpc_ChannelCredentialsValidationError is the validation
// error returned by GrpcService_GoogleGrpc_ChannelCredentials.Validate if the
// designated constraints aren't met.
type GrpcService_GoogleGrpc_ChannelCredentialsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_ChannelCredentialsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_ChannelCredentialsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_ChannelCredentialsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcService_GoogleGrpc_ChannelCredentialsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_ChannelCredentialsValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_ChannelCredentialsValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_ChannelCredentialsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_ChannelCredentials.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_ChannelCredentialsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_ChannelCredentialsValidationError{}

// Validate checks the field values on GrpcService_GoogleGrpc_CallCredentials
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *GrpcService_GoogleGrpc_CallCredentials) Validate() error {
	if m == nil {
		return nil
	}

	switch m.CredentialSpecifier.(type) {

	case *GrpcService_GoogleGrpc_CallCredentials_AccessToken:
		// no validation rules for AccessToken

	case *GrpcService_GoogleGrpc_CallCredentials_GoogleComputeEngine:

		if v, ok := interface{}(m.GetGoogleComputeEngine()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentialsValidationError{
					field:  "GoogleComputeEngine",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_CallCredentials_GoogleRefreshToken:
		// no validation rules for GoogleRefreshToken

	case *GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJwtAccess:

		if v, ok := interface{}(m.GetServiceAccountJwtAccess()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentialsValidationError{
					field:  "ServiceAccountJwtAccess",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_CallCredentials_GoogleIam:

		if v, ok := interface{}(m.GetGoogleIam()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentialsValidationError{
					field:  "GoogleIam",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_CallCredentials_FromPlugin:

		if v, ok := interface{}(m.GetFromPlugin()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentialsValidationError{
					field:  "FromPlugin",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_CallCredentials_StsService_:

		if v, ok := interface{}(m.GetStsService()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentialsValidationError{
					field:  "StsService",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		return GrpcService_GoogleGrpc_CallCredentialsValidationError{
			field:  "CredentialSpecifier",
			reason: "value is required",
		}

	}

	return nil
}

// GrpcService_GoogleGrpc_CallCredentialsValidationError is the validation
// error returned by GrpcService_GoogleGrpc_CallCredentials.Validate if the
// designated constraints aren't met.
type GrpcService_GoogleGrpc_CallCredentialsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_CallCredentialsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_CallCredentialsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_CallCredentialsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GrpcService_GoogleGrpc_CallCredentialsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_CallCredentialsValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_CallCredentialsValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_CallCredentialsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_CallCredentials.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_CallCredentialsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_CallCredentialsValidationError{}

// Validate checks the field values on
// GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentials
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentials) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for JsonKey

	// no validation rules for TokenLifetimeSeconds

	return nil
}

// GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError
// is the validation error returned by
// GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentials.Validate
// if the designated constraints aren't met.
type GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError) Field() string {
	return e.field
}

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError) Cause() error {
	return e.cause
}

// Key function returns key value.
func (e GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError) Key() bool {
	return e.key
}

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentials.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_CallCredentials_ServiceAccountJWTAccessCredentialsValidationError{}

// Validate checks the field values on
// GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentials with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentials) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for AuthorizationToken

	// no validation rules for AuthoritySelector

	return nil
}

// GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError
// is the validation error returned by
// GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentials.Validate if the
// designated constraints aren't met.
type GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError) Field() string {
	return e.field
}

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError) Cause() error {
	return e.cause
}

// Key function returns key value.
func (e GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError) Key() bool {
	return e.key
}

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentials.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_CallCredentials_GoogleIAMCredentialsValidationError{}

// Validate checks the field values on
// GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin with
// the rules defined in the proto definition for this message. If any rules
// are violated, an error is returned.
func (m *GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Name

	switch m.ConfigType.(type) {

	case *GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_Config:

		if v, ok := interface{}(m.GetConfig()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError{
					field:  "Config",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig:

		if v, ok := interface{}(m.GetTypedConfig()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError{
					field:  "TypedConfig",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError
// is the validation error returned by
// GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin.Validate
// if the designated constraints aren't met.
type GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError) Field() string {
	return e.field
}

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError) Cause() error {
	return e.cause
}

// Key function returns key value.
func (e GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError) Key() bool {
	return e.key
}

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPluginValidationError{}

// Validate checks the field values on
// GrpcService_GoogleGrpc_CallCredentials_StsService with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *GrpcService_GoogleGrpc_CallCredentials_StsService) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for TokenExchangeServiceUri

	// no validation rules for Resource

	// no validation rules for Audience

	// no validation rules for Scope

	// no validation rules for RequestedTokenType

	if len(m.GetSubjectTokenPath()) < 1 {
		return GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError{
			field:  "SubjectTokenPath",
			reason: "value length must be at least 1 bytes",
		}
	}

	if len(m.GetSubjectTokenType()) < 1 {
		return GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError{
			field:  "SubjectTokenType",
			reason: "value length must be at least 1 bytes",
		}
	}

	// no validation rules for ActorTokenPath

	// no validation rules for ActorTokenType

	return nil
}

// GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError is the
// validation error returned by
// GrpcService_GoogleGrpc_CallCredentials_StsService.Validate if the
// designated constraints aren't met.
type GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError) Field() string {
	return e.field
}

// Reason function returns reason value.
func (e GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError) Cause() error {
	return e.cause
}

// Key function returns key value.
func (e GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError) ErrorName() string {
	return "GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError"
}

// Error satisfies the builtin error interface
func (e GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGrpcService_GoogleGrpc_CallCredentials_StsService.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GrpcService_GoogleGrpc_CallCredentials_StsServiceValidationError{}
