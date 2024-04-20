# go-sqlite3

This repo is a fork of the original go-sqlite3 repo, which is a sqlite3 driver for go that using cgo.

The original repo is [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3).

## Main changes

- Add missing operator definitions (LIMIT and OFFSET) for `sqlite3.Op` enum.
- The original repo add omit checking for constraints handling of a virtual table, this fork keeps them unless it's an offset constraint. This allows us to prevent a faulty virtual table implementation from returning incorrect results.
- You can't let SQLite choose the entrypoint for a runtime loadable extension because go-sqlite3 pass directly the string. Therefore, it can pass be a null pointer. This fork adds a check if the string is empty and pass a null pointer if it is.

## Using this fork

This fork is mostly made for anyquery. The behaviour of this fork is different from the original repo, and change are not backward compatible. Therefore, I don't recommend using this fork in your project unless you know what you are doing.

## License

99.9% of the code is under the same license as the original repo. The changes made by this fork are under the same license as the original repo.
I want to thank mattn for his work on the original repo and the SQLite team for their work on SQLite.
