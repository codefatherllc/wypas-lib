package taxonomy

import (
	"os"
	"path/filepath"
	"testing"
)

func loadFromJSON(t *testing.T, blob string) *Taxonomy {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "taxonomy.json"), []byte(blob), 0o644); err != nil {
		t.Fatal(err)
	}
	tax, err := LoadTaxonomy(dir)
	if err != nil {
		t.Fatal(err)
	}
	return tax
}

func TestVocabLookups(t *testing.T) {
	tax := loadFromJSON(t, `{
		"ground_groups": {},
		"deco_vocab": {"2030": 5},
		"ground_id_vocab": {"4526": 1, "103": 2},
		"border_vocab": {"4644": 1}
	}`)

	if got := tax.GroundIDIndex(4526); got != 1 {
		t.Errorf("GroundIDIndex(4526) = %d, want 1", got)
	}
	if got := tax.GroundIDIndex(9999); got != 0 {
		t.Errorf("GroundIDIndex(9999) = %d, want 0", got)
	}
	if got := tax.GroundIDForVocab(2); got != 103 {
		t.Errorf("GroundIDForVocab(2) = %d, want 103", got)
	}
	if got := tax.GroundIDForVocab(7); got != 0 {
		t.Errorf("GroundIDForVocab(7) = %d, want 0", got)
	}
	if got := tax.BorderIndex(4644); got != 1 {
		t.Errorf("BorderIndex(4644) = %d, want 1", got)
	}
	if got := tax.BorderItemID(1); got != 4644 {
		t.Errorf("BorderItemID(1) = %d, want 4644", got)
	}
	if got := tax.DecoItemID(5); got != 2030 {
		t.Errorf("DecoItemID(5) = %d, want 2030", got)
	}
}

func TestVocabAbsentFields(t *testing.T) {
	tax := loadFromJSON(t, `{"ground_groups": {}, "deco_vocab": {}}`)

	if got := tax.GroundIDIndex(4526); got != 0 {
		t.Errorf("GroundIDIndex without vocab = %d, want 0", got)
	}
	if got := tax.GroundIDForVocab(1); got != 0 {
		t.Errorf("GroundIDForVocab without vocab = %d, want 0", got)
	}
	if got := tax.BorderIndex(4644); got != 0 {
		t.Errorf("BorderIndex without vocab = %d, want 0", got)
	}
	if got := tax.BorderItemID(1); got != 0 {
		t.Errorf("BorderItemID without vocab = %d, want 0", got)
	}
}
