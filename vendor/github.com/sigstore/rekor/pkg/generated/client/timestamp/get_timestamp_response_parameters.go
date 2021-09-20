// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package timestamp

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetTimestampResponseParams creates a new GetTimestampResponseParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetTimestampResponseParams() *GetTimestampResponseParams {
	return &GetTimestampResponseParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetTimestampResponseParamsWithTimeout creates a new GetTimestampResponseParams object
// with the ability to set a timeout on a request.
func NewGetTimestampResponseParamsWithTimeout(timeout time.Duration) *GetTimestampResponseParams {
	return &GetTimestampResponseParams{
		timeout: timeout,
	}
}

// NewGetTimestampResponseParamsWithContext creates a new GetTimestampResponseParams object
// with the ability to set a context for a request.
func NewGetTimestampResponseParamsWithContext(ctx context.Context) *GetTimestampResponseParams {
	return &GetTimestampResponseParams{
		Context: ctx,
	}
}

// NewGetTimestampResponseParamsWithHTTPClient creates a new GetTimestampResponseParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetTimestampResponseParamsWithHTTPClient(client *http.Client) *GetTimestampResponseParams {
	return &GetTimestampResponseParams{
		HTTPClient: client,
	}
}

/* GetTimestampResponseParams contains all the parameters to send to the API endpoint
   for the get timestamp response operation.

   Typically these are written to a http.Request.
*/
type GetTimestampResponseParams struct {

	// Request.
	//
	// Format: binary
	Request io.ReadCloser

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get timestamp response params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTimestampResponseParams) WithDefaults() *GetTimestampResponseParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get timestamp response params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTimestampResponseParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get timestamp response params
func (o *GetTimestampResponseParams) WithTimeout(timeout time.Duration) *GetTimestampResponseParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get timestamp response params
func (o *GetTimestampResponseParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get timestamp response params
func (o *GetTimestampResponseParams) WithContext(ctx context.Context) *GetTimestampResponseParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get timestamp response params
func (o *GetTimestampResponseParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get timestamp response params
func (o *GetTimestampResponseParams) WithHTTPClient(client *http.Client) *GetTimestampResponseParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get timestamp response params
func (o *GetTimestampResponseParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithRequest adds the request to the get timestamp response params
func (o *GetTimestampResponseParams) WithRequest(request io.ReadCloser) *GetTimestampResponseParams {
	o.SetRequest(request)
	return o
}

// SetRequest adds the request to the get timestamp response params
func (o *GetTimestampResponseParams) SetRequest(request io.ReadCloser) {
	o.Request = request
}

// WriteToRequest writes these params to a swagger request
func (o *GetTimestampResponseParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Request != nil {
		if err := r.SetBodyParam(o.Request); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
