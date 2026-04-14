package engine

import (
	"regexp"
	"strings"
	"unicode"
)

// Rule defines a PII/PHI detection rule
type Rule struct {
	Name              string
	Description       string
	Regex             *regexp.Regexp
	Tags              []string // e.g., "finance", "healthcare", "pii"
	ProximityKeywords []string // optional keywords that boost confidence when found near a match
}

// Match represents a found PII/PHI instance
type Match struct {
	RuleName         string
	Content          string
	Index            int
	Line             int
	ValidatedByLuhn  bool
	KeywordProximity bool // true when a proximity keyword was found near the match
}

var Rules = []Rule{
	{
		Name:              "Credit Card",
		Description:       "Potential Credit Card Number (13-16 digits)",
		Regex:             regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
		Tags:              []string{"finance", "pii"},
		ProximityKeywords: []string{"card", "credit", "visa", "mastercard", "payment", "amex", "discover"},
	},
	{
		Name:              "SSN",
		Description:       "US Social Security Number",
		Regex:             regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		Tags:              []string{"finance", "pii", "government"},
		ProximityKeywords: []string{"ssn", "social", "security", "social security"},
	},
	{
		Name:        "Email",
		Description: "Email Address",
		Regex:       regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		Tags:        []string{"pii", "communication"},
	},
	{
		Name:        "Phone",
		Description: "Phone Number (US/Intl)",
		Regex:       regexp.MustCompile(`\b(?:\+?1[-. ]?)?(\(?([0-9]{3})\)?[-. ]?([0-9]{3})[-. ]?([0-9]{4}))\b`),
		Tags:        []string{"pii", "communication"},
	},
	{
		Name:        "IBAN",
		Description: "International Bank Account Number",
		Regex:       regexp.MustCompile(`\b[A-Z]{2}\d{2}[A-Z0-9]{4}\d{7}([A-Z0-9]?){0,16}\b`),
		Tags:        []string{"finance"},
	},
	{
		Name:        "MRN",
		Description: "Medical Record Number (Generic 6-12 digits)",
		Regex:       regexp.MustCompile(`\b(MRN|Medical Record Number)\s*[:#-]?\s*(\d{6,12})\b`),
		Tags:        []string{"healthcare", "phi"},
	},
	{
		Name:        "DOB",
		Description: "Date of Birth detected near keywords",
		Regex:       regexp.MustCompile(`(?i)\b(dob|birth|born)\s*[:#-]?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})\b`),
		Tags:        []string{"pii", "phi"},
	},
	{
		Name:        "ICD-10",
		Description: "ICD-10 Medical Code",
		Regex:       regexp.MustCompile(`\b[A-TV-Z][0-9][0-9AB](?:\.[0-9A-K]{1,4})?\b`),
		Tags:        []string{"healthcare", "phi"},
	},
	{
		Name:              "Bitcoin Address",
		Description:       "Bitcoin Wallet Address (P2PKH, P2SH, Bech32)",
		Regex:             regexp.MustCompile(`\b([13][a-km-zA-HJ-NP-Z1-9]{25,34}|bc1[a-zA-HJ-NP-Z0-9]{25,90})\b`),
		Tags:              []string{"finance", "crypto"},
		ProximityKeywords: []string{"bitcoin", "btc", "wallet", "crypto", "address"},
	},
	{
		Name:              "Ethereum Address",
		Description:       "Ethereum Wallet Address (0x-prefixed)",
		Regex:             regexp.MustCompile(`\b0x[0-9a-fA-F]{40}\b`),
		Tags:              []string{"finance", "crypto"},
		ProximityKeywords: []string{"ethereum", "eth", "wallet", "crypto", "address", "erc20"},
	},
	{
		Name:              "Solana Address",
		Description:       "Solana Wallet Address (Base58, 32-44 chars)",
		Regex:             regexp.MustCompile(`\b[1-9A-HJ-NP-Za-km-z]{32,44}\b`),
		Tags:              []string{"finance", "crypto"},
		ProximityKeywords: []string{"solana", "sol", "wallet", "crypto", "address"},
	},
}

// LuhnValid checks if a numeric string passes the Luhn algorithm.
// Non-digit characters (spaces, dashes) are stripped before validation.
func LuhnValid(s string) bool {
	var digits []int
	for _, r := range s {
		if unicode.IsDigit(r) {
			digits = append(digits, int(r-'0'))
		}
	}
	if len(digits) < 2 {
		return false
	}
	sum := 0
	nDigits := len(digits)
	parity := nDigits % 2
	for i, d := range digits {
		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// checkKeywordProximity searches for any of the given keywords within ±50 characters
// of the match position in the content. The search is case-insensitive.
func checkKeywordProximity(content string, matchStart, matchEnd int, keywords []string) bool {
	if len(keywords) == 0 {
		return false
	}

	// Define the proximity window (±50 chars from match boundaries)
	windowStart := matchStart - 50
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := matchEnd + 50
	if windowEnd > len(content) {
		windowEnd = len(content)
	}

	window := strings.ToLower(content[windowStart:windowEnd])

	for _, kw := range keywords {
		if strings.Contains(window, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// indexToLine returns the 1-based line number for a byte offset in content.
func indexToLine(content string, index int) int {
	return strings.Count(content[:index], "\n") + 1
}

// ScanContent scans a string content for all defined rules
func ScanContent(content string) []Match {
	return ScanContentWithTags(content, nil)
}

// ScanContentWithTags scans content using only rules matching the given tags.
// If tags is nil or empty, all rules are used.
func ScanContentWithTags(content string, tags []string) []Match {
	var matches []Match

	for _, rule := range Rules {
		if len(tags) > 0 && !ruleMatchesTags(rule, tags) {
			continue
		}

		found := rule.Regex.FindAllStringIndex(content, -1)
		for _, loc := range found {
			matchContent := content[loc[0]:loc[1]]
			proximity := checkKeywordProximity(content, loc[0], loc[1], rule.ProximityKeywords)

			// For Credit Card matches, validate with Luhn to reduce false positives
			if rule.Name == "Credit Card" {
				if !LuhnValid(matchContent) {
					continue // skip false positives
				}
				matches = append(matches, Match{
					RuleName:         rule.Name,
					Content:          matchContent,
					Index:            loc[0],
					Line:             indexToLine(content, loc[0]),
					ValidatedByLuhn:  true,
					KeywordProximity: proximity,
				})
			} else {
				matches = append(matches, Match{
					RuleName:         rule.Name,
					Content:          matchContent,
					Index:            loc[0],
					Line:             indexToLine(content, loc[0]),
					KeywordProximity: proximity,
				})
			}
		}
	}

	return matches
}

// ruleMatchesTags returns true if the rule has at least one tag in common with the filter.
func ruleMatchesTags(rule Rule, tags []string) bool {
	for _, ruleTag := range rule.Tags {
		for _, filterTag := range tags {
			if ruleTag == filterTag {
				return true
			}
		}
	}
	return false
}
