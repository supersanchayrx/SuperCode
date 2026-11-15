package tools

import (
	"fmt"
	"strings"
)

const (
	diffContext     = 2
	maxPreviewLines = 40
)

func diffPreview(before, after string) string {
	a := splitLines(before)
	b := splitLines(after)

	start := 0
	for start < len(a) && start < len(b) && a[start] == b[start] {
		start++
	}

	endA, endB := len(a), len(b)
	for endA > start && endB > start && a[endA-1] == b[endB-1] {
		endA--
		endB--
	}

	if start == endA && start == endB {
		return "(no changes)"
	}

	var lines []string

	for i := max(0, start-diffContext); i < start; i++ {
		lines = append(lines, "  "+a[i])
	}
	for i := start; i < endA; i++ {
		lines = append(lines, "- "+a[i])
	}
	for i := start; i < endB; i++ {
		lines = append(lines, "+ "+b[i])
	}
	for i := endA; i < min(len(a), endA+diffContext); i++ {
		lines = append(lines, "  "+a[i])
	}

	if len(lines) > maxPreviewLines {
		hidden := len(lines) - maxPreviewLines
		lines = append(lines[:maxPreviewLines], fmt.Sprintf("... (%d more lines)", hidden))
	}

	header := fmt.Sprintf("@@ line %d: -%d +%d @@", start+1, endA-start, endB-start)
	return header + "\n" + strings.Join(lines, "\n")
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")

	return strings.Split(s, "\n")
}
