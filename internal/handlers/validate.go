package handlers

import (
	"bytes"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/italia/publiccode-validator-api/internal/common"
	publiccodeParser "github.com/publiccodeyml/libpubliccode/v5"
)

type PubliccodeymlValidatorHandler struct {
	parser               *publiccodeParser.Parser
	parserExternalChecks *publiccodeParser.Parser
}

func NewPubliccodeymlValidatorHandler() *PubliccodeymlValidatorHandler {
	parser, err := publiccodeParser.NewParser(publiccodeParser.ParserConfig{DisableExternalChecks: true})
	if err != nil {
		panic("can't create a publiccode.yml parser: " + err.Error())
	}

	parserExternalChecks, err := publiccodeParser.NewDefaultParser()
	if err != nil {
		panic("can't create a publiccode.yml parser: " + err.Error())
	}

	return &PubliccodeymlValidatorHandler{parser: parser, parserExternalChecks: parserExternalChecks}
}

func (vh *PubliccodeymlValidatorHandler) Query(ctx fiber.Ctx) error {
	var normalized *string

	valid := true

	parser, invalidExternalChecks := vh.parserForExternalChecks(ctx)
	if invalidExternalChecks {
		return common.Error(
			fiber.StatusBadRequest,
			"invalid external-checks parameter",
			"external-checks must be true or false",
		)
	}

	if len(ctx.Body()) == 0 {
		return common.Error(fiber.StatusBadRequest, "empty body", "need a body to validate")
	}

	results := make(publiccodeParser.ValidationResults, 0)

	reader := bytes.NewReader(ctx.Body())

	parsed, err := parser.ParseStream(reader)
	if err != nil {
		var validationResults publiccodeParser.ValidationResults
		if errors.As(err, &validationResults) {
			var validationError publiccodeParser.ValidationError
			for _, res := range validationResults {
				if errors.As(res, &validationError) {
					valid = false
				}

				results = append(results, res)
			}
		}
	}

	if valid && parsed != nil {
		yaml, err := parsed.ToYAML()
		if err == nil {
			s := string(yaml)
			normalized = &s
		}
	}

	//nolint:wrapcheck
	return ctx.JSON(fiber.Map{"valid": valid, "results": results, "normalized": normalized})
}

func (vh *PubliccodeymlValidatorHandler) parserForExternalChecks(ctx fiber.Ctx) (*publiccodeParser.Parser, bool) {
	args := ctx.Request().URI().QueryArgs()
	if !args.Has("external-checks") {
		return vh.parser, false
	}

	switch string(args.Peek("external-checks")) {
	case "true":
		return vh.parserExternalChecks, false
	case "false":
		return vh.parser, false
	default:
		return nil, true
	}
}
