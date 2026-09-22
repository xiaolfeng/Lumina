package mcp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var resultFieldPriority = map[string]int{
	"status":    0,
	"message":   1,
	"state":     2,
	"next_tool": 3,
}

func labeledTextResult(title string, fields map[string]any) *mcp.CallToolResult {
	return textResult(formatLabeledText(title, fields))
}

func formatLabeledText(title string, fields map[string]any) string {
	var builder strings.Builder
	builder.WriteString("[")
	builder.WriteString(title)
	builder.WriteString("]\n")
	writeLabeledMap(&builder, fields, 0)
	return strings.TrimRight(builder.String(), "\n")
}

func writeLabeledMap(builder *strings.Builder, fields map[string]any, indent int) {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		leftPriority, leftKnown := resultFieldPriority[keys[i]]
		rightPriority, rightKnown := resultFieldPriority[keys[j]]
		if leftKnown != rightKnown {
			return leftKnown
		}
		if leftKnown && leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return keys[i] < keys[j]
	})

	for _, key := range keys {
		writeLabeledValue(builder, key, fields[key], indent)
	}
}

func writeLabeledValue(builder *strings.Builder, key string, value any, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch typed := value.(type) {
	case nil:
		fmt.Fprintf(builder, "%s%s: （无）\n", prefix, key)
	case string:
		if strings.Contains(typed, "\n") {
			fmt.Fprintf(builder, "%s%s:\n%s--- BEGIN %s ---\n%s\n%s--- END %s ---\n", prefix, key, prefix, key, typed, prefix, key)
			return
		}
		if typed == "" {
			typed = "（无）"
		}
		fmt.Fprintf(builder, "%s%s: %s\n", prefix, key, typed)
	case map[string]any:
		fmt.Fprintf(builder, "%s%s:\n", prefix, key)
		writeLabeledMap(builder, typed, indent+1)
	case []string:
		fmt.Fprintf(builder, "%s%s:\n", prefix, key)
		if len(typed) == 0 {
			fmt.Fprintf(builder, "%s  （无）\n", prefix)
			return
		}
		for _, item := range typed {
			fmt.Fprintf(builder, "%s  - %s\n", prefix, item)
		}
	case []map[string]any:
		fmt.Fprintf(builder, "%s%s:\n", prefix, key)
		if len(typed) == 0 {
			fmt.Fprintf(builder, "%s  （无）\n", prefix)
			return
		}
		for index, item := range typed {
			fmt.Fprintf(builder, "%s  - item: %d\n", prefix, index+1)
			writeLabeledMap(builder, item, indent+2)
		}
	case []any:
		fmt.Fprintf(builder, "%s%s:\n", prefix, key)
		if len(typed) == 0 {
			fmt.Fprintf(builder, "%s  （无）\n", prefix)
			return
		}
		for index, item := range typed {
			if object, ok := item.(map[string]any); ok {
				fmt.Fprintf(builder, "%s  - item: %d\n", prefix, index+1)
				writeLabeledMap(builder, object, indent+2)
				continue
			}
			fmt.Fprintf(builder, "%s  - %v\n", prefix, item)
		}
	default:
		fmt.Fprintf(builder, "%s%s: %v\n", prefix, key, typed)
	}
}
