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
