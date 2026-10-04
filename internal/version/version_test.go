package version

import "testing"

func TestStringUsesExplicitVersion(t *testing.T) {
	original := Version
	Version = "1.2.3"
	t.Cleanup(func() {
		Version = original
	})

	if got := String(); got != "1.2.3" {
		t.Fatalf("String() = %q, want %q", got, "1.2.3")
	}
}

func TestReportIncludesReleaseDetails(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalDate := BuildDate
	Version = "1.2.3"
	Commit = "abc123"
	BuildDate = "2026-07-19T00:00:00Z"
	t.Cleanup(func() {
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalDate
	})

	if got := String(); got != "1.2.3" {
		t.Fatalf("String() = %q, want the version only", got)
	}

	const want = "1.2.3\ncommit: abc123\nbuilt: 2026-07-19T00:00:00Z"
	if got := Report(); got != want {
		t.Fatalf("Report() = %q, want %q", got, want)
	}
}

func TestReportOmitsDefaultBuildDetails(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalDate := BuildDate
	Version = "dev"
	Commit = "none"
	BuildDate = "unknown"
	t.Cleanup(func() {
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalDate
	})

	if got := Report(); got != String() {
		t.Fatalf("Report() = %q, want String() %q", got, String())
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		moduleVersion string
		want          string
	}{
		{name: "tagged module", moduleVersion: "v0.1.1", want: "0.1.1"},
		{name: "development build", moduleVersion: "(devel)", want: "dev"},
		{name: "missing build info", moduleVersion: "", want: "dev"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := resolve("dev", test.moduleVersion); got != test.want {
				t.Fatalf("resolve() = %q, want %q", got, test.want)
			}
		})
	}
}
