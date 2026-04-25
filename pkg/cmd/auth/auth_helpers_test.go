package auth_test

// auth_helpers_test.go — shared test helpers for auth sub-command tests.

// authConfig is the shared single-host config used by auth command tests.
// Previously declared as `logoutConfig` in logout_test.go and `statusConfig`
// in status_test.go — consolidated here to eliminate the duplication.
const authConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
