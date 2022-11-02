// Package hash provides functions for hashing values from strings.
//
// ## Metadata
// introduced: 0.187.1
//
package hash


// time returns the current system time.
//
// ## Examples
//
// ### Return a stream of tables with the current system time
// ```
// import "hash"
//
// fingerprint = hash.test("something")
// ```
//
// ## Metadata
// tags: convert
//
builtin test : (v: string) => string
