# Release Process

This project uses [GoReleaser](https://goreleaser.com/) and GitHub Actions for automated releases.

## How to create a new release

1. Update the codebase with your changes
2. Ensure all tests pass: `make test`
3. Commit your changes: `git commit -m "Your commit message"`
4. Tag a new version following semantic versioning:
   ```
   git tag -a v1.0.0 -m "Version 1.0.0"
   ```
5. Push the tag to GitHub:
   ```
   git push origin v1.0.0
   ```

The GitHub Actions workflow will automatically:
- Build binaries for multiple platforms (Linux, macOS, Windows)
- Create a GitHub release with these binaries
- Generate checksums for verification

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- MAJOR version for incompatible API changes
- MINOR version for backward-compatible functionality additions
- PATCH version for backward-compatible bug fixes

## Release Notes

When creating a new tag, consider adding a detailed description of the changes in the tag message or in the GitHub release.