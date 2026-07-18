package config

import (
	"os"
	"strconv"
	"strings"
)

// LuaConfig holds the handful of authoritative game scalars the website/API
// mirrors from the game server's config.lua. The game server (classic TFS)
// owns these values; api reads them here instead of duplicating them in
// config.yaml.
type LuaConfig struct {
	ServerName     string
	WorldType      string
	WorldID        int
	RateExperience float64
	RateSkill      float64
	RateMagic      float64
	RateLoot       float64
	RateSpawnMin   float64
	RateSpawnMax   float64
}

// LoadLua reads the authoritative scalars from a TFS config.lua. It is a
// deliberately small key=value reader for plain string/number literals — NOT a
// Lua interpreter — because every value it needs is a simple top-level
// assignment (serverName = "Wypas", worldId = 0, rateExperience = 8, ...).
// Keys assigned an expression or absent keep their zero value; callers decide
// fallbacks.
func LoadLua(path string) (*LuaConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	kv := parseLuaScalars(string(data))
	return &LuaConfig{
		ServerName:     kv["serverName"],
		WorldType:      kv["worldType"],
		WorldID:        int(luaNum(kv, "worldId")),
		RateExperience: luaNum(kv, "rateExperience"),
		RateSkill:      luaNum(kv, "rateSkill"),
		RateMagic:      luaNum(kv, "rateMagic"),
		RateLoot:       luaNum(kv, "rateLoot"),
		RateSpawnMin:   luaNum(kv, "rateSpawnMin"),
		RateSpawnMax:   luaNum(kv, "rateSpawnMax"),
	}, nil
}

// parseLuaScalars extracts `identifier = <literal>` assignments whose value is a
// bare string or number literal. It anchors on the assignment target (the token
// left of `=`), so a value that merely mentions an identifier (e.g. loginMessage
// = "..." .. serverName .. "...") never shadows the real serverName assignment.
// Values that are expressions (e.g. 14 * 24 * 60 * 60) or lookups are skipped.
func parseLuaScalars(src string) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		if !isLuaIdent(key) {
			continue
		}
		if _, seen := out[key]; seen {
			continue // first assignment wins
		}
		if v, ok := cleanLuaLiteral(strings.TrimSpace(line[eq+1:])); ok {
			out[key] = v
		}
	}
	return out
}

// cleanLuaLiteral returns the value of a bare string or number literal and
// whether the RHS was one. A quoted string yields its contents; a bare number
// yields its leading numeric token (dropping any trailing comment). Anything
// else (identifier, expression, table) returns ok=false.
func cleanLuaLiteral(rhs string) (string, bool) {
	if rhs == "" {
		return "", false
	}
	if q := rhs[0]; q == '"' || q == '\'' {
		if end := strings.IndexByte(rhs[1:], q); end >= 0 {
			return rhs[1 : 1+end], true
		}
		return "", false
	}
	// Number: take the leading token, reject if it isn't wholly numeric so an
	// arithmetic RHS (`14 * 24`) or a lookup doesn't masquerade as a literal.
	tok := rhs
	if i := strings.IndexAny(tok, " \t"); i >= 0 {
		tok = tok[:i]
	}
	if _, err := strconv.ParseFloat(tok, 64); err != nil {
		return "", false
	}
	// Guard against `14 * 24 ...`: if what follows the token is an operator, the
	// RHS is an expression, not a literal — skip it.
	rest := strings.TrimSpace(rhs[len(tok):])
	if rest != "" && !strings.HasPrefix(rest, "--") {
		return "", false
	}
	return tok, true
}

func isLuaIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_':
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

func luaNum(kv map[string]string, key string) float64 {
	if v, ok := kv[key]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
