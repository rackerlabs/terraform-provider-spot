package spotvalidator

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Float64 = decimalDigitsAtMostValidator{}

type decimalDigitsAtMostValidator struct {
	digitsAtMost int
}

func (validator decimalDigitsAtMostValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must have up to %d decimal digits", validator.digitsAtMost)
}

func (validator decimalDigitsAtMostValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator decimalDigitsAtMostValidator) ValidateFloat64(ctx context.Context, request validator.Float64Request, response *validator.Float64Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueFloat64()

	if decDigits := DecimalDigitsCount(value); decDigits > validator.digitsAtMost {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			strconv.Itoa(decDigits),
		))
	}
}

func IsFloatHasUptoNDecimalDigits(f float64, n int) bool {
	scale := math.Pow10(n)
	return f == math.Trunc(f*scale)/scale
}

func DecimalDigitsCount(f float64) int {
	str := strconv.FormatFloat(f, 'f', -1, 64)
	str = strings.TrimRight(str, "0")
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return 0
	}
	return len(parts[1])
}

func DecimalDigitsAtMost(digitsAtMost int) validator.Float64 {
	if digitsAtMost < 0 {
		return nil
	}
	return decimalDigitsAtMostValidator{
		digitsAtMost: digitsAtMost,
	}
}
