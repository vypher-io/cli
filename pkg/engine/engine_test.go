package engine

import (
	"testing"
)

func TestLuhnValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "Valid Visa", input: "4111111111111111", want: true},
		{name: "Valid Visa with spaces", input: "4111 1111 1111 1111", want: true},
		{name: "Valid Visa with dashes", input: "4111-1111-1111-1111", want: true},
		{name: "Valid Mastercard", input: "5500000000000004", want: true},
		{name: "Invalid card number", input: "1234567890123456", want: false},
		{name: "All zeros", input: "0000000000000000", want: true},
		{name: "Single digit", input: "5", want: false},
		{name: "Empty string", input: "", want: false},
		{name: "Valid Amex", input: "378282246310005", want: true},
		{name: "Invalid modified Visa", input: "4111111111111112", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LuhnValid(tt.input); got != tt.want {
				t.Errorf("LuhnValid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestScanContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
		wantNil bool
	}{
		{
			name:    "No Match",
			content: "This is a clean string with no secrets.",
			wantNil: true,
		},
		{
			name:    "Credit Card (valid Luhn)",
			content: "Here is a credit card: 4111 1111 1111 1111",
			wantLen: 1,
		},
		{
			name:    "Credit Card (invalid Luhn, filtered out)",
			content: "Not a real card: 1234 5678 9012 3456",
			wantNil: true,
		},
		{
			name:    "SSN",
			content: "My SSN is 123-45-6789.",
			wantLen: 1,
		},
		{
			name:    "Email",
			content: "Contact me at test@example.com",
			wantLen: 1,
		},
		{
			name:    "Multiple Matches",
			content: "Email: test@example.com, SSN: 987-65-4321",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScanContent(tt.content)
			if tt.wantNil {
				if len(got) != 0 {
					t.Errorf("ScanContent() = %v, want nil/empty", got)
				}
			} else {
				if len(got) != tt.wantLen {
					t.Errorf("ScanContent() returned %d matches, want %d", len(got), tt.wantLen)
				}
			}
		})
	}
}

func TestCryptoAddresses(t *testing.T) {
	// Bitcoin P2PKH address
	btcContent := "BTC wallet: 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	results := ScanContent(btcContent)
	foundBTC := false
	for _, m := range results {
		if m.RuleName == "Bitcoin Address" {
			foundBTC = true
			if !m.KeywordProximity {
				t.Error("Bitcoin match with 'wallet' keyword should have KeywordProximity=true")
			}
		}
	}
	if !foundBTC {
		t.Error("Expected to find Bitcoin Address match")
	}

	// Ethereum address
	ethContent := "ETH: 0x742d35Cc6634C0532925a3b844Bc9e7595f2bD3e"
	results = ScanContent(ethContent)
	foundETH := false
	for _, m := range results {
		if m.RuleName == "Ethereum Address" {
			foundETH = true
		}
	}
	if !foundETH {
		t.Error("Expected to find Ethereum Address match")
	}

	// Tag filtering — crypto tag should find crypto rules
	cryptoResults := ScanContentWithTags(btcContent, []string{"crypto"})
	foundCrypto := false
	for _, m := range cryptoResults {
		if m.RuleName == "Bitcoin Address" {
			foundCrypto = true
		}
	}
	if !foundCrypto {
		t.Error("Expected crypto tag filter to find Bitcoin Address")
	}
}

func TestScanContentWithTags(t *testing.T) {
	content := "Email: test@example.com, SSN: 987-65-4321"

	// Filter by "finance" tag - should only match SSN
	financeResults := ScanContentWithTags(content, []string{"finance"})
	if len(financeResults) != 1 {
		t.Errorf("ScanContentWithTags(finance) returned %d matches, want 1", len(financeResults))
	} else if financeResults[0].RuleName != "SSN" {
		t.Errorf("ScanContentWithTags(finance) match = %s, want SSN", financeResults[0].RuleName)
	}

	// Filter by "communication" tag - should only match Email
	commResults := ScanContentWithTags(content, []string{"communication"})
	if len(commResults) != 1 {
		t.Errorf("ScanContentWithTags(communication) returned %d matches, want 1", len(commResults))
	} else if commResults[0].RuleName != "Email" {
		t.Errorf("ScanContentWithTags(communication) match = %s, want Email", commResults[0].RuleName)
	}

	// Filter by "pii" tag - should match both SSN and Email
	piiResults := ScanContentWithTags(content, []string{"pii"})
	if len(piiResults) != 2 {
		t.Errorf("ScanContentWithTags(pii) returned %d matches, want 2", len(piiResults))
	}

	// No tags - should match all
	allResults := ScanContentWithTags(content, nil)
	if len(allResults) != 2 {
		t.Errorf("ScanContentWithTags(nil) returned %d matches, want 2", len(allResults))
	}

	// Non-existent tag - should match nothing
	emptyResults := ScanContentWithTags(content, []string{"nonexistent"})
	if len(emptyResults) != 0 {
		t.Errorf("ScanContentWithTags(nonexistent) returned %d matches, want 0", len(emptyResults))
	}
}

func TestKeywordProximity(t *testing.T) {
	// SSN with keyword nearby — should have KeywordProximity=true
	withKeyword := "My SSN is 123-45-6789 on file."
	results := ScanContent(withKeyword)
	foundSSN := false
	for _, m := range results {
		if m.RuleName == "SSN" {
			foundSSN = true
			if !m.KeywordProximity {
				t.Error("SSN match with 'SSN' keyword nearby should have KeywordProximity=true")
			}
		}
	}
	if !foundSSN {
		t.Error("Expected to find SSN match")
	}

	// SSN without any keywords nearby — should have KeywordProximity=false
	// Use content where the SSN appears far from any proximity keywords
	withoutKeyword := "The number is 123-45-6789 listed in the database."
	results = ScanContent(withoutKeyword)
	foundSSN = false
	for _, m := range results {
		if m.RuleName == "SSN" {
			foundSSN = true
			if m.KeywordProximity {
				t.Error("SSN match without proximity keywords should have KeywordProximity=false")
			}
		}
	}
	if !foundSSN {
		t.Error("Expected to find SSN match")
	}

	// Credit Card with keyword nearby
	ccWithKeyword := "Credit card number: 4111 1111 1111 1111"
	results = ScanContent(ccWithKeyword)
	foundCC := false
	for _, m := range results {
		if m.RuleName == "Credit Card" {
			foundCC = true
			if !m.KeywordProximity {
				t.Error("Credit Card match with 'credit card' keyword nearby should have KeywordProximity=true")
			}
			if !m.ValidatedByLuhn {
				t.Error("Credit Card match should have ValidatedByLuhn=true")
			}
		}
	}
	if !foundCC {
		t.Error("Expected to find Credit Card match")
	}

	// Email — no proximity keywords defined, should always be false
	emailContent := "Email: test@example.com"
	results = ScanContent(emailContent)
	for _, m := range results {
		if m.RuleName == "Email" && m.KeywordProximity {
			t.Error("Email match should not have KeywordProximity (no proximity keywords defined)")
		}
	}
}

// matchesEqual checks if two slices of Match are equal.
func matchesEqual(a, b []Match) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
