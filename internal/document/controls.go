package document

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// Validate checks stored argument values. Required empty inputs are accepted
// here so they can clear dependencies; the scheduler blocks their execution.
func (v Input) Validate(value string) string {
	if strings.ContainsAny(value, "\r\n") || (v.MaxLength > 0 && utf8.RuneCountInString(value) > v.MaxLength) {
		return "参数必须为单行且不得超过 max_length"
	}
	switch v.Type {
	case "toggle":
		if value != "true" && value != "false" {
			return "开关值必须为 true 或 false"
		}
	case "select":
		for _, option := range v.Options {
			if option.Value == value {
				return ""
			}
		}
		return "请选择 options 中的值"
	case "number":
		if value == "" {
			return ""
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return "请输入 64 位有符号整数"
		}
		if v.Min != nil && n < *v.Min {
			return fmt.Sprintf("数值不得小于 %d", *v.Min)
		}
		if v.Max != nil && n > *v.Max {
			return fmt.Sprintf("数值不得大于 %d", *v.Max)
		}
	}
	return ""
}

func (c *compiler) input(b block, id string) Input {
	c.allowed(b, "id label type default required max_length options min max")
	v := Input{ID: id, Label: c.str(b, "label", id, false), Type: c.str(b, "type", "text", false), Required: c.boolean(b, "required", false), MaxLength: c.integer(b, "max_length", 256)}
	if v.Type != "text" && v.Type != "toggle" && v.Type != "select" && v.Type != "number" {
		c.fail(b.line, "type", "使用 text、toggle、select 或 number")
	}
	if v.MaxLength < 1 || v.MaxLength > 4096 {
		c.fail(b.line, "max_length", "取值范围 1～4096")
	}
	for _, key := range []string{"min", "max"} {
		if node, ok := b.fields[key]; ok {
			if v.Type != "number" || node.Kind != yaml.ScalarNode || node.Tag != "!!int" {
				c.fail(b.line+node.Line, key, "仅 number 可设置整数边界")
			}
			n, err := strconv.ParseInt(node.Value, 10, 64)
			if err != nil {
				c.fail(b.line+node.Line, key, "使用 64 位十进制整数")
			}
			if key == "min" {
				v.Min = &n
			} else {
				v.Max = &n
			}
		}
	}
	if v.Min != nil && v.Max != nil && *v.Min > *v.Max {
		c.fail(b.line, "min", "min 不得大于 max")
	}
	if node, ok := b.fields["options"]; ok {
		if v.Type != "select" || node.Kind != yaml.SequenceNode || len(node.Content) == 0 || len(node.Content) > 256 {
			c.fail(b.line+node.Line, "options", "仅 select 可设置 1～256 个选项")
		}
		seen := map[string]bool{}
		for _, option := range node.Content {
			if option.Kind != yaml.MappingNode {
				c.fail(b.line+option.Line, "options", "每项使用 value、label 映射")
			}
			fields := map[string]*yaml.Node{}
			for i := 0; i < len(option.Content); i += 2 {
				fields[option.Content[i].Value] = option.Content[i+1]
			}
			ob := block{line: b.line, fields: fields}
			c.allowed(ob, "value label")
			value := c.str(ob, "value", "", true)
			label := c.str(ob, "label", value, false)
			if value == "" || strings.TrimSpace(label) == "" || strings.ContainsAny(value+label, "\r\n") || utf8.RuneCountInString(value) > v.MaxLength || utf8.RuneCountInString(label) > 256 || seen[value] {
				c.fail(b.line+option.Line, "options", "选项值须唯一且非空；标签须非空、单行且不超过 256 字符；值不得超过 max_length")
			}
			seen[value] = true
			v.Options = append(v.Options, Option{Value: value, Label: label})
		}
	} else if v.Type == "select" {
		c.fail(b.line, "options", "select 必须提供 options")
	}
	if v.Type == "toggle" {
		v.Default = "false"
	}
	if len(v.Options) > 0 {
		v.Default = v.Options[0].Value
	}
	if node, ok := b.fields["default"]; ok {
		validTag := node.Tag == "!!str" || (v.Type == "toggle" && node.Tag == "!!bool") || (v.Type == "number" && node.Tag == "!!int")
		if node.Kind != yaml.ScalarNode || !validTag {
			c.fail(b.line+node.Line, "default", "默认值类型与控件不匹配；文本及选项值请使用字符串")
		}
		v.Default = node.Value
		if v.Type == "toggle" {
			v.Default = strings.ToLower(v.Default)
		}
	}
	if reason := v.Validate(v.Default); reason != "" {
		c.fail(b.line, "default", reason)
	}
	return v
}
