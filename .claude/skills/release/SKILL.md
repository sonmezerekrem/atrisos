---
name: release
description: Create a new Atrisos release — bumps version, writes release notes from commits, creates an annotated git tag, and pushes it to trigger the goreleaser GitHub Actions workflow.
user-invocable: true
allowed-tools:
  - Bash(git *)
  - Bash(gh *)
  - AskUserQuestion
---

# Atrisos Release

You are releasing the Atrisos project. A `v*` tag push triggers the
`release.yml` GitHub Actions workflow, which runs goreleaser and publishes
four binaries to GitHub Releases. Your job is to determine the correct next
version, write good release notes, create the annotated tag, and push it.

## Step 1 — Pre-flight checks

Run all of these in parallel:

```sh
git status --porcelain          # must be clean
git rev-parse --abbrev-ref HEAD # must be main
git fetch --tags --quiet
git tag --sort=-v:refname | head -1   # latest tag = current version
git log $(git describe --tags --abbrev=0)..HEAD --oneline   # unreleased commits
git rev-list HEAD ^$(git describe --tags --abbrev=0) --count # commit count
```

If working tree is dirty or branch is not `main`, stop and tell the user
what needs to be fixed. Do not proceed.

## Step 2 — Determine the next version

Parse the latest tag (e.g. `v0.2.2`) into MAJOR.MINOR.PATCH.

Analyze the unreleased commits and apply these rules:
- Any commit that breaks the public CLI interface → **major** bump
- Any commit with `feat:` prefix, or that adds a new command, flag, or user-visible feature → **minor** bump
- Fixes, docs, refactors, CI changes, chore → **patch** bump

Show the user:
- Current version
- Unreleased commit count
- Your recommended bump type and resulting version

If the user passed an explicit version as an argument (e.g. `/release v0.3.0`),
use that instead and skip the bump analysis.

Ask the user to confirm or choose a different bump:

```
What version should this release be?
- options: patch (v0.X.Y+1), minor (v0.X+1.0), major (v1.0.0), or custom
```

## Step 3 — Write release notes

Categorize the unreleased commits into sections. Use the full commit subject
(not just the prefix). Skip merge commits and commits with subjects like
"wip", "tmp", or "fixup!".

Section order and rules:
1. **Features** — new user-visible functionality (`feat:`, new command/flag, new template)
2. **Bug Fixes** — defect corrections (`fix:`)
3. **Documentation** — docs, README, agent prompts (`docs:`)
4. **Improvements** — refactors, performance, UX polish (`refactor:`, `chore:`, `style:`)
5. **CI / Build** — workflow, goreleaser, Makefile (`ci:`, `build:`)

If a section has no entries, omit it.

Write in sentence case. Strip conventional commit prefixes (`feat:`, `fix:`,
etc.) from subjects — they become the section heading instead.

Format:

```
## What's Changed

### Features
- Added postgres and valkey templates with auto-generated passwords

### Bug Fixes
- Fixed spurious systemd errors on `atrisos down` when backup is disabled

### Documentation
- Redesigned README with badges and cleaner layout for new users

**Full changelog**: https://github.com/sonmezerekrem/atrisos/compare/vPREV...vNEW
```

Show the release notes to the user and ask for confirmation or edits before
proceeding. This is a good time to make wording adjustments.

## Step 4 — Create and push the tag

After the user confirms:

1. Create an annotated tag with the release notes as the tag message:
   ```sh
   git tag -a vNEW -m "$(cat <<'EOF'
   Release vNEW

   <release notes here>
   EOF
   )"
   ```

2. Push the tag:
   ```sh
   git push origin vNEW
   ```

3. Tell the user: the `release.yml` GitHub Actions workflow is now running.
   They can watch it at:
   `https://github.com/sonmezerekrem/atrisos/actions/workflows/release.yml`

4. Once goreleaser completes (usually 2–3 minutes), update the GitHub release
   body with the formatted release notes using:
   ```sh
   gh release edit vNEW --notes "<release notes>"
   ```
   Goreleaser auto-generates notes from commits; this replaces them with the
   curated version.

## Notes

- Never force-push tags or re-tag an existing version.
- The binary names goreleaser produces are:
  `atrisos-linux-amd64`, `atrisos-linux-arm64`,
  `atrisos-darwin-amd64`, `atrisos-darwin-arm64`
- The install script uses the GitHub releases API to find the latest tag —
  publishing the release makes the new version available to `self-update`.
- If goreleaser fails, delete the tag locally and remotely before retrying:
  ```sh
  git tag -d vNEW && git push origin :refs/tags/vNEW
  ```
