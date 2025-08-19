// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var (
	//go:embed templates/*
	templateFS embed.FS
	templates  = template.Must(
		template.New("templates").
			Funcs(funcs).
			Funcs(sprig.FuncMap()).
			ParseFS(templateFS, "templates/*"),
	)
)

var funcs = template.FuncMap{
	"quote": func(v any) string {
		switch v := v.(type) {
		case string:
			return fmt.Sprintf("%q", v)
		case int:
			return fmt.Sprintf("%q", strconv.Itoa(v))
		case bool:
			return fmt.Sprintf("%q", strconv.FormatBool(v))
		}
		return fmt.Sprintf("%q", fmt.Sprintf("%v", v))
	},
}

func ExecTmpl(name string, data any) []byte {
	buf := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(buf, name, data)
	if err != nil {
		logger.Error("failed to execute template", "name", name, "error", err)
		os.Exit(1)
	}
	return buf.Bytes()
}
