// Package coreerr define os erros sentinelas da biblioteca.
//
// Centraliza todas as definições de erros conhecidos (timeout, input inválido,
// cancelamento, erros de exportação) para que consumidores da API
// possam realizar comparações via errors.Is() com segurança.
//
// Main exports:
// - ErrInvalidInput: Format errors in IP or configurations.
// - ErrTimeout: Errors indicating that an operation timed out.
// - ErrCancelled: Indicates that the scan context was cancelled.
// - ErrExport: Errors during data serialization.
package coreerr
