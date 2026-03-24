package openai

import "testing"

func TestInferIntentFromPrompt(t *testing.T) {
	tests := []struct {
		prompt string
		want   structuredIntent
	}{
		{"create plan for database migration", intentPlan},
		{"create a task list to track deploy", intentQueue},
		{"ask for approval before delete", intentConfirmation},
		{"create deployment runbook template", intentArtifact},
		{"explain svelte with citations and resources", intentCitation},
		{"hello there", intentNone},
	}

	for _, tt := range tests {
		got := inferIntentFromPrompt(tt.prompt)
		if got != tt.want {
			t.Fatalf("inferIntentFromPrompt(%q) = %q, want %q", tt.prompt, got, tt.want)
		}
	}
}

func TestBuildStructuredPart(t *testing.T) {
	planContent := "1. Assess schema\n2. Create migration\n3. Validate data"
	plan, ok := buildStructuredPart(intentPlan, "plan migration", planContent)
	if !ok || plan.Type != "plan" {
		t.Fatalf("expected plan part, got ok=%v type=%q", ok, plan.Type)
	}

	queueContent := "- Setup\n- Test\n- Deploy"
	queue, ok := buildStructuredPart(intentQueue, "task list", queueContent)
	if !ok || queue.Type != "queue" {
		t.Fatalf("expected queue part, got ok=%v type=%q", ok, queue.Type)
	}

	citationContent := "Read [Svelte Docs](https://svelte.dev/docs) and https://developer.mozilla.org."
	citation, ok := buildStructuredPart(intentCitation, "with citations", citationContent)
	if !ok || citation.Type != "citation" {
		t.Fatalf("expected citation part, got ok=%v type=%q", ok, citation.Type)
	}
}
