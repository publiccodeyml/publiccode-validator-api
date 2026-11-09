# Contributing

Thank you for contributing to publiccode-validator-api. We welcome bug
reports, questions, documentation improvements, and code changes.

## Issues

Before opening an issue, please search the [existing
issues](https://github.com/publiccodeyml/publiccode-validator-api/issues) to
avoid duplicates. Use a new issue to report a reproducible problem, suggest an
improvement, or ask a question about this repository.

When reporting a problem, include the command you ran, the expected result,
the actual result, and the version or commit you used when possible.

## Pull requests

1. Fork the repository and create a branch from `main` for your change.
2. Make the change and add or update tests and documentation where relevant.
3. Run the test suite:

   ```sh
   go test -race ./...
   ```

4. Open a pull request against
   [`main`](https://github.com/publiccodeyml/publiccode-validator-api/tree/main).
   Describe the problem it solves, the approach taken, and any related issue.

Keep pull requests focused. Maintainers may ask for changes during review
before merging a contribution.

## License

By contributing, you agree that your work is licensed under the
[AGPL-3.0-or-later](LICENSE).
