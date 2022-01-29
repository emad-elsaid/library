package main

import (
	"fmt"
	"html/template"
)

func Helpers() {
	helpers["partial"] = func(path string, data interface{}) (template.HTML, error) {
		return template.HTML(partial(path, data)), nil
	}

	helpers["meta_property"] = func(meta map[string]string, name string) string {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		return fmt.Sprintf(`<meta property="%s" value="%s"/>`, template.HTMLEscapeString(name), template.HTMLEscapeString(v))
	}

	helpers["meta_name"] = func(meta map[string]string, name string) string {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		return fmt.Sprintf(`<meta name="%s" value="%s"/>`, name, v)
	}

	helpers["can"] = func(verb string, record interface{}) bool {
		return true
	}
}
