package xconfig

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/dv-net/xconfig/flat"
	"github.com/dv-net/xconfig/plugins"
)

const usageTag = "usage"

func init() {
	plugins.RegisterTag(usageTag)
}

// Usage prints out the current config fields, flags, env vars
// and any other source and setting.
func (c *config) Usage() (string, error) {
	setUsageMeta(c.fields)
	headers := getHeaders(c.fields)

	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "\nSupported Fields:\n")
	fmt.Fprintln(w, strings.ToUpper(strings.Join(headers, "\t")))

	dashes := make([]string, len(headers))
	for i, f := range headers {
		n := len(f)
		if n < 5 {
			n = 5
		}
		dashes[i] = strings.Repeat("-", n)
	}
	fmt.Fprintln(w, strings.Join(dashes, "\t"))

	for _, f := range c.fields {
		if !f.FieldType().IsExported() {
			continue
		}

		values := make([]string, len(headers))
		values[0] = f.Name()
		for i, header := range headers[1:] {
			value := f.Meta()[header]

			if header == "default" && f.FieldValue().CanInterface() {
				value = fmt.Sprintf("%v", f.FieldValue().Interface())
				if _, ok := f.Tag("secret"); ok {
					value = ""
				}
			}

			if header == "secret" {
				if _, ok := f.Tag("secret"); ok {
					value = "✅"
				}
			}

			values[i+1] = value
		}

		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	if err := w.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func setUsageMeta(fs flat.Fields) {
	for _, f := range fs {
		usage, ok := f.Tag(usageTag)
		if !ok {
			continue
		}

		f.Meta()[usageTag] = usage
	}
}

func getHeaders(fs flat.Fields) []string {
	tagMap := map[string]struct{}{}

	for _, f := range fs {
		for key := range f.Meta() {
			tagMap[key] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagMap)+2)

	tags = append(tags, "field")

	for key := range tagMap {
		tags = append(tags, key)
	}

	weights := map[string]int{
		"field": 1,
		"usage": 99,
		"flag":  3,
		"env":   4,
	}

	weight := func(tags []string, i int) int {
		key := tags[i]
		w, ok := weights[key]
		if !ok {
			return 98
		}
		return w
	}

	sort.SliceStable(tags, func(i, j int) bool {
		iw := weight(tags, i)
		jw := weight(tags, j)

		if iw == jw {
			return tags[i] < tags[j]
		}

		return iw < jw
	})

	return tags
}
