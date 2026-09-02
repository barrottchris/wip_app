package app

import (
	"strconv"
	"strings"
)

// InferBrowseURL derives a likely local browse URL for a component from the
// start command. This is intentionally heuristic: many dev servers listen on a
// known port and a human-friendly URL is easier to act on than reading the raw
// command text. The result is used only as a UI hint and is never treated as a
// hard guarantee that the process is healthy.
func InferBrowseURL(component Component) string {
	cmd := strings.TrimSpace(component.StartCommand)
	if cmd == "" {
		return ""
	}

	patterns := []string{
		"--port ",
		"-p ",
		"-l ",
		"--listen ",
		"--host ",
		"--bind ",
	}

	for _, token := range patterns {
		if idx := strings.Index(cmd, token); idx >= 0 {
			value := strings.TrimSpace(cmd[idx+len(token):])
			if value != "" {
				if portMatch := extractPort(value); portMatch != "" {
					return "http://localhost:" + portMatch
				}
			}
		}
	}

	if idx := strings.Index(cmd, "http.server"); idx >= 0 {
		if portMatch := extractPort(cmd[idx:]); portMatch != "" {
			return "http://127.0.0.1:" + portMatch
		}
	}

	if idx := strings.Index(cmd, "serve"); idx >= 0 {
		if portMatch := extractPort(cmd[idx:]); portMatch != "" {
			return "http://localhost:" + portMatch
		}
	}

	if portMatch := extractPort(cmd); portMatch != "" {
		return "http://localhost:" + portMatch
	}

	return ""
}

// InferBrowseURLFromLogs returns the first HTTP(S) URL found in process output.
func InferBrowseURLFromLogs(logs []string) string {
	for _, line := range logs {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if idx := strings.Index(trimmed, "http://"); idx >= 0 {
			if url, ok := readURLToken(trimmed[idx:]); ok {
				return url
			}
		}
		if idx := strings.Index(trimmed, "https://"); idx >= 0 {
			if url, ok := readURLToken(trimmed[idx:]); ok {
				return url
			}
		}
	}
	return ""
}

// readURLToken trims punctuation and whitespace from a candidate URL token.
func readURLToken(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	for _, end := range []string{" ", "\n", "\t", "\"", "'", ")", "]", "}"} {
		if idx := strings.Index(trimmed, end); idx >= 0 {
			trimmed = trimmed[:idx]
		}
	}
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, true
	}
	return "", false
}

// extractPort scans command text for a numeric port or host-and-port value.
func extractPort(value string) string {
	parts := strings.Fields(value)
	for i, part := range parts {
		if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
			if i+1 < len(parts) {
				if port, ok := parsePort(parts[i+1]); ok {
					return port
				}
			}
		}
		if port, ok := parsePort(part); ok {
			return port
		}
		if strings.Contains(part, ":") {
			if port, ok := parsePort(strings.TrimSpace(strings.Split(part, ":")[1])); ok {
				return port
			}
		}
	}
	return ""
}

// parsePort validates a port token after removing common command-output
// punctuation.
func parsePort(value string) (string, bool) {
	trimmed := strings.Trim(strings.TrimSpace(value), "\"'")
	trimmed = strings.TrimSuffix(trimmed, ",")
	trimmed = strings.TrimSuffix(trimmed, ";")
	trimmed = strings.TrimSuffix(trimmed, ")")
	trimmed = strings.TrimSuffix(trimmed, "]")
	if trimmed == "" {
		return "", false
	}
	if _, err := strconv.Atoi(trimmed); err == nil {
		return trimmed, true
	}
	return "", false
}