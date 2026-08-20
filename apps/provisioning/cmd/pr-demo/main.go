// SPDX-License-Identifier: AGPL-3.0-only
//
// pr-demo is a small local CLI that translates a natural language change
// description into a structured draft pull request (title, description, and a
// hardcoded list of changed files). It performs no network or git operations
// and is intended purely as a local demonstration of PR metadata generation
// for the Grafana provisioning module.

package prdemo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	com.github.grafana.grafana.apps.provisioning.pkg.repository
)

const maxInputLengthBytes = 1 << 20

// DraftPR is the structured draft pull request emitted by the CLI.
type DraftPR struct {
	Title        string                             `json:"title"`
	Description  string                             `json:"description"`
	ChangedFiles []repository.VersionedFileChange `json:"changedFiles"`
}

// demoChangedFiles is the static, hardcoded list of changed files that is
// always attached to the output regardless of the input description. It is
// intentionally deterministic so that manual review of the output is
// reproducible across invocations.
var demoChangedFiles = []repository.VersionedFileChange{
	{
		Action:       "updated",
		Path:         "apps/provisioning/pkg/repository/repository.go",
		Ref:          "main",
		PreviousRef:  "main",
		PreviousPath: "apps/provisioning/pkg/repository/repository.go",
	},
	{
		Action:       "created",
		Path:         "apps/provisioning/pkg/demo/new_feature.go",
		Ref:          "feat-demo-branch",
		PreviousRef:  "",
		PreviousPath: "",
	},
	{
		Action:       "deleted",
		Path:         "apps/provisioning/pkg/legacy/deprecated.go",
		Ref:          "main",
		PreviousRef:  "main",
		PreviousPath: "apps/provisioning/pkg/legacy/deprecated.go",
	},
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: pr-demo \"<description of the change>\"\n")
		os.Exit(1)
	}

	if len(args[1]) > maxInputLengthBytes {
		fmt.Fprintf(
			os.Stderr,
			"Error: input exceeds the maximum supported length of %d bytes\n",
			maxInputLengthBytes,
		)
		os.Exit(1)
	}

	draft := BuildDraftPR(args[1])
	payload, err := json.MarshalIndent(&draft, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to serialize draft PR: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(payload)
}

// BuildDraftPR derives a title and description from the input text. When the
// input is vague or empty it consistently falls back to a structured
// placeholder draft so that the CLI never crashes or panics.
func BuildDraftPR(input string) DraftPR {
	clean := strings.TrimSpace(input)
	if isVagueInput(clean) {
		return DraftPR{
			Title:       "chore: [placeholder: provide a specific change description]",
			Description: fmt.Sprintf("<!-- TODO: Expand on the intended change. The input '%s' was too vague to produce a detailed description. -->", clean),
			ChangedFiles: demoChangedFiles,
		}
	}

	words := strings.Fields(clean)
	verb := strings.ToLower(words[0])
	object := strings.Join(words[1:], " ")
	if object == "" {
		object = "the requested change"
	}

	return DraftPR{
		Title:       fmt.Sprintf("%s: %s", inferChangeType(verb), capitalize(clean)),
		Description: fmt.Sprintf(
			"This PR %s %s. The change aligns with the Grafana provisioning module conventions and updates the demo repository state. Reviewers should validate the draft against the intended change before merging.",
			verbToThirdPerson(verb),
			object,
		),
		ChangedFiles: demoChangedFiles,
	}
}

func isVagueInput(input string) bool {
	if input == "" {
		return true
	}
	words := strings.Fields(input)
	if len(words) < 2 {
		return true
	}
	for _, word := range words {
		if isVagueToken(strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func isVagueToken(word string) bool {
	switch word {
	case "stuff", "thing", "things", "something", "anything", "misc", "various", "unspecified", "whatever":
		return true
	default:
		return false
	}
}

func inferChangeType(verb string) string {
	switch verb {
	case "add", "create", "implement", "introduce", "feature", "new":
		return "feat"
	case "fix", "repair", "correct", "resolve", "patch":
		return "fix"
	default:
		return "chore"
	}
}

func verbToThirdPerson(verb string) string {
	if strings.HasSuffix(verb, "ch") || strings.HasSuffix(verb, "sh") || strings.HasSuffix(verb, "ss") {
		return verb + "es"
	}
	if strings.HasSuffix(verb, "y") {
		return verb[:len(verb)-1] + "ies"
	}
	return verb + "s"
}

func capitalize(input string) string {
	if input == "" {
		return input
	}
	return strings.ToUpper(input[:1]) + input[1:]
}
