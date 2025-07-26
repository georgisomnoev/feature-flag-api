//go:build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/hexdigest/gowrap/cmd/gowrap"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
)
