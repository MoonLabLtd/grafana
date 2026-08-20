// SPDX-License-Identifier: AGPL-3.0-only

package prdemo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDraftPRDerivesTitleForClearInput(t *testing.T) {
	draft := BuildDraftPR("add user login audit logging")
	assert.Equal(t, "feat: Add user login audit logging", draft.Title)
	assert.Contains(t, draft.Description, "user login audit logging")
	assert.Equal(t, 3, len(draft.ChangedFiles))
}

func TestBuildDraftPRFixVerb(t *testing.T) {
	draft := BuildDraftPR("fix a broken dashboard panel")
	assert.Equal(t, "fix: Fix a broken dashboard panel", draft.Title)
	assert.Contains(t, draft.Description, "broken dashboard panel")
	assert.True(t, draft.Description[0:13] == "This PR fixes ")
}

func TestBuildDraftPRVagueInputFallsBackToPlaceholder(t *testing.T) {
	for _, input := range []string{"fix stuff", "update", "", "   "} {
		draft := BuildDraftPR(input)
		assert.Equal(t, "chore: [placeholder: provide a specific change description]", draft.Title)
		assert.Contains(t, draft.Description, "too vague")
	}
}

func TestBuildChangedFilesAreHardcodedAndStable(t *testing.T) {
	draft := BuildDraftPR("add telemetry")
	assert.Equal(t, 3, len(draft.ChangedFiles))
	assert.Equal(t, "updated", draft.ChangedFiles[0].Action)
	assert.Equal(t, "apps/provisioning/pkg/repository/repository.go", draft.ChangedFiles[0].Path)
	assert.Equal(t, "main", draft.ChangedFiles[0].Ref)
	assert.Equal(t, "created", draft.ChangedFiles[1].Action)
	assert.Equal(t, "feat-demo-branch", draft.ChangedFiles[1].Ref)
	assert.Equal(t, "deleted", draft.ChangedFiles[2].Action)
}

func TestBuildDraftPRDeterministicOutput(t *testing.T) {
	a := BuildDraftPR("add retrieval helpers")
	b := BuildDraftPR("add retrieval helpers")
	assert.Equal(t, a, b)
	assert.Equal(t, json.MarshalIndent(&a, "", "  "), json.MarshalIndent(&b, "", "  "))
}
