package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	os.MkdirAll(filepath.Join(tmp, "palm"), 0o755)
}

func TestAddEntity(t *testing.T) {
	g := New()

	if err := g.AddEntity("Alice", "person"); err != nil {
		t.Fatalf("AddEntity failed: %v", err)
	}

	if len(g.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(g.Entities))
	}

	e, err := g.GetEntity("Alice")
	if err != nil {
		t.Fatalf("GetEntity failed: %v", err)
	}
	if e.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", e.Name)
	}
	if e.Type != "person" {
		t.Errorf("expected type 'person', got %q", e.Type)
	}
}

func TestAddEntityDuplicate(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")

	err := g.AddEntity("Alice", "person")
	if err == nil {
		t.Fatal("expected error for duplicate entity")
	}
}

func TestAddEntityCaseInsensitive(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")

	err := g.AddEntity("alice", "person")
	if err == nil {
		t.Fatal("expected error for case-insensitive duplicate")
	}

	// Lookup should be case-insensitive
	e, err := g.GetEntity("ALICE")
	if err != nil {
		t.Fatalf("case-insensitive GetEntity failed: %v", err)
	}
	if e.Name != "Alice" {
		t.Errorf("expected display name 'Alice', got %q", e.Name)
	}
}

func TestAddEntityEmpty(t *testing.T) {
	g := New()
	err := g.AddEntity("", "test")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetEntityNotFound(t *testing.T) {
	g := New()
	_, err := g.GetEntity("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent entity")
	}
}

func TestRemoveEntity(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")
	g.AddEntity("Bob", "person")
	g.AddRelation("Alice", "knows", "Bob")

	if err := g.RemoveEntity("Alice"); err != nil {
		t.Fatalf("RemoveEntity failed: %v", err)
	}

	if len(g.Entities) != 1 {
		t.Errorf("expected 1 entity after removal, got %d", len(g.Entities))
	}
	if len(g.Relations) != 0 {
		t.Errorf("expected 0 relations after cascade removal, got %d", len(g.Relations))
	}
}

func TestRemoveEntityNotFound(t *testing.T) {
	g := New()
	err := g.RemoveEntity("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent entity")
	}
}

func TestAddObservation(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")

	if err := g.AddObservation("Alice", "Likes coffee"); err != nil {
		t.Fatalf("AddObservation failed: %v", err)
	}

	e, _ := g.GetEntity("Alice")
	if len(e.Observations) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(e.Observations))
	}
	if e.Observations[0] != "Likes coffee" {
		t.Errorf("expected 'Likes coffee', got %q", e.Observations[0])
	}
}

func TestRemoveObservation(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")
	g.AddObservation("Alice", "fact1")
	g.AddObservation("Alice", "fact2")
	g.AddObservation("Alice", "fact3")

	if err := g.RemoveObservation("Alice", 1); err != nil {
		t.Fatalf("RemoveObservation failed: %v", err)
	}

	e, _ := g.GetEntity("Alice")
	if len(e.Observations) != 2 {
		t.Fatalf("expected 2 observations, got %d", len(e.Observations))
	}
	if e.Observations[0] != "fact1" || e.Observations[1] != "fact3" {
		t.Errorf("unexpected observations: %v", e.Observations)
	}
}

func TestRemoveObservationOutOfBounds(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")

	err := g.RemoveObservation("Alice", 0)
	if err == nil {
		t.Fatal("expected error for out-of-bounds index")
	}

	err = g.RemoveObservation("Alice", -1)
	if err == nil {
		t.Fatal("expected error for negative index")
	}
}

func TestAddRelation(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")
	g.AddEntity("Bob", "person")

	if err := g.AddRelation("Alice", "knows", "Bob"); err != nil {
		t.Fatalf("AddRelation failed: %v", err)
	}

	if len(g.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(g.Relations))
	}
}

func TestAddRelationMissingEntity(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")

	err := g.AddRelation("Alice", "knows", "Bob")
	if err == nil {
		t.Fatal("expected error for missing target entity")
	}

	err = g.AddRelation("Charlie", "knows", "Alice")
	if err == nil {
		t.Fatal("expected error for missing source entity")
	}
}

func TestAddRelationDuplicate(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")
	g.AddEntity("Bob", "person")
	g.AddRelation("Alice", "knows", "Bob")

	err := g.AddRelation("Alice", "knows", "Bob")
	if err == nil {
		t.Fatal("expected error for duplicate relation")
	}
}

