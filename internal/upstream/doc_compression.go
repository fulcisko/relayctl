// Package upstream provides balancing, routing, and transport
// utilities for relayctl's reverse proxy.
//
// # Compression
//
// The Compression middleware transparently compresses HTTP responses
// using gzip when the client sends an Accept-Encoding header that
// includes "gzip".
//
// Responses smaller than MinLength bytes are passed through unchanged.
// The Content-Length header is removed when compression is applied
// because the encoded body length differs from the original.
//
// Usage:
//
//	handler = upstream.Compression(upstream.CompressionOptions{
//	    MinLength: 1024,
//	})(handler)
//
// The default MinLength is 512 bytes.
package upstream
