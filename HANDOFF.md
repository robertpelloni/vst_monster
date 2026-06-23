# Session Handoff Notes

## Context & Key Findings
- Addressed code review feedback about tracked `node_modules` and fake crawler features.
- Explicitly untracked `node_modules` folders to prevent bloated diffs.
- Fixed the Go crawler (`crawler/main.go`) to no longer fetch from `vst-monster.com`. Due to websites like KVR Audio blocking bots with 403 Forbidden statuses, the crawler was repointed to Wikipedia's list of music software to test proper data extraction without being blocked.
- Fixed the crawler's `CalculateSHA256` function to actually download the binary content to a temp file on disk, correctly verify the SHA-256 against the downloaded data, and safely clean up the temp file.
- The `rust` application requires linux native gtk packages (`libwebkit2gtk-4.1-dev`, `libglib2.0-dev`) to compile which are not provided in `Cargo.toml`. These were installed during this session via `apt-get` and allow the Tauri back-end to successfully `cargo check` and `cargo test`.
- All `pre-commit` checklist tasks and reviews complete and passing.

## Next Steps
- Implement and link the Go Crawler to actually deploy against the real live postgres instances instead of localhost.
- Proceed to implement more UI components in the Tauri application that utilize the Express back-end logic.
- Integrate github submodules and upstream repositories in the sync process.
