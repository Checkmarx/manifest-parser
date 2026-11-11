package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// want is the Version and VersionSpec a package is expected to report.
type want struct{ version, spec string }

// versionFixtures holds, per dependency file, the packages it declares and what
// each one reports. Every case is run with WithVersionSpecs off and on: Version
// is the same both ways, and VersionSpec is set only when the option is on and
// Version is "latest". A value a parser passes through as if it were exact
// ("1.x", "LATEST", "file:../x") stays in Version and gets no VersionSpec.
var versionFixtures = []struct {
	file    string
	content string
	want    map[string]want
}{
	{
		file: "package.json",
		content: `{"dependencies": {"exact": "4.0.0", "caret": "^4.0.0", "x": "1.x",
  "file": "file:../langchain", "git": "git+https://github.com/x/y.git", "star": "*"}}`,
		want: map[string]want{
			"exact": {"4.0.0", ""},
			"caret": {"latest", "^4.0.0"},
			"x":     {"1.x", ""},
			"file":  {"file:../langchain", ""},
			"git":   {"git+https://github.com/x/y.git", ""},
			"star":  {"latest", "*"},
		},
	},
	{
		file:    "requirements.txt",
		content: "exact==4.0.0\nranged>=1.0,<2\nwild==1.*\nmulti==1.0,<2\nbare\n",
		want: map[string]want{
			"exact":  {"4.0.0", ""},
			"ranged": {"latest", ">=1.0,<2"},
			"wild":   {"latest", "==1.*"},
			"multi":  {"1.0,<2", ""},
			"bare":   {"latest", ""},
		},
	},
	{
		file: "pyproject.toml",
		content: `[tool.poetry.dependencies]
python = "^3.10"
exact = "4.0.0"
caret = "^4.0.0"
git = { git = "https://github.com/x/y.git" }
`,
		want: map[string]want{
			"exact": {"4.0.0", ""},
			"caret": {"latest", "^4.0.0"},
			"git":   {"latest", ""},
		},
	},
	{
		file: "setup.cfg",
		content: `[options]
install_requires =
    exact==4.0.0
    ranged>=1.0,<2
    bare
`,
		want: map[string]want{
			"exact":  {"4.0.0", ""},
			"ranged": {"latest", ">=1.0,<2"},
			"bare":   {"latest", ""},
		},
	},
	{
		file: "setup.py",
		content: `from setuptools import setup
setup(
    name="x",
    install_requires=["exact==4.0.0", "ranged>=1.0,<2", "wild==1.*"],
)
`,
		want: map[string]want{
			"exact":  {"4.0.0", ""},
			"ranged": {"latest", ">=1.0,<2"},
			"wild":   {"latest", "==1.*"},
		},
	},
	{
		file: "pom.xml",
		content: `<project><groupId>g</groupId><artifactId>a</artifactId><version>1</version>
<properties><openai.version>0.7.2</openai.version><range>[1.0,2.0)</range></properties>
<dependencyManagement><dependencies>
<dependency><groupId>g</groupId><artifactId>managed</artifactId><version>[1.0,2.0)</version></dependency>
</dependencies></dependencyManagement>
<dependencies>
<dependency><groupId>g</groupId><artifactId>exact</artifactId><version>4.0.0</version></dependency>
<dependency><groupId>g</groupId><artifactId>prop</artifactId><version>${openai.version}</version></dependency>
<dependency><groupId>g</groupId><artifactId>range</artifactId><version>${range}</version></dependency>
<dependency><groupId>g</groupId><artifactId>managed</artifactId></dependency>
<dependency><groupId>g</groupId><artifactId>none</artifactId></dependency>
<dependency><groupId>g</groupId><artifactId>latest</artifactId><version>LATEST</version></dependency>
<dependency><groupId>g</groupId><artifactId>missing</artifactId><version>${missing}</version></dependency>
</dependencies></project>`,
		want: map[string]want{
			"g:exact":   {"4.0.0", ""},
			"g:prop":    {"0.7.2", ""},
			"g:range":   {"latest", "[1.0,2.0)"},
			"g:managed": {"latest", "[1.0,2.0)"},
			"g:none":    {"latest", ""},
			"g:latest":  {"LATEST", ""},
			"g:missing": {"${missing}", ""},
		},
	},
	{
		file: "build.gradle",
		content: "dependencies {\n" +
			"    implementation 'g:exact:4.0.0'\n" +
			"    implementation 'g:range:1.+'\n" +
			"    implementation 'g:dynamic:latest.release'\n" +
			"    implementation 'g:unresolved:$undefinedVersion'\n" +
			"}\n",
		want: map[string]want{
			"g:exact":      {"4.0.0", ""},
			"g:range":      {"latest", "1.+"},
			"g:dynamic":    {"latest.release", ""},
			"g:unresolved": {"$undefinedVersion", ""},
		},
	},
	{
		file: "libs.versions.toml",
		content: `[versions]
openai = "1.0.0"
range = "[1.0,2.0)"

[libraries]
exact = { group = "g", name = "exact", version.ref = "openai" }
ranged = { group = "g", name = "ranged", version.ref = "range" }
dynamic = { group = "g", name = "dynamic", version = "1.+" }
`,
		want: map[string]want{
			"g:exact":   {"1.0.0", ""},
			"g:ranged":  {"[1.0,2.0)", ""},
			"g:dynamic": {"1.+", ""},
		},
	},
	{
		file: "app.csproj",
		content: `<Project><ItemGroup>
<PackageReference Include="Exact" Version="4.0.0" />
<PackageReference Include="Caret" Version="^4.0.0" />
<PackageReference Include="Range" Version="[1.0,2.0)" />
<PackageReference Include="Prop" Version="$(OpenAIVersion)" />
</ItemGroup></Project>`,
		want: map[string]want{
			"Exact": {"4.0.0", ""},
			"Caret": {"latest", "^4.0.0"},
			"Range": {"latest", "[1.0,2.0)"},
			"Prop":  {"latest", "$(OpenAIVersion)"},
		},
	},
	{
		file: "packages.config",
		content: `<packages>
<package id="Exact" version="4.0.0" targetFramework="net48" />
<package id="Caret" version="^4.0.0" targetFramework="net48" />
</packages>`,
		want: map[string]want{
			"Exact": {"4.0.0", ""},
			"Caret": {"latest", "^4.0.0"},
		},
	},
	{
		file: "Directory.Packages.props",
		content: `<Project><ItemGroup>
<PackageVersion Include="Exact" Version="4.0.0" />
<PackageVersion Include="Wild" Version="4.*" />
</ItemGroup></Project>`,
		want: map[string]want{
			"Exact": {"4.0.0", ""},
			"Wild":  {"latest", "4.*"},
		},
	},
}

