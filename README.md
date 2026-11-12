# publiccode.yml RESTful validator API

A RESTful API for validating [publiccode.yml](https://github.com/publiccodeyml/publiccode.yml)
files, using the [libpubliccode](https://github.com/publiccodeyml/libpubliccode)
library.

## API reference

The OpenAPI spec is in [`publiccode-validator-api.oas.yaml`](publiccode-validator-api.oas.yaml).

[Browse the interactive Swagger UI preview](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/publiccodeyml/publiccode-validator-api/main/publiccode-validator-api.oas.yaml)

## Usage

```console
go run main.go
```

### Validate a publiccode.yml file

```console
curl -X POST "http://localhost:3000/v1/validate" --data-binary "@./publiccode.yml"
```

The endpoint also accepts the `QUERY` method:

```console
curl -X QUERY "http://localhost:3000/v1/validate" \
--data-binary $'publiccodeYmlVersion: "0.5"\ndevelopmentStatus: stable\n [...]\n'
```

### Run external validation checks

External URL reachability checks are disabled by default. Enable them when
needed with the `external-checks` query parameter:

```console
curl -X POST "http://localhost:3000/v1/validate?external-checks=true" \
--data-binary "@./publiccode.yml"
```

This mode makes network requests for the external checks, so it can take
longer than the default validation.

### Example response (valid publiccode.yml)

For [`testdata/valid.publiccode.yml`](testdata/valid.publiccode.yml):

```json
{
  "valid": true,
  "normalized": "publiccodeYmlVersion: \"0\"\nname: Medusa\nurl: https://github.com/example/project.git\nreleaseDate: null\nplatforms:\n- web\ncategories:\n- cloud-management\ndevelopmentStatus: development\nsoftwareType: standalone/other\ndescription:\n  en-GB:\n    localisedName: Medusa\n    genericName: \"\"\n    shortDescription: |\n      A rather short description which is probably useless\n    longDescription: |\n      Very long description of this software, also split on multiple rows. You should note what the software is and why one should need it. This is 158 characters. Very long description of this software, also split on multiple rows. You should note what the software is and why one should need it. This is 316 characters. Very long description of this software, also split on multiple rows. You should note what the software is and why one should need it. This is 474 characters. Very long description of this software, also split on multiple rows. You should note what the software is and why one should need it. This is 632 characters.\n    features:\n    - Just one feature\nlegal:\n  license: AGPL-3.0-or-later\nmaintenance:\n  type: community\n  contacts:\n  - name: Francesco Rossi\nlocalisation:\n  localisationReady: true\n  availableLanguages:\n  - en\n",
  "results": []
}
```

### Example response (with errors / warnings)

For the same file with `publiccodeYmlVersion: "0.2"` and `legal.license: invalid`:

```json
{
  "valid": false,
  "normalized": null,
  "results": [
    {
      "file": "",
      "key": "publiccodeYmlVersion",
      "description": "v0.2 is not the latest version, use '0'. Parsing this file as v0.8.",
      "line": 1,
      "column": 1,
      "type": "warning"
    },
    {
      "file": "",
      "key": "legal.license",
      "description": "license must be a valid license (see https://spdx.org/licenses)",
      "line": 39,
      "column": 3,
      "type": "error"
    }
  ]
}
```

### Example error response

All error responses are returned using the `application/problem+json` media type,
in accordance with [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html).

```http
HTTP/1.1 400 Bad Request
Content-Type: application/problem+json
```

```json
{
  "title": "empty body",
  "detail": "need a body to validate",
  "status": 400
}
```

## Contributing

Contributing is always appreciated, see [CONTRIBUTING.md](CONTRIBUTING.md).
Feel free to open issues, fork or submit a Pull Request.

## Maintainers

This software is maintained by community maintainers.

## License

Licensed under AGPL-3.0.
The version control system provides attribution for specific lines of code.
