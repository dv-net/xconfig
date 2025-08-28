package xconfig

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dv-net/xconfig/flat"
	"github.com/dv-net/xconfig/internal/utils"
)

const cellSeparator = "|"

func GenerateMarkdown(cfg any, opts ...Option) (string, error) {
	c, err := Load(cfg, opts...)
	if err != nil {
		return "", err
	}

	fields, err := flat.View(cfg)
	if err != nil {
		return "", err
	}

	var table [][]string //nolint:prealloc

	table = append(table, []string{
		"**Name**", "**Required**", "**Secret**", "**Default value**", "**Usage**", "**Example**",
	})

	sizes := make([]int, len(table[0]))

	var lineSize int
	for i, cell := range table[0] {
		sizes[i] = utf8.RuneCountInString(cell) + 2
	}

	for _, f := range fields {
		// skip if field is not exported
		if !f.FieldType().IsExported() {
			continue
		}

		envName := f.EnvName()
		if c.Options().envPrefix != "" {
			envName = c.Options().envPrefix + "_" + envName
		}

		var isRequired bool
		var isSecret bool
		var defaultValue string
		var usage string
		var example string

		if _, ok := f.Tag("required"); ok {
			isRequired = true
		}

		if !isRequired {
			if val, ok := f.Tag("validate"); ok && strings.Contains(val, "required") {
				isRequired = true
			}
		}

		if _, ok := f.Tag("secret"); ok {
			isSecret = true
		}

		val, err := utils.LookupString(cfg, f.Name())
		if err != nil {
			return "", fmt.Errorf("failed to lookup value for %s: %w", f.Name(), err)
		}

		if val.CanInterface() && !isSecret {
			defaultValue = fmt.Sprintf("%v", val.Interface())
		}

		if val, ok := f.Tag("usage"); ok {
			usage = val
		}

		if val, ok := f.Tag("example"); ok {
			example = val
		}

		cell := []string{
			"`" + envName + "`",
			boolIcon(isRequired),
			boolIcon(isSecret),
			codeBlock(defaultValue),
			usage,
			codeBlock(example),
		}
		table = append(table, cell)

		lineSize = 0
		for i, item := range cell {
			if size := utf8.RuneCountInString(item); size+2 > sizes[i] {
				sizes[i] = size + 2
			}

			lineSize += sizes[i] // recalculate line size
		}
	}

	var out strings.Builder
	for i, row := range table {
		_, _ = out.WriteString(cellSeparator)

		for j, cell := range row {
			size := utf8.RuneCountInString(" " + cell + " ")

			data := strings.Repeat(" ", sizes[j]-size)

			_, _ = out.WriteString(" " + cell + " ")
			_, _ = out.WriteString(data)

			if len(row)-1 != j {
				_, _ = out.WriteString(cellSeparator)
			}
		}

		if i == 0 {
			_, _ = out.WriteString(cellSeparator)
			_, _ = out.WriteRune('\n')

			_, _ = out.WriteString(cellSeparator)
			for j, item := range sizes {
				dashes := strings.Repeat("-", item)
				_, _ = out.WriteString(dashes)

				if len(sizes)-1 != j {
					_, _ = out.WriteString(cellSeparator)
				}
			}
		}

		_, _ = out.WriteString(cellSeparator)
		_, _ = out.WriteRune('\n')
	}

	return strings.TrimSpace(out.String()), nil
}

func boolIcon(value bool) string {
	if value {
		return "✅"
	}

	return " "
}

func codeBlock(val string) string {
	if val == "" {
		return val
	}

	return "`" + val + "`"
}
