package j_name

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	"koding/remoteapi/models"
)

// NewPostRemoteAPIJNameOneParams creates a new PostRemoteAPIJNameOneParams object
// with the default values initialized.
func NewPostRemoteAPIJNameOneParams() *PostRemoteAPIJNameOneParams {
	var ()
	return &PostRemoteAPIJNameOneParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewPostRemoteAPIJNameOneParamsWithTimeout creates a new PostRemoteAPIJNameOneParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPostRemoteAPIJNameOneParamsWithTimeout(timeout time.Duration) *PostRemoteAPIJNameOneParams {
	var ()
	return &PostRemoteAPIJNameOneParams{

		timeout: timeout,
	}
}

// NewPostRemoteAPIJNameOneParamsWithContext creates a new PostRemoteAPIJNameOneParams object
// with the default values initialized, and the ability to set a context for a request
func NewPostRemoteAPIJNameOneParamsWithContext(ctx context.Context) *PostRemoteAPIJNameOneParams {
	var ()
	return &PostRemoteAPIJNameOneParams{

		Context: ctx,
	}
}

/*PostRemoteAPIJNameOneParams contains all the parameters to send to the API endpoint
for the post remote API j name one operation typically these are written to a http.Request
*/
type PostRemoteAPIJNameOneParams struct {

	/*Body
	  body of the request

	*/
	Body models.DefaultSelector

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the post remote API j name one params
func (o *PostRemoteAPIJNameOneParams) WithTimeout(timeout time.Duration) *PostRemoteAPIJNameOneParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the post remote API j name one params
func (o *PostRemoteAPIJNameOneParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the post remote API j name one params
func (o *PostRemoteAPIJNameOneParams) WithContext(ctx context.Context) *PostRemoteAPIJNameOneParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the post remote API j name one params
func (o *PostRemoteAPIJNameOneParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithBody adds the body to the post remote API j name one params
func (o *PostRemoteAPIJNameOneParams) WithBody(body models.DefaultSelector) *PostRemoteAPIJNameOneParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the post remote API j name one params
func (o *PostRemoteAPIJNameOneParams) SetBody(body models.DefaultSelector) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *PostRemoteAPIJNameOneParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	r.SetTimeout(o.timeout)
	var res []error

	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
