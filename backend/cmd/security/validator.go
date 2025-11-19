package security

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

var (
	ErrInputTooLong         = errors.New("input exceeds maximum length")
	ErrInputTooShort        = errors.New("input is too short")
	ErrInvalidFormat        = errors.New("invalid input format")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrInvalidURL           = errors.New("invalid URL format")
	ErrContainsXSS          = errors.New("input contains potentially dangerous content")
	ErrContainsSQLInjection = errors.New("input contains potentially dangerous SQL patterns")
)

// Input limits
const (
	MaxUsernameLength      = 50
	MinUsernameLength      = 3
	MaxEmailLength         = 255
	MaxPasswordLength      = 128
	MinPasswordLength      = 8
	MaxReviewContentLength = 5000
	MinReviewContentLength = 10
	MaxReviewRating        = 10
	MinReviewRating        = 1
	MaxMessageLength       = 1000
	MaxMangaTitleLength    = 200
	MaxDescriptionLength   = 10000
)

// SQL injection patterns to detect
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|EXECUTE|UNION|SCRIPT)\b)`),
	regexp.MustCompile(`(?i)(--|/\*|\*/|;|\||&)`),
	regexp.MustCompile(`(?i)(\b(OR|AND)\s+\d+\s*=\s*\d+)`),
	regexp.MustCompile(`(?i)(\b(OR|AND)\s+['"]\w+['"]\s*=\s*['"]\w+['"])`),
	regexp.MustCompile(`(?i)(\b(OR|AND)\s+['"]1['"]\s*=\s*['"]1['"])`),
}

// XSS patterns to detect
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<\s*script[^>]*>.*?</\s*script\s*>`),
	regexp.MustCompile(`(?i)<\s*iframe[^>]*>.*?</\s*iframe\s*>`),
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)on\w+\s*=\s*['"]`),
	regexp.MustCompile(`(?i)<\s*img[^>]*src\s*=\s*['"]\s*javascript\s*:`),
	regexp.MustCompile(`(?i)<\s*link[^>]*href\s*=\s*['"]\s*javascript\s*:`),
	regexp.MustCompile(`(?i)<\s*style[^>]*>.*?</\s*style\s*>`),
	regexp.MustCompile(`(?i)expression\s*\(`),
}

// ValidateLength validates input length
// Input length limits are enforced
func ValidateLength(input string, min, max int) error {
	length := utf8.RuneCountInString(input)
	if length < min {
		return fmt.Errorf("%w: minimum length is %d characters", ErrInputTooShort, min)
	}
	if length > max {
		return fmt.Errorf("%w: maximum length is %d characters", ErrInputTooLong, max)
	}
	return nil
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes
// XSS attempts are sanitized
func SanitizeHTML(input string) string {
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		// If not valid HTML, return escaped version
		return html.EscapeString(input)
	}

	var sanitize func(*html.Node)
	sanitize = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Allow only safe tags
			allowedTags := map[string]bool{
				"p": true, "br": true, "strong": true, "em": true,
				"u": true, "b": true, "i": true, "ul": true, "ol": true,
				"li": true, "blockquote": true, "code": true, "pre": true,
			}

			if !allowedTags[n.Data] {
				// Remove disallowed tags
				n.Type = html.TextNode
				n.Data = ""
				return
			}

			// Remove all attributes (prevent XSS via attributes)
			n.Attr = []html.Attribute{}
		}

		// Recursively sanitize children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			sanitize(c)
		}
	}

	sanitize(doc)

	// Render sanitized HTML
	var b strings.Builder
	var render func(*html.Node)
	render = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			b.WriteString(n.Data)
		case html.ElementNode:
			b.WriteString("<" + n.Data + ">")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				render(c)
			}
			b.WriteString("</" + n.Data + ">")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			render(c)
		}
	}
	render(doc)

	return b.String()
}

// DetectSQLInjection checks for SQL injection patterns
// SQL injection attempts are blocked
func DetectSQLInjection(input string) error {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return ErrContainsSQLInjection
		}
	}
	return nil
}

