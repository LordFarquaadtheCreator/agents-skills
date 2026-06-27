package validate

import (
	"fmt"
	"regexp"
	"strings"
)

func URL(s string) error {
	matched, _ := regexp.MatchString(`^https?://`, s)
	if !matched {
		return fmt.Errorf("link must be a valid URL starting with http:// or https://")
	}
	return nil
}

func Email(s string) error {
	if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
		return fmt.Errorf("email must be a valid email address")
	}
	return nil
}

func Phone(s string) error {
	digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(s, "")
	if len(digits) < 10 || len(digits) > 15 {
		return fmt.Errorf("phone number must be 10-15 digits")
	}
	return nil
}
