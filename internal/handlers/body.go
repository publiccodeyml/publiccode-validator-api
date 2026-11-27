package handlers

import (
	"errors"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/italia/publiccode-validator-api/internal/common"
	"github.com/valyala/fasthttp"
)

func contentEncodings(ctx fiber.Ctx) []string {
	var encodings []string

	for _, value := range ctx.Request().Header.PeekAll(fiber.HeaderContentEncoding) {
		for encoding := range strings.SplitSeq(string(value), ",") {
			if encoding = strings.ToLower(strings.TrimSpace(encoding)); encoding != "" {
				encodings = append(encodings, encoding)
			}
		}
	}

	return encodings
}

func validationBody(ctx fiber.Ctx) ([]byte, error) {
	encodings := contentEncodings(ctx)
	if len(encodings) > 5 {
		return nil, fiber.ErrUnsupportedMediaType
	}

	body := ctx.BodyRaw()

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)

	for _, encoding := range slices.Backward(encodings) {
		switch encoding {
		case "identity":
			continue
		case "x-gzip":
			encoding = "gzip"
		case "brotli":
			encoding = "br"
		case "compress", "x-compress":
			return nil, fiber.ErrNotImplemented
		}

		request.SetBodyRaw(body)
		request.Header.SetContentEncoding(encoding)

		decoded, err := request.BodyUncompressedWithLimit(ctx.App().Config().BodyLimit)
		if err != nil {
			return nil, decodingError(err)
		}

		body = decoded
	}

	if len(body) == 0 {
		return nil, common.ProblemJSONError{
			Code: "", Status: fiber.StatusBadRequest, Title: "empty body", Detail: "need a body to validate",
		}
	}

	return body, nil
}

func decodingError(err error) error {
	switch {
	case errors.Is(err, fasthttp.ErrBodyTooLarge):
		return fiber.ErrRequestEntityTooLarge
	case errors.Is(err, fasthttp.ErrContentEncodingUnsupported):
		return fiber.ErrUnsupportedMediaType
	default:
		return common.ProblemJSONError{
			Code: "", Status: fiber.StatusBadRequest, Title: "invalid compressed body", Detail: "cannot decode request body",
		}
	}
}
