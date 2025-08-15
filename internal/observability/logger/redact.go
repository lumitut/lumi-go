// Package logger provides structured logging with Zap
package logger

import (
	"regexp"
	"strings"
)

// Common PII patterns
var (
	// Email pattern
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	// Credit card pattern (basic - matches 13-19 digit numbers)
	creditCardRegex = regexp.MustCompile(`\b\d{13,19}\b`)
	// SSN pattern (US)
	ssnRegex = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	// Phone number pattern (various formats)
	phoneRegex = regexp.MustCompile(`\b(\+?\d{1,3}[-.\s]?)?\(?\d{1,4}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}\b`)
	// IP Address pattern
	ipRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	// JWT pattern
	jwtRegex = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`)
	// API Key patterns (common formats)
	apiKeyRegex = regexp.MustCompile(`(api[_-]?key|apikey|api_secret|secret[_-]?key)[\s:="']+(\S+)`)
	// Password in JSON/URL
	passwordRegex = regexp.MustCompile(`(password|passwd|pwd)[\s:="']+(\S+)`)
)

// RedactOption configures PII redaction
type RedactOption struct {
	Emails       bool
	CreditCards  bool
	SSNs         bool
	Phones       bool
	IPs          bool
	JWTs         bool
	APIKeys      bool
	Passwords    bool
	CustomRegex  *regexp.Regexp
	CustomRedact string
}

// DefaultRedactOptions returns default redaction options
func DefaultRedactOptions() RedactOption {
	return RedactOption{
		Emails:      true,
		CreditCards: true,
		SSNs:        true,
		Phones:      true,
		IPs:         false, // IPs might be needed for debugging
		JWTs:        true,
		APIKeys:     true,
		Passwords:   true,
	}
}

// RedactPII redacts personally identifiable information from a string
func RedactPII(input string, opts RedactOption) string {
	result := input

	if opts.Emails {
		result = emailRegex.ReplaceAllString(result, "[REDACTED_EMAIL]")
	}
	if opts.CreditCards {
		result = creditCardRegex.ReplaceAllStringFunc(result, func(match string) string {
			// Check if it looks like a credit card (Luhn check would be better)
			if len(match) >= 13 && len(match) <= 19 {
				return "[REDACTED_CC]" 
			}
			return match
		})
	}
	if opts.SSNs {
		result = ssnRegex.ReplaceAllString(result, "[REDACTED_SSN]")
	}
	if opts.Phones {
		result = phoneRegex.ReplaceAllStringFunc(result, func(match string) string {
			// Only redact if it looks like a phone number (has enough digits)
			digits := strings.Count(match, "0") + strings.Count(match, "1") +
				strings.Count(match, "2") + strings.Count(match, "3") +
				strings.Count(match, "4") + strings.Count(match, "5") +
				strings.Count(match, "6") + strings.Count(match, "7") +
				strings.Count(match, "8") + strings.Count(match, "9")
			if digits >= 7 {
				return "[REDACTED_PHONE]"
			}
			return match
		})
	}
	if opts.IPs {
		result = ipRegex.ReplaceAllString(result, "[REDACTED_IP]")
	}
	if opts.JWTs {
		result = jwtRegex.ReplaceAllString(result, "[REDACTED_JWT]")
	}
	if opts.APIKeys {
		result = apiKeyRegex.ReplaceAllString(result, "$1=[REDACTED_API_KEY]")
	}
	if opts.Passwords {
		result = passwordRegex.ReplaceAllString(result, "$1=[REDACTED_PASSWORD]")
	}
	if opts.CustomRegex != nil {
		redactValue := opts.CustomRedact
		if redactValue == "" {
			redactValue = "[REDACTED]"
		}
		result = opts.CustomRegex.ReplaceAllString(result, redactValue)
	}

	return result
}

// RedactMap redacts PII from a map of strings
func RedactMap(input map[string]string, opts RedactOption) map[string]string {
	result := make(map[string]string, len(input))
	for k, v := range input {
		// Redact values
		result[k] = RedactPII(v, opts)
		
		// Also redact sensitive keys entirely
		lowerKey := strings.ToLower(k)
		if strings.Contains(lowerKey, "password") ||
			strings.Contains(lowerKey, "secret") ||
			strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "api_key") ||
			strings.Contains(lowerKey, "apikey") {
			result[k] = "[REDACTED]"
		}
	}
	return result
}

// RedactHeaders redacts sensitive HTTP headers
func RedactHeaders(headers map[string][]string) map[string][]string {
	result := make(map[string][]string, len(headers))
	opts := DefaultRedactOptions()
	
	sensitiveHeaders := map[string]bool{
		"authorization":  true,
		"cookie":         true,
		"set-cookie":     true,
		"x-api-key":      true,
		"x-auth-token":   true,
		"x-csrf-token":   true,
		"x-access-token": true,
	}
	
	for k, values := range headers {
		lowerKey := strings.ToLower(k)
		if sensitiveHeaders[lowerKey] {
			// Completely redact sensitive headers
			result[k] = []string{"[REDACTED]"}
		} else {
			// Redact PII from other headers
			redactedValues := make([]string, len(values))
			for i, v := range values {
				redactedValues[i] = RedactPII(v, opts)
			}
			result[k] = redactedValues
		}
	}
	return result
}

// RedactJSON redacts PII from JSON strings
// This is a simple implementation - for production use consider a proper JSON parser
func RedactJSON(jsonStr string, opts RedactOption) string {
	// First apply general PII redaction
	result := RedactPII(jsonStr, opts)
	
	// Then handle JSON-specific sensitive fields
	sensitiveFields := []string{
		"password", "secret", "token", "api_key", "apiKey",
		"private_key", "privateKey", "credit_card", "creditCard",
		"ssn", "social_security", "socialSecurity",
	}
	
	for _, field := range sensitiveFields {
		// Handle both quoted and unquoted field names
		patterns := []string{
			`"` + field + `"\s*:\s*"[^"]*"`,
			`'` + field + `'\s*:\s*'[^']*'`,
			`"` + field + `"\s*:\s*[^,}\s]+`,
		}
		
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			result = re.ReplaceAllStringFunc(result, func(match string) string {
				parts := strings.SplitN(match, ":", 2)
				if len(parts) == 2 {
					return parts[0] + `: "[REDACTED]"`
				}
				return match
			})
		}
	}
	
	return result
}