// parseFixture writes content to a file named name and parses it with the
// parser built with opts, returning the packages by name.
func parseFixture(t *testing.T, name, content string, opts ...parser.Option) map[string]models.Package {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	p := parser.ParsersFactoryWithOptions(path, opts...)
	if p == nil {
		t.Fatalf("no parser for %s", name)
	}

	pkgs, err := p.Parse(path)
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", name, err)
	}

	byName := make(map[string]models.Package, len(pkgs))
	for _, pkg := range pkgs {
		byName[pkg.PackageName] = pkg
	}

	return byName
}

func TestVersionFixtures(t *testing.T) {
	for _, fx := range versionFixtures {
		t.Run(fx.file, func(t *testing.T) {
			off := parseFixture(t, fx.file, fx.content)
			on := parseFixture(t, fx.file, fx.content, parser.WithVersionSpecs())

			if len(off) != len(fx.want) || len(on) != len(fx.want) {
				t.Fatalf("got %d (off) and %d (on) packages, want %d: off=%v", len(off), len(on), len(fx.want), off)
			}

			for name, w := range fx.want {
				gotOff, ok := off[name]
				if !ok {
					t.Errorf("%s: not parsed", name)
					continue
				}

				gotOn := on[name]

				if gotOff.Version != w.version || gotOn.Version != w.version {
					t.Errorf("%s: Version = %q (off), %q (on); want %q", name, gotOff.Version, gotOn.Version, w.version)
				}

				if gotOff.VersionSpec != "" {
					t.Errorf("%s: VersionSpec = %q without WithVersionSpecs; want empty", name, gotOff.VersionSpec)
				}

				if gotOn.VersionSpec != w.spec {
					t.Errorf("%s: VersionSpec = %q with WithVersionSpecs; want %q", name, gotOn.VersionSpec, w.spec)
				}
			}
		})
	}
}

// TestParsersFactoryMatchesWithOptions checks that ParsersFactory is the factory
// with no options, so existing callers see the same output.
func TestParsersFactoryMatchesWithOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(path, []byte(`{"dependencies": {"caret": "^4.0.0", "exact": "1.2.3"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	plain, err := parser.ParsersFactory(path).Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	withOpts, err := parser.ParsersFactoryWithOptions(path).Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(plain) != len(withOpts) {
		t.Fatalf("ParsersFactory gave %d packages, ParsersFactoryWithOptions gave %d", len(plain), len(withOpts))
	}

	for i := range plain {
		if plain[i].Version != withOpts[i].Version || plain[i].VersionSpec != withOpts[i].VersionSpec {
			t.Errorf("package %d differs: %+v vs %+v", i, plain[i], withOpts[i])
		}
	}
}
