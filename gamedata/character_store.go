package gamedata

import (
	"context"
	"database/sql"
	"fmt"
)

// CharacterOpts describes a new character to create from a vocation sample.
type CharacterOpts struct {
	AccountID int
	Name      string
	Sex       int
	Vocation  int
	TownID    int
	LookHead  int
	LookBody  int
	LookLegs  int
	LookFeet  int
	// AllowedTowns is the set of towns the user may choose for non-None
	// vocations. Empty means the sample's town is always used.
	AllowedTowns []int
}

// characterSamples maps vocation ID to the sample character name in the DB.
var characterSamples = map[int]string{
	0: "Rook Sample",
	1: "Sorcerer Sample",
	2: "Druid Sample",
	3: "Paladin Sample",
	4: "Knight Sample",
}

type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// CreateCharacterFromSample creates a character by copying the vocation's
// sample player inside its own transaction. See CreateCharacterFromSampleTx.
func CreateCharacterFromSample(ctx context.Context, db *sql.DB, opts CharacterOpts) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	id, err := CreateCharacterFromSampleTx(ctx, tx, opts)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// CreateCharacterFromSampleTx copies a sample character (world, level, stats, spawn
// position, town, items, spells, storage) into a new player row owned by the
// caller's transaction. Unknown vocations fall back to the Rook sample.
// Health and mana are set to their maximums regardless of the sample's
// current values — samples are live rows and may be saved mid-regeneration.
// Skills are auto-created by the oncreate_players DB trigger.
func CreateCharacterFromSampleTx(ctx context.Context, tx *sql.Tx, opts CharacterOpts) (int64, error) {
	return createCharacterFromSample(ctx, tx, opts)
}

func createCharacterFromSample(ctx context.Context, q dbtx, opts CharacterOpts) (int64, error) {
	sampleName, ok := characterSamples[opts.Vocation]
	if !ok {
		sampleName = characterSamples[0]
	}

	// ── 1. Read sample player ────────────────────────────────────────
	var sampleID int
	var sWorldID int
	var sLevel, sVocation int
	var sHealthMax, sManaMax int
	var sExperience int64
	var sManaSpent int64
	var sMagLevel, sSoul, sCap int
	var sLookType, sLookHead, sLookBody, sLookLegs, sLookFeet, sLookAddons int
	var sTownID int
	var sPosX, sPosY, sPosZ int
	var sStamina int64
	var sLossExp, sLossMana, sLossSkills, sLossContainers, sLossItems int

	err := q.QueryRowContext(ctx, `
		SELECT id, world_id, level, vocation, max_health, experience, maglevel,
		       max_mana, manaspent, soul, town_id, cap,
		       looktype, lookhead, lookbody, looklegs, lookfeet, lookaddons,
		       stamina, loss_experience, loss_mana, loss_skills, loss_containers, loss_items,
		       posx, posy, posz
		FROM players WHERE name = ? LIMIT 1`, sampleName).Scan(
		&sampleID, &sWorldID, &sLevel, &sVocation, &sHealthMax, &sExperience, &sMagLevel,
		&sManaMax, &sManaSpent, &sSoul, &sTownID, &sCap,
		&sLookType, &sLookHead, &sLookBody, &sLookLegs, &sLookFeet, &sLookAddons,
		&sStamina, &sLossExp, &sLossMana, &sLossSkills, &sLossContainers, &sLossItems,
		&sPosX, &sPosY, &sPosZ,
	)
	if err != nil {
		return 0, fmt.Errorf("sample character %q not found: %w", sampleName, err)
	}

	// ── 2. Resolve overrides ─────────────────────────────────────────
	// Looktype from sex (htdocs behaviour: female=136, male keeps sample value).
	looktype := sLookType
	if opts.Sex == 0 {
		looktype = 136
	}

	// Look colours: use caller values if provided, else fall back to sample.
	lHead, lBody, lLegs, lFeet := sLookHead, sLookBody, sLookLegs, sLookFeet
	if opts.LookHead != 0 || opts.LookBody != 0 || opts.LookLegs != 0 || opts.LookFeet != 0 {
		lHead, lBody, lLegs, lFeet = opts.LookHead, opts.LookBody, opts.LookLegs, opts.LookFeet
	}

	// Town: None always uses the sample's town (Rookgaard). Otherwise the
	// user's choice applies when it is in the allowed set.
	townID := sTownID
	if opts.Vocation != 0 && opts.TownID != 0 {
		for _, t := range opts.AllowedTowns {
			if t == opts.TownID {
				townID = opts.TownID
				break
			}
		}
	}

	// Spawn position: only meaningful when the character stays in the
	// sample's town — the legacy game drops pos 0/0/0 at the town temple,
	// which is exactly what a user-chosen town needs.
	posX, posY, posZ := sPosX, sPosY, sPosZ
	if townID != sTownID {
		posX, posY, posZ = 0, 0, 0
	}

	// ── 3. Insert new player ─────────────────────────────────────────
	res, err := q.ExecContext(ctx, `
		INSERT INTO players (name, account_id, world_id, sex, vocation, town_id, posx, posy, posz,
		  level, health, max_health, experience, maglevel,
		  mana, max_mana, manaspent, soul, cap,
		  looktype, lookhead, lookbody, looklegs, lookfeet, lookaddons,
		  conditions, stamina,
		  loss_experience, loss_mana, loss_skills, loss_containers, loss_items,
		  created)
		VALUES (?,?,?,?,?,?, ?,?,?, ?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?,?, '',?, ?,?,?,?,?, UNIX_TIMESTAMP())`,
		opts.Name, opts.AccountID, sWorldID, opts.Sex, sVocation, townID, posX, posY, posZ,
		sLevel, sHealthMax, sHealthMax, sExperience, sMagLevel,
		sManaMax, sManaMax, sManaSpent, sSoul, sCap,
		looktype, lHead, lBody, lLegs, lFeet, sLookAddons,
		sStamina,
		sLossExp, sLossMana, sLossSkills, sLossContainers, sLossItems,
	)
	if err != nil {
		return 0, err
	}
	playerID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// ── 4. Copy items from sample ────────────────────────────────────
	// The legacy game reads the classic player_items table (pid/sid/itemtype/
	// count/attributes PropStream blob), NOT the v2 items table. Copy verbatim.
	if _, err := q.ExecContext(ctx, `
		INSERT INTO player_items (player_id, pid, sid, itemtype, count, attributes)
		SELECT ?, pid, sid, itemtype, count, attributes
		FROM player_items WHERE player_id = ?`, playerID, sampleID); err != nil {
		return 0, err
	}

	// ── 5. Copy spells from sample ───────────────────────────────────
	if sVocation != 0 {
		if _, err := q.ExecContext(ctx, `
			INSERT INTO player_spells (player_id, name)
			SELECT ?, name FROM player_spells WHERE player_id = ?`, playerID, sampleID); err != nil {
			return 0, err
		}
	}

	// ── 6. Copy storage from sample (quest flags etc.) ───────────────
	if _, err := q.ExecContext(ctx,
		"INSERT INTO player_storage (player_id, `key`, value) "+
			"SELECT ?, `key`, value FROM player_storage WHERE player_id = ?", playerID, sampleID); err != nil {
		return 0, err
	}

	return playerID, nil
}
