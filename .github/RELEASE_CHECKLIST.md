# Release Checklist

## Preparation

1. Check tests, linters.
2. Check issues and pull requests, update milestones and labels.
3. Check if there are breaking changes or anything that should be reflected in the changelog and/or user documentation.
4. Create draft release on GitHub with autogenerated release notes.
5. Generate changelog:
   - Run `task changelog MILESTONE_TITLE=vX.Y.Z PREVIOUS_MILESTONE_TITLE=vX.Y.Z` to generate release changelog for the given milestone.
   - Copy the changelog to the `CHANGELOG.md` file.
   - Sort items within sections according to importance if needed.
   - Add first-time contributor credits if any (copy this data from the release draft).
6. Run `task docs-version VERSION=X.Y`.
   Update `versions` in `docusaurus.config.js`.
   Remove the oldest version from `versioned_docs`, `versioned_sidebars`, `versions.json`.
7. Run `task docs-fmt`.
8. Commit and push changes to the PR.
9. Merge PR, pull and check `git status`.
10. Create a release of [FerretDB/documentdb](https://github.com/FerretDB/documentdb) for this FerretDB version.

## Git tag

1. Make a signed tag `vX.Y.Z` with the relevant section of the changelog using `--cleanup=verbatim`.
2. Check `task gen-version; git status` output.
3. Push it!
4. Refresh
   - `env GOPROXY=https://proxy.golang.org go mod download -x github.com/FerretDB/FerretDB/v2@<vX.Y.Z>`
   - `https://pkg.go.dev/github.com/FerretDB/FerretDB/v2@<vX.Y.Z>` from https://pkg.go.dev/github.com/FerretDB/FerretDB/v2?tab=versions.

## Release

1. Copy release notes from `CHANGELOG.md` and trim them.
2. Wait for the [packages CI build](https://github.com/FerretDB/FerretDB/actions/workflows/packages.yml?query=event%3Apush)
   to finish.
3. Upload binaries and packages to the draft release.
4. Check:
   - https://hub.docker.com/r/ferretdb/ferretdb/tags
   - https://hub.docker.com/r/ferretdb/ferretdb-dev/tags
   - https://hub.docker.com/r/ferretdb/ferretdb-eval/tags
   - https://github.com/FerretDB/FerretDB/pkgs/container/ferretdb
   - https://github.com/FerretDB/FerretDB/pkgs/container/ferretdb-dev
   - https://github.com/FerretDB/FerretDB/pkgs/container/ferretdb-eval
   - https://quay.io/repository/ferretdb/ferretdb?tab=tags
   - https://quay.io/repository/ferretdb/ferretdb-dev?tab=tags
   - https://quay.io/repository/ferretdb/ferretdb-eval?tab=tags
5. Upload DocumentDB `.deb` packages for this FerretDB version from
   [FerretDB/documentdb releases](https://github.com/FerretDB/documentdb/releases) to the draft release.
6. Close milestone in issues.
7. Publish release on GitHub.
8. Announce it on Slack.

## Soon after

1. Bump the latest version on https://beacon.ferretdb.com and https://beacon.ferretdb.io.
2. Publish and announce blog post.
3. Tweet, toot.
4. Update NixOS package: https://github.com/NixOS/nixpkgs/tree/master/pkgs/servers/nosql/ferretdb.
5. Update Civo package: https://github.com/civo/kubernetes-marketplace/tree/master/ferretdb.
