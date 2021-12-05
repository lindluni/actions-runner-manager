//go:build tools

/**
SPDX-License-Identifier: Apache-2.0
*/

package tools

import (
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "github.com/swaggo/swag/cmd/swag"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
	_ "mvdan.cc/gofumpt"
)
