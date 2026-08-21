package app

import "strings"

// DefaultPromptTemplate is the built-in prompt used to ask the model for a
// commit message. It supports {{diff}} and {{stat}} placeholders, which are
// substituted with the staged diff and diffstat respectively.
const DefaultPromptTemplate = `You are an expert software engineer writing a git commit message.

Analyse the following staged diff and produce a commit message that:
- Follows the Conventional Commits spec (feat/fix/chore/refactor/docs/test/ci/build/perf/style)
- Has a concise subject line (imperative mood, <= 72 characters, no trailing period)
- Optionally includes a blank line followed by a short body (max 3 bullet points) if the change warrants explanation
- Does NOT include any preamble, explanation, or markdown fencing -- output the commit message only

--- git diff --stat ---
{{stat}}

--- git diff --cached ---
{{diff}}
`

// BuildPrompt substitutes the {{diff}} and {{stat}} placeholders in tmpl
// with the given diff and stat text.
func BuildPrompt(tmpl, diff, stat string) string {
	replacer := strings.NewReplacer(
		"{{diff}}", diff,
		"{{stat}}", stat,
	)
	return replacer.Replace(tmpl)
}
