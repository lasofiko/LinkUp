// Package interaction contains the shared Sprint 2 contract for a pair of users.
//
// The package deliberately has no HTTP, database, or discovery dependencies.
// Contract D owns state transitions and the in-memory Store implementation;
// contracts A, B, and C depend only on the exported types and interfaces here.
package interaction
