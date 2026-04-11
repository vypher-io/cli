# Vypher CLI

Vypher is a command-line tool designed to scan directories for Personally Identifiable Information (PII) and Protected Health Information (PHI) within files. It focuses on identifying sensitive data related to finance and healthcare.

## Installation

**Via `go install`** (requires Go 1.20+):

```bash
go install github.com/vypher-io/cli@latest
```

**Or build from source:**

```bash
git clone https://github.com/vypher-io/cli.git
cd cli
go build -o vypher
```

**macOS via Homebrew:**

```bash
brew install vypher-io/vypher/vypher
```

## Testing

To run the unit tests for the project:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

To run tests with race condition detection:

```bash
go test -race ./...
```

## Usage

### Scanning a Directory

To scan the current directory:

```bash
./vypher scan
```

To scan a specific directory:

```bash
./vypher scan --target /path/to/scan
```

### Output Formats

Vypher supports multiple output formats: **console** (default), **JSON**, and **SARIF**.

**Console Output:**

```bash
./vypher scan -t ./src
```

**JSON Output:**

```bash
./vypher scan -t ./src -o json
```

**SARIF Output** (for GitHub Security, VS Code integration):

```bash
./vypher scan -t ./src -o sarif
```

### Excluding Files

Use the `--exclude` flag with glob patterns to skip files or directories:

```bash
./vypher scan -t ./src --exclude "*_test.go" --exclude "*.log"
```

### Filtering by Rule Tags

Use `--rules` to scan only for specific rule categories:

```bash
./vypher scan -t ./src --rules finance,phi
```

**Available tags:**

| Tag | Patterns Included |
|-----|-------------------|
| `finance` | Credit Card, SSN, IBAN, Bitcoin, Ethereum, Solana |
| `crypto` | Bitcoin, Ethereum, Solana |
| `pii` | Credit Card, SSN, Email, Phone, DOB |
| `healthcare` | MRN, ICD-10 |
| `phi` | MRN, DOB, ICD-10 |
| `communication` | Email, Phone |
| `government` | SSN |

**Examples:**

```bash
# Scan for finance-only patterns (cards, SSN, IBAN, crypto wallets)
./vypher scan -t ./src --rules finance

# Scan for crypto wallet addresses only
./vypher scan -t ./src --rules crypto

# Scan for healthcare data only (MRN, ICD-10, DOB)
./vypher scan -t ./src --rules healthcare,phi

# Scan for general PII (emails, phones, SSNs, cards, DOB)
./vypher scan -t ./src --rules pii

# Combine multiple tags
./vypher scan -t ./src --rules finance,healthcare -o sarif --fail-on-match
```

### Limiting Scan Depth

Use `--max-depth` to limit how deep the scanner recurses:

```bash
./vypher scan -t ./src --max-depth 3
```

### CI/CD Integration

Use `--fail-on-match` to exit with code 1 when issues are found:

```bash
./vypher scan -t ./src --fail-on-match
```

This is useful for enforcing compliance in CI/CD pipelines.

### Configuration File

Create a `.vypher.yaml` file to define default scan settings:

```yaml
exclude:
  - "*_test.go"
  - "*.log"
rules:
  - finance
  - phi
output: sarif
max_depth: 5
fail_on_match: true
```

Load it with the `--config` flag:

```bash
./vypher scan --config .vypher.yaml -t ./src
```

CLI flags always override config file values.

### Default Ignored Directories

By default, Vypher ignores the following directories:
- `.git`
- `node_modules`
- `vendor`
- `.venv`
- `__pycache__`

## Detected Patterns

Vypher ships with 11 built-in detection patterns:

| # | Pattern | Description | Tags | Validation |
|---|---------|-------------|------|------------|
| 1 | **Credit Card** | 13-16 digit card numbers | `finance`, `pii` | Luhn ✓, Proximity |
| 2 | **SSN** | US Social Security Numbers (XXX-XX-XXXX) | `finance`, `pii`, `government` | Proximity |
| 3 | **Email** | Email addresses | `pii`, `communication` | — |
| 4 | **Phone** | US/International phone numbers | `pii`, `communication` | — |
| 5 | **IBAN** | International Bank Account Numbers | `finance` | — |
| 6 | **MRN** | Medical Record Numbers (6-12 digits) | `healthcare`, `phi` | — |
| 7 | **DOB** | Date of Birth near keywords | `pii`, `phi` | — |
| 8 | **ICD-10** | ICD-10 medical diagnosis codes | `healthcare`, `phi` | — |
| 9 | **Bitcoin** | P2PKH, P2SH, Bech32 wallet addresses | `finance`, `crypto` | Proximity |
| 10 | **Ethereum** | 0x-prefixed 40 hex char addresses | `finance`, `crypto` | Proximity |
| 11 | **Solana** | Base58 32-44 char wallet addresses | `finance`, `crypto` | Proximity |

### Credit Card Validation

Credit card numbers detected by regex are validated using the **Luhn algorithm** to reduce false positives, as recommended by PCI DSS.

### Keyword Proximity

SSN and Credit Card matches are annotated with a **keyword proximity** indicator when relevant keywords (e.g., "ssn", "social", "credit", "card") are found within ±50 characters of the match. This helps distinguish high-confidence detections from potential false positives.

## Performance

Vypher uses **parallel file scanning** with a worker pool automatically sized to the number of CPU cores. File collection (directory walk) is sequential, while file reading and pattern matching run concurrently for maximum throughput on large codebases.

## Disclaimer

This tool uses regex-based pattern matching and may produce false positives. It is intended as an aid for developers and security professionals, not as a guaranteed solution for compliance.
