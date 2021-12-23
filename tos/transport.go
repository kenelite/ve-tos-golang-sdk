package tos

import (
	"context"
	"net"
	"net/http"
	"time"
)

type TransportConfig struct {

	// MaxIdleConns same as http.Transport MaxIdleConns
	MaxIdleConns int

	// RequestTimeout same as http.Client Timeout
	RequestTimeout time.Duration

	// DialTimeout same as net.Dialer Timeout
	DialTimeout time.Duration

	// KeepAlive same as net.Dialer KeepAlive
	KeepAlive time.Duration

	// IdleConnTimeout same as http.Transport IdleConnTimeout
	IdleConnTimeout time.Duration

	// TLSHandshakeTimeout same as http.Transport TLSHandshakeTimeout
	TLSHandshakeTimeout time.Duration

	// ResponseHeaderTimeout same as http.Transport ResponseHeaderTimeout
	ResponseHeaderTimeout time.Duration

	// ExpectContinueTimeout same as http.Transport ExpectContinueTimeout
	ExpectContinueTimeout time.Duration

	// ReadTimeout see net.Conn SetReadDeadline
	ReadTimeout time.Duration

	// WriteTimeout set net.Conn SetWriteDeadline
	WriteTimeout time.Duration
}

type Transport interface {
	RoundTrip(context.Context, *Request) (*Response, error)
}

type DefaultTransport struct {
	client http.Client
}

// NewDefaultTransport create a DefaultTransport with config
func NewDefaultTransport(config *TransportConfig) *DefaultTransport {
	return &DefaultTransport{
		client: http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   config.DialTimeout,
					KeepAlive: config.KeepAlive,
				}).DialContext,
				//DialContext: (&TimeoutDialer{
				//	Dialer: net.Dialer{
				//		Timeout:   config.DialTimeout,
				//		KeepAlive: config.KeepAlive,
				//	},
				//	ReadTimeout:  config.ReadTimeout,
				//	WriteTimeout: config.WriteTimeout,
				//}).DialContext,
				MaxIdleConns:          config.MaxIdleConns,
				IdleConnTimeout:       config.IdleConnTimeout,
				TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
				ResponseHeaderTimeout: config.ResponseHeaderTimeout,
				ExpectContinueTimeout: config.ExpectContinueTimeout,
				DisableCompression:    true,
			},
		},
	}
}

// NewDefaultTransportWithClient crate a DefaultTransport with a http.Client
func NewDefaultTransportWithClient(client http.Client) *DefaultTransport {
	return &DefaultTransport{client: client}
}

func (dt *DefaultTransport) RoundTrip(ctx context.Context, req *Request) (*Response, error) {
	hr, err := http.NewRequestWithContext(ctx, req.Method, req.URL(), req.Content)
	if err != nil {
		return nil, err
	}

	if req.ContentLength != nil {
		hr.ContentLength = *req.ContentLength
	}

	for key, values := range req.Header {
		hr.Header[key] = values
	}

	res, err := dt.client.Do(hr)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:    res.StatusCode,
		ContentLength: res.ContentLength,
		Header:        res.Header,
		Body:          res.Body,
	}, nil
}

type TimeoutConn struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	zero         time.Time
}

func NewTimeoutConn(conn net.Conn, readTimeout, writeTimeout time.Duration) *TimeoutConn {
	return &TimeoutConn{
		Conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (tc *TimeoutConn) Read(b []byte) (n int, err error) {
	timeout := tc.readTimeout > 0
	if timeout {
		_ = tc.SetReadDeadline(time.Now().Add(tc.readTimeout))
	}

	n, err = tc.Conn.Read(b)
	if timeout {
		_ = tc.SetReadDeadline(tc.zero)
	}
	return n, err
}

func (tc *TimeoutConn) Write(b []byte) (n int, err error) {
	timeout := tc.writeTimeout > 0
	if timeout {
		_ = tc.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
	}

	n, err = tc.Conn.Read(b)
	if timeout {
		_ = tc.SetWriteDeadline(tc.zero)
	}
	return n, err
}

type TimeoutDialer struct {
	net.Dialer
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (d *TimeoutDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewTimeoutConn(conn, d.ReadTimeout, d.WriteTimeout), nil
}
