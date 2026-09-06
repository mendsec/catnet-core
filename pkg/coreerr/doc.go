// Package coreerr defines the sentinel errors of the library.
//
// Centralizes all definitions of known errors (timeout, invalid input,
// cancellation, export errors) so that API consumers
// can safely perform comparisons via errors.Is().
//
// Main exports:
// - ErrInvalidInput: Format errors in IP or configurations.
// - ErrTimeout: Errors indicating that an operation timed out.
// - ErrCancelled: Indicates that the scan context was cancelled.
// - ErrExport: Errors during data serialization.
package coreerr
