# Development Tasks

- [x] DDD Architecture Refactoring
- [x] Resolve Python 3.13 compatibility with `pyflp` library
- [x] VS Code Debug Configuration
- [x] Fix `pyflp` API usage in `flp_resolver.py`

## Compatibility & Architecture Resolved
- The Python 3.13 `EventEnum` issue has been resolved by patching the `pyflp` library's metaclass to handle empty Enums and optimize member lookup.
- The project is now properly structured following DDD principles.
- VS Code is configured to launch the Go server from the correct entry point.
- The FLP resolver script now correctly uses the `pyflp` API for mixer track and slot iteration.
