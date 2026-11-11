package versionspec

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		ecosystem, value string
		wantExact, want  string
	}{
		{"npm", "4.0.0", "4.0.0", ""},
		{"npm", "4.0.0-beta.1+build.5", "4.0.0-beta.1+build.5", ""},
		{"npm", "^4.0.0", "", "^4.0.0"},
		{"npm", "1.x", "", "1.x"},
		{"npm", "1.0", "", "1.0"},
		{"npm", "*", "", "*"},
		{"npm", "next", "", "next"},
		{"npm", ">=7.0.0 <8", "", ">=7.0.0 <8"},
		// Not a registry version: looked up by name.
		{"npm", "file:../langchain", "", ""},
		{"npm", "git+https://github.com/x/y.git", "", ""},
		{"npm", "github:user/repo", "", ""},
		{"npm", "user/repo", "", ""},
		{"npm", "workspace:*", "", ""},
		{"npm", "npm:ai@4.0.0", "", ""},

		{"pypi", "1.2.3", "1.2.3", ""},
		{"pypi", "1.*", "", "1.*"},
		{"pypi", "1.0,<2", "", "1.0,<2"},
		{"pypi", "~=0.30", "", "~=0.30"},

		{"maven", "0.7.2", "0.7.2", ""},
		{"maven", "1.0-SNAPSHOT", "1.0-SNAPSHOT", ""},
		{"maven", "[1.0,2.0)", "", "[1.0,2.0)"},
		{"maven", "LATEST", "", "LATEST"},
		{"maven", "RELEASE", "", "RELEASE"},
		{"maven", "${openai.version}", "", ""},

		{"gradle", "2.8.0", "2.8.0", ""},
		{"gradle", "1.+", "", "1.+"},
		{"gradle", "(,2.0]", "", "(,2.0]"},
		{"gradle", "latest.release", "", "latest.release"},
		{"gradle", "$openaiVersion", "", ""},

		{"nuget", "2.1.0-beta.2", "2.1.0-beta.2", ""},
		{"nuget", "2.*", "", "2.*"},
		{"nuget", "[2.0,3.0)", "", "[2.0,3.0)"},
		{"nuget", "^2.0.0", "", "^2.0.0"},
		{"nuget", "$(OpenAIVersion)", "", ""},

		{"go", "v3.6.1", "v3.6.1", ""},
		{"go", "v0.0.0-20240101120000-abcdef123456", "v0.0.0-20240101120000-abcdef123456", ""},

		{"npm", "", "", ""},
		{"npm", "  4.0.0 ", "4.0.0", ""},
	}

	for _, tt := range tests {
		exact, spec := Classify(tt.ecosystem, tt.value)
		if exact != tt.wantExact || spec != tt.want {
			t.Errorf("Classify(%q, %q) = (%q, %q), want (%q, %q)", tt.ecosystem, tt.value, exact, spec, tt.wantExact, tt.want)
		}
	}
}

func TestResult(t *testing.T) {
	tests := []struct {
		name            string
		exact, spec     string
		report          bool
		wantVer, wantSp string
	}{
		{"exact is kept, spec dropped", "1.0.0", "", false, "1.0.0", ""},
		{"exact wins over a spec", "1.0.0", "^1.0.0", true, "1.0.0", ""},
		{"range not reported by default", "", "^1.0.0", false, Latest, ""},
		{"range reported when asked", "", "^1.0.0", true, Latest, "^1.0.0"},
		{"nothing declared", "", "", true, Latest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, spec := Result(tt.exact, tt.spec, tt.report)
			if ver != tt.wantVer || spec != tt.wantSp {
				t.Errorf("Result(%q, %q, %v) = (%q, %q), want (%q, %q)", tt.exact, tt.spec, tt.report, ver, spec, tt.wantVer, tt.wantSp)
			}
		})
	}
}

func TestPEP508Version(t *testing.T) {
	tests := []struct {
		requirement     string
		wantExact, want string
	}{
		{"pkg==1.0", "1.0", "==1.0"},
		{"pkg===1.0", "1.0", "===1.0"},
		{"pkg[extra]==1.0", "1.0", "==1.0"},
		{"pkg==1.0; python_version < '3.8'", "1.0", "==1.0"},
		{"pkg==1.*", "", "==1.*"},
		{"pkg==1.0,<2", "", "==1.0,<2"},
		{"pkg>=1.0", "", ">=1.0"},
		{"pkg", "", ""},
		{"pkg @ https://example.com/pkg.whl", "", ""},
	}

	for _, tt := range tests {
		exact, spec := PEP508Version(tt.requirement)
		if exact != tt.wantExact || spec != tt.want {
			t.Errorf("PEP508Version(%q) = (%q, %q), want (%q, %q)", tt.requirement, exact, spec, tt.wantExact, tt.want)
		}
	}
}
