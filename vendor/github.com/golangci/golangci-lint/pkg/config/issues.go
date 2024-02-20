package config

import (
	"fmt"
	"regexp"
)

const excludeRuleMinConditionsCount = 2

var DefaultExcludePatterns = []ExcludePattern{
	{
		ID: "EXC0001",
		Pattern: "Error return value of .((os\\.)?std(out|err)\\..*|.*Close" +
			"|.*Flush|os\\.Remove(All)?|.*print(f|ln)?|os\\.(Un)?Setenv). is not checked",
		Linter: "errcheck",
		Why:    "Almost all programs ignore errors on these functions and in most cases it's ok",
	},
	{
		ID: "EXC0002",
		Pattern: "(comment on exported (method|function|type|const)|" +
			"should have( a package)? comment|comment should be of the form)",
		Linter: "golint",
		Why:    "Annoying issue about not having a comment. The rare codebase has such comments",
	},
	{
		ID:      "EXC0003",
		Pattern: "func name will be used as test\\.Test.* by other packages, and that stutters; consider calling this",
		Linter:  "golint",
		Why:     "False positive when tests are defined in package 'test'",
	},
	{
		ID:      "EXC0004",
		Pattern: "(possible misuse of unsafe.Pointer|should have signature)",
		Linter:  "govet",
		Why:     "Common false positives",
	},
	{
		ID:      "EXC0005",
		Pattern: "ineffective break statement. Did you mean to break out of the outer loop",
		Linter:  "staticcheck",
		Why:     "Developers tend to write in C-style with an explicit 'break' in a 'switch', so it's ok to ignore",
	},
	{
		ID:      "EXC0006",
		Pattern: "Use of unsafe calls should be audited",
		Linter:  "gosec",
		Why:     "Too many false-positives on 'unsafe' usage",
	},
	{
		ID:      "EXC0007",
		Pattern: "Subprocess launch(ed with variable|ing should be audited)",
		Linter:  "gosec",
		Why:     "Too many false-positives for parametrized shell calls",
	},
	{
		ID:      "EXC0008",
		Pattern: "(G104)",
		Linter:  "gosec",
		Why:     "Duplicated errcheck checks",
	},
	{
		ID:      "EXC0009",
		Pattern: "(Expect directory permissions to be 0750 or less|Expect file permissions to be 0600 or less)",
		Linter:  "gosec",
		Why:     "Too many issues in popular repos",
	},
	{
		ID:      "EXC0010",
		Pattern: "Potential file inclusion via variable",
		Linter:  "gosec",
		Why:     "False positive is triggered by 'src, err := ioutil.ReadFile(filename)'",
	},
	{
		ID: "EXC0011",
		Pattern: "(comment on exported (method|function|type|const)|" +
			"should have( a package)? comment|comment should be of the form)",
		Linter: "stylecheck",
		Why:    "Annoying issue about not having a comment. The rare codebase has such comments",
	},
	{
		ID:      "EXC0012",
		Pattern: `exported (.+) should have comment( \(or a comment on this block\))? or be unexported`,
		Linter:  "revive",
		Why:     "Annoying issue about not having a comment. The rare codebase has such comments",
	},
	{
		ID:      "EXC0013",
		Pattern: `package comment should be of the form "(.+)...`,
		Linter:  "revive",
		Why:     "Annoying issue about not having a comment. The rare codebase has such comments",
	},
	{
		ID:      "EXC0014",
		Pattern: `comment on exported (.+) should be of the form "(.+)..."`,
		Linter:  "revive",
		Why:     "Annoying issue about not having a comment. The rare codebase has such comments",
	},
	{
		ID:      "EXC0015",
		Pattern: `should have a package comment`,
		Linter:  "revive",
		Why:     "Annoying issue about not having a comment. The rare codebase has such comments",
	},
}

type Issues struct {
	IncludeDefaultExcludes []string      `mapstructure:"include"`
	ExcludeCaseSensitive   bool          `mapstructure:"exclude-case-sensitive"`
	ExcludePatterns        []string      `mapstructure:"exclude"`
	ExcludeRules           []ExcludeRule `mapstructure:"exclude-rules"`
	UseDefaultExcludes     bool          `mapstructure:"exclude-use-default"`

	MaxIssuesPerLinter int `mapstructure:"max-issues-per-linter"`
	MaxSameIssues      int `mapstructure:"max-same-issues"`

	DiffFromRevision  string `mapstructure:"new-from-rev"`
	DiffPatchFilePath string `mapstructure:"new-from-patch"`
	WholeFiles        bool   `mapstructure:"whole-files"`
	Diff              bool   `mapstructure:"new"`

	NeedFix bool `mapstructure:"fix"`
}

type ExcludeRule struct {
	BaseRule `mapstructure:",squash"`
}

func (e *ExcludeRule) Validate() error {
	return e.BaseRule.Validate(excludeRuleMinConditionsCount)
}

type BaseRule struct {
	Linters    []string
	Path       string
	PathExcept string `mapstructure:"path-except"`
	Text       string
	Source     string
}

func (b *BaseRule) Validate(minConditionsCount int) error {
	if err := validateOptionalRegex(b.Path); err != nil {
		return fmt.Errorf("invalid path regex: %w", err)
	}
	if err := validateOptionalRegex(b.PathExcept); err != nil {
		return fmt.Errorf("invalid path-except regex: %w", err)
	}
	if err := validateOptionalRegex(b.Text); err != nil {
		return fmt.Errorf("invalid text regex: %w", err)
	}
	if err := validateOptionalRegex(b.Source); err != nil {
		return fmt.Errorf("invalid source regex: %w", err)
	}
	nonBlank := 0
	if len(b.Linters) > 0 {
		nonBlank++
	}
	// Filtering by path counts as one condition, regardless how it is done (one or both).
	// Otherwise, a rule with Path and PathExcept set would pass validation
	// whereas before the introduction of path-except that wouldn't have been precise enough.
	if b.Path != "" || b.PathExcept != "" {
		nonBlank++
	}
	if b.Text != "" {
		nonBlank++
	}
	if b.Source != "" {
		nonBlank++
	}
	if nonBlank < minConditionsCount {
		return fmt.Errorf("at least %d of (text, source, path[-except],  linters) should be set", minConditionsCount)
	}
	return nil
}

func validateOptionalRegex(value string) error {
	if value == "" {
		return nil
	}
	_, err := regexp.Compile(value)
	return err
}

type ExcludePattern struct {
	ID      string
	Pattern string
	Linter  string
	Why     string
}

func GetDefaultExcludePatternsStrings() []string {
	ret := make([]string, len(DefaultExcludePatterns))
	for i, p := range DefaultExcludePatterns {
		ret[i] = p.Pattern
	}
	return ret
}

// TODO(ldez): this behavior must be changed in v2, because this is confusing.
func GetExcludePatterns(include []string) []ExcludePattern {
	includeMap := make(map[string]struct{}, len(include))
	for _, inc := range include {
		includeMap[inc] = struct{}{}
	}

	var ret []ExcludePattern
	for _, p := range DefaultExcludePatterns {
		if _, ok := includeMap[p.ID]; !ok {
			ret = append(ret, p)
		}
	}

	return ret
}