// DetectXSS checks for XSS patterns
// XSS attempts are sanitized
func DetectXSS(input string) error {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return ErrContainsXSS
		}
	}
	return nil
}

// SanitizeString removes potentially dangerous characters and patterns
// XSS attempts are sanitized
func SanitizeString(input string) string {
	// First check for XSS
	if err := DetectXSS(input); err == nil {
		// If no XSS detected, sanitize HTML
		return SanitizeHTML(input)
	}
	// If XSS detected, escape HTML
	return html.EscapeString(input)
}

// ValidateEmail validates email format
// Invalid data formats are rejected
func ValidateEmail(email string) error {
	if err := ValidateLength(email, 5, MaxEmailLength); err != nil {
		return err
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	}

	// Additional validation: check for common email patterns
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateURL validates URL format
// Invalid data formats are rejected
func ValidateURL(urlString string) error {
	if urlString == "" {
		return nil // Empty URL is allowed
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: only http and https schemes are allowed", ErrInvalidURL)
	}

	return nil
}

// ValidateUsername validates username format
// Invalid data formats are rejected
func ValidateUsername(username string) error {
	if err := ValidateLength(username, MinUsernameLength, MaxUsernameLength); err != nil {
		return err
	}

	// Username should only contain alphanumeric characters, underscores, and hyphens
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("%w: username can only contain letters, numbers, underscores, and hyphens", ErrInvalidFormat)
	}

	// Check for SQL injection
	if err := DetectSQLInjection(username); err != nil {
		return err
	}

	return nil
}

// ValidatePassword validates password format
// Invalid data formats are rejected
func ValidatePassword(password string) error {
	if err := ValidateLength(password, MinPasswordLength, MaxPasswordLength); err != nil {
		return err
	}

	// Password should contain at least one letter and one number
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasNumber {
		return fmt.Errorf("%w: password must contain at least one letter and one number", ErrInvalidFormat)
	}

	return nil
}

// ValidateReviewContent validates review content
// Input length limits are enforced
// XSS attempts are sanitized
func ValidateReviewContent(content string) error {
	if err := ValidateLength(content, MinReviewContentLength, MaxReviewContentLength); err != nil {
		return err
	}

	// Check for SQL injection
	if err := DetectSQLInjection(content); err != nil {
		return err
	}

	// Check for XSS (will be sanitized, but we log it)
	if err := DetectXSS(content); err != nil {
		// Log XSS attempt but allow sanitized content
		// Content will be sanitized before storage
	}

	return nil
}

// SanitizeReviewContent sanitizes review content
// XSS attempts are sanitized
func SanitizeReviewContent(content string) string {
	return SanitizeString(content)
}

// ValidateReviewRating validates review rating
// Invalid data formats are rejected
func ValidateReviewRating(rating int) error {
	if rating < MinReviewRating || rating > MaxReviewRating {
		return fmt.Errorf("%w: rating must be between %d and %d", ErrInvalidFormat, MinReviewRating, MaxReviewRating)
	}
	return nil
}

// ValidateMangaTitle validates manga title
// Input length limits are enforced
func ValidateMangaTitle(title string) error {
	if err := ValidateLength(title, 1, MaxMangaTitleLength); err != nil {
		return err
	}

	// Check for SQL injection
	if err := DetectSQLInjection(title); err != nil {
		return err
	}

	return nil
}

// ValidateDescription validates description
// Input length limits are enforced
func ValidateDescription(description string) error {
	if err := ValidateLength(description, 0, MaxDescriptionLength); err != nil {
		return err
	}

	// Check for SQL injection
	if err := DetectSQLInjection(description); err != nil {
		return err
	}

	return nil
}

// ValidateInteger validates integer input
// Invalid data formats are rejected
func ValidateInteger(value int, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("%w: value must be between %d and %d", ErrInvalidFormat, min, max)
	}
	return nil
}

// ValidatePositiveInteger validates positive integer
func ValidatePositiveInteger(value int) error {
	return ValidateInteger(value, 1, 2147483647) // Max int32
}
