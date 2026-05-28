package automigrate

import (
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

var safeIdentifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func ValidateIdentifier(identifier string) error {
	if !safeIdentifierPattern.MatchString(identifier) {
		return gerror.Newf("unsafe postgres identifier %q", identifier)
	}
	return nil
}

func QuoteIdentifier(identifier string) (string, error) {
	if err := ValidateIdentifier(identifier); err != nil {
		return "", err
	}
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`, nil
}