func TestRemoveRelation(t *testing.T) {
	g := New()
	g.AddEntity("Alice", "person")
	g.AddEntity("Bob", "person")
	g.AddRelation("Alice", "knows", "Bob")

	if err := g.RemoveRelation("Alice", "knows", "Bob"); err != nil {
		t.Fatalf("RemoveRelation failed: %v", err)
	}

	if len(g.Relations) != 0 {
		t.Errorf("expected 0 relations, got %d", len(g.Relations))
	}
}

func TestRemoveRelationNotFound(t *testing.T) {
	g := New()
	err := g.RemoveRelation("Alice", "knows", "Bob")
	if err == nil {
		t.Fatal("expected error for removing nonexistent relation")
	}
}

func TestRelationsOf(t *testing.T) {
	g := New()
	g.AddEntity("A", "node")
	g.AddEntity("B", "node")
	g.AddEntity("C", "node")
	g.AddRelation("A", "to", "B")
	g.AddRelation("C", "to", "A")

	out, in := g.RelationsOf("A")
	if len(out) != 1 {
		t.Errorf("expected 1 outgoing, got %d", len(out))
	}
	if len(in) != 1 {
		t.Errorf("expected 1 incoming, got %d", len(in))
	}
}

func TestSearch(t *testing.T) {
	g := New()
	g.AddEntity("palm", "project")
	g.AddObservation("palm", "AI tool manager")
	g.AddEntity("Claude Code", "tool")
	g.AddEntity("person1", "person")

	// Search by name
	results := g.Search("palm")
	if len(results) == 0 {
		t.Fatal("expected search results for 'palm'")
	}
	if results[0].Entity.Name != "palm" {
		t.Errorf("expected 'palm' as top result, got %q", results[0].Entity.Name)
	}

	// Search by type
	results = g.Search("tool")
	if len(results) == 0 {
		t.Fatal("expected search results for type 'tool'")
	}

	// Search by observation
	results = g.Search("manager")
	if len(results) == 0 {
		t.Fatal("expected search results for observation 'manager'")
	}

	// Name match scores higher than type match
	results = g.Search("palm")
	nameScore := results[0].Score
	results2 := g.Search("project")
	typeScore := results2[0].Score
	if nameScore <= typeScore {
		t.Errorf("name match score (%d) should be higher than type match score (%d)", nameScore, typeScore)
	}
}

func TestSearchNoResults(t *testing.T) {
	g := New()
	g.AddEntity("palm", "project")

	results := g.Search("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetStats(t *testing.T) {
	g := New()
	g.AddEntity("A", "type1")
	g.AddEntity("B", "type2")
	g.AddObservation("A", "obs1")
	g.AddObservation("A", "obs2")
	g.AddRelation("A", "rel", "B")

	stats := g.GetStats()
	if stats.Entities != 2 {
		t.Errorf("expected 2 entities, got %d", stats.Entities)
	}
	if stats.Relations != 1 {
		t.Errorf("expected 1 relation, got %d", stats.Relations)
	}
	if stats.Observations != 2 {
		t.Errorf("expected 2 observations, got %d", stats.Observations)
	}
	if stats.Types != 2 {
		t.Errorf("expected 2 types, got %d", stats.Types)
	}
}

func TestEntityNames(t *testing.T) {
	g := New()
	g.AddEntity("Charlie", "person")
	g.AddEntity("Alice", "person")
	g.AddEntity("Bob", "person")

	names := g.EntityNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "Alice" || names[1] != "Bob" || names[2] != "Charlie" {
		t.Errorf("expected sorted names, got %v", names)
	}
}

func TestLoadSaveEncryptionRoundtrip(t *testing.T) {
	setupTestEnv(t)

	g := New()
	g.AddEntity("TestEntity", "test")
	g.AddObservation("TestEntity", "encrypted fact")
	g.AddEntity("Other", "test")
	g.AddRelation("TestEntity", "links", "Other")

	if err := Save(g); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify the file is encrypted (not readable as JSON)
	data, err := os.ReadFile(graphPath())
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	var probe map[string]interface{}
	if json.Unmarshal(data, &probe) == nil {
		t.Fatal("graph.enc should not be valid JSON (should be encrypted)")
	}

	// Load back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Entities) != 2 {
		t.Errorf("expected 2 entities after load, got %d", len(loaded.Entities))
	}
	if len(loaded.Relations) != 1 {
		t.Errorf("expected 1 relation after load, got %d", len(loaded.Relations))
	}
	e, err := loaded.GetEntity("TestEntity")
	if err != nil {
		t.Fatalf("GetEntity after load failed: %v", err)
	}
	if len(e.Observations) != 1 || e.Observations[0] != "encrypted fact" {
		t.Errorf("observations not preserved: %v", e.Observations)
	}
}

