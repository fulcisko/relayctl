// Package upstream — warmup handler
//
// The warmup handler exposes HTTP endpoints for managing backend warmup state.
//
// GET  /admin/warmup          — list all registered backends and their warmup weight
// PUT  /admin/warmup          — register or update a backend warmup config
// DELETE /admin/warmup?backend=<url> — remove a backend from warmup tracking
//
// Example PUT body:
//
//	{
//	  "backend": "http://10.0.0.1:8080",
//	  "duration_seconds": 60,
//	  "start_weight": 0.1
//	}
package upstream
