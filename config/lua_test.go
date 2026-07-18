package config

import (
	"os"
	"path/filepath"
	"testing"
)

// sample mirrors the shape of the real prod config.lua: tab indentation,
// comment lines, an inline comment on a number, `serverName` reused inside
// later concatenations, and expression/function RHS that must be ignored.
const sample = "" +
	"-- The Forgotten Server Config\n" +
	"\tworldType = \"open\"\n" +
	"\tworldId = 0\n" +
	"\t-- a comment line = not an assignment\n" +
	"\tredSkullLength = 14 * 24 * 60 * 60\n" +
	"\tserverName = \"Wypas\"\n" +
	"\tloginMessage = \"Welcome to WypasOTS, gameworld \" .. serverName .. \". visit https://wypas.eu\"\n" +
	"\tsqlPass = os.getenv(\"WYPAS_DB_PASSWORD\") or \"\"\n" +
	"\tformulaLevel = 5.0            -- engine default (weapons path)\n" +
	"\trateExperience = 8\n" +
	"\trateSkill = 8.0\n" +
	"\trateMagic = 4.0\n" +
	"\trateLoot = 0.9\n" +
	"\trateSpawnMax = 1\n" +
	"\trateSpawnMin = 1\n" +
	"\tprefixChannelLogs = serverName:gsub(\"'\", \"\") .. \" - \"\n"

func TestLoadLua(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.lua")
	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}

	lc, err := LoadLua(path)
	if err != nil {
		t.Fatalf("LoadLua: %v", err)
	}

	if lc.ServerName != "Wypas" {
		t.Errorf("ServerName = %q, want Wypas (must not be shadowed by loginMessage/prefixChannelLogs)", lc.ServerName)
	}
	if lc.WorldType != "open" {
		t.Errorf("WorldType = %q, want open", lc.WorldType)
	}
	if lc.WorldID != 0 {
		t.Errorf("WorldID = %d, want 0", lc.WorldID)
	}
	if lc.RateExperience != 8 {
		t.Errorf("RateExperience = %v, want 8", lc.RateExperience)
	}
	if lc.RateSkill != 8 {
		t.Errorf("RateSkill = %v, want 8", lc.RateSkill)
	}
	if lc.RateMagic != 4 {
		t.Errorf("RateMagic = %v, want 4", lc.RateMagic)
	}
	if lc.RateLoot != 0.9 {
		t.Errorf("RateLoot = %v, want 0.9", lc.RateLoot)
	}
	if lc.RateSpawnMin != 1 || lc.RateSpawnMax != 1 {
		t.Errorf("RateSpawn = %v/%v, want 1/1", lc.RateSpawnMin, lc.RateSpawnMax)
	}
}

func TestParseLuaScalars_SkipsExpressionsAndFuncs(t *testing.T) {
	kv := parseLuaScalars(sample)

	// Expression and function-call RHS must not produce a literal.
	if v, ok := kv["redSkullLength"]; ok {
		t.Errorf("redSkullLength should be skipped (expression), got %q", v)
	}
	if v, ok := kv["sqlPass"]; ok {
		t.Errorf("sqlPass should be skipped (function call), got %q", v)
	}
	// A trailing inline comment on a number must be stripped.
	if kv["formulaLevel"] != "5.0" {
		t.Errorf("formulaLevel = %q, want 5.0 (inline comment stripped)", kv["formulaLevel"])
	}
	// serverName resolves to the real assignment, not the concatenations.
	if kv["serverName"] != "Wypas" {
		t.Errorf("serverName = %q, want Wypas", kv["serverName"])
	}
}

func TestLoadLua_MissingFile(t *testing.T) {
	if _, err := LoadLua(filepath.Join(t.TempDir(), "nope.lua")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