func TestLoadEmptyGraph(t *testing.T) {
	setupTestEnv(t)

	g, err := Load()
	if err != nil {
		t.Fatalf("Load empty failed: %v", err)
	}
	if len(g.Entities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(g.Entities))
	}
}

func TestExportJSON(t *testing.T) {
	g := New()
	g.AddEntity("Test", "type")
	g.AddObservation("Test", "a fact")

	data, err := g.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Should be valid JSON
	var parsed Graph
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("ExportJSON produced invalid JSON: %v", err)
	}
	if len(parsed.Entities) != 1 {
		t.Errorf("expected 1 entity in export, got %d", len(parsed.Entities))
	}
}

func TestExportDOT(t *testing.T) {
	g := New()
	g.AddEntity("A", "node")
	g.AddEntity("B", "node")
	g.AddRelation("A", "connects", "B")

	dot := g.ExportDOT()
	if !contains(dot, "digraph palm_graph") {
		t.Error("DOT output missing digraph header")
	}
	if !contains(dot, "connects") {
		t.Error("DOT output missing relation label")
	}
	if !contains(dot, "->") {
		t.Error("DOT output missing edge arrow")
	}
}

func TestImportJSON(t *testing.T) {
	g := New()
	g.AddEntity("Existing", "type1")
	g.AddObservation("Existing", "old fact")

	importData := `{
		"entities": {
			"existing": {"name": "Existing", "type": "type1", "observations": ["old fact", "new fact"]},
			"newnode": {"name": "NewNode", "type": "type2", "observations": ["fresh"]}
		},
		"relations": [
			{"from": "Existing", "to": "NewNode", "type": "links"}
		]
	}`

	added, merged, relAdded, err := g.ImportJSON([]byte(importData))
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}

	if added != 1 {
		t.Errorf("expected 1 added, got %d", added)
	}
	if merged != 1 {
		t.Errorf("expected 1 merged, got %d", merged)
	}
	if relAdded != 1 {
		t.Errorf("expected 1 relation added, got %d", relAdded)
	}

	// Check merged observations (should have both old and new, no duplicates)
	e, _ := g.GetEntity("Existing")
	if len(e.Observations) != 2 {
		t.Errorf("expected 2 observations after merge, got %d: %v", len(e.Observations), e.Observations)
	}
}

func TestShowEntity(t *testing.T) {
	g := New()
	g.AddEntity("Center", "hub")
	g.AddEntity("Left", "node")
	g.AddEntity("Right", "node")
	g.AddRelation("Center", "to", "Right")
	g.AddRelation("Left", "to", "Center")

	result, err := g.ShowEntity("Center")
	if err != nil {
		t.Fatalf("ShowEntity failed: %v", err)
	}
	if result.Entity.Name != "Center" {
		t.Errorf("expected entity name 'Center', got %q", result.Entity.Name)
	}
	if len(result.Outgoing) != 1 {
		t.Errorf("expected 1 outgoing, got %d", len(result.Outgoing))
	}
	if len(result.Incoming) != 1 {
		t.Errorf("expected 1 incoming, got %d", len(result.Incoming))
	}
}

func TestRenderShow(t *testing.T) {
	g := New()
	g.AddEntity("Test", "type")
	g.AddObservation("Test", "a fact")

	identity := func(s string) string { return s }
	output, err := RenderShow(g, "Test", identity, identity, identity)
	if err != nil {
		t.Fatalf("RenderShow failed: %v", err)
	}
	if !contains(output, "Test") {
		t.Error("RenderShow output missing entity name")
	}
	if !contains(output, "a fact") {
		t.Error("RenderShow output missing observation")
	}
}

func TestExportHTML(t *testing.T) {
	g := New()
	g.AddEntity("Node1", "type1")
	g.AddEntity("Node2", "type2")
	g.AddRelation("Node1", "links", "Node2")

	html := g.ExportHTML()
	if !contains(html, "palm graph") {
		t.Error("HTML output missing title")
	}
	if !contains(html, "Node1") {
		t.Error("HTML output missing node name")
	}
	if !contains(html, "canvas") {
		t.Error("HTML output missing canvas element")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
