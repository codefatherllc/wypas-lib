package gamedata

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var sampleCols = []string{
	"id", "world_id", "level", "vocation", "max_health", "experience", "maglevel",
	"max_mana", "manaspent", "soul", "town_id", "cap",
	"looktype", "lookhead", "lookbody", "looklegs", "lookfeet", "lookaddons",
	"stamina", "loss_experience", "loss_mana", "loss_skills", "loss_containers", "loss_items",
	"posx", "posy", "posz",
}

func rookSampleRow() *sqlmock.Rows {
	return sqlmock.NewRows(sampleCols).AddRow(
		2, 5, 1, 0, 150, 0, 0,
		55, 0, 100, 10, 400,
		128, 78, 68, 58, 76, 0,
		151200000, 100, 100, 100, 100, 100,
		1067, 523, 5,
	)
}

func paladinSampleRow() *sqlmock.Rows {
	return sqlmock.NewRows(sampleCols).AddRow(
		5, 5, 8, 3, 185, 4200, 0,
		90, 0, 100, 1, 470,
		128, 78, 68, 58, 76, 0,
		151200000, 100, 100, 100, 100, 100,
		434, 504, 7,
	)
}

func TestCreateCharacterRookCopiesSampleAndNormalizesStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, world_id, level, vocation, max_health").
		WithArgs("Rook Sample").
		WillReturnRows(rookSampleRow())
	mock.ExpectExec("INSERT INTO players").
		WithArgs(
			"Newbie", 42, 5, 1, 0, 10, 1067, 523, 5, // name, account, sex, voc, sample town, sample pos
			1, 150, 150, 0, 0, // level, health=max, max_health, exp, maglevel
			55, 55, 0, 100, 400, // mana=max, max_mana, manaspent, soul, cap
			128, 78, 68, 58, 76, 0, // looks from sample (male keeps looktype)
			151200000, 100, 100, 100, 100, 100,
		).
		WillReturnResult(sqlmock.NewResult(77, 1))
	mock.ExpectExec("INSERT INTO player_items").
		WithArgs(int64(77), 2).
		WillReturnResult(sqlmock.NewResult(0, 11))
	// vocation 0: no spell copy
	mock.ExpectExec("INSERT INTO player_storage").
		WithArgs(int64(77), 2).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	// TownID=3 must be ignored for vocation None — Rookgaard always.
	id, err := CreateCharacterFromSample(context.Background(), db, CharacterOpts{
		AccountID: 42, Name: "Newbie", Sex: 1, Vocation: 0, TownID: 3,
		AllowedTowns: []int{1, 3, 9},
	})
	if err != nil {
		t.Fatalf("CreateCharacterFromSample: %v", err)
	}
	if id != 77 {
		t.Fatalf("id = %d, want 77", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCharacterVocationTownChoiceZeroesPos(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, world_id, level, vocation, max_health").
		WithArgs("Paladin Sample").
		WillReturnRows(paladinSampleRow())
	mock.ExpectExec("INSERT INTO players").
		WithArgs(
			"Palladyn", 42, 5, 0, 3, 3, 0, 0, 0, // chosen town 3 → pos zeroed (temple drop)
			8, 185, 185, 4200, 0,
			90, 90, 0, 100, 470,
			136, 10, 20, 30, 40, 0, // female looktype 136, caller colours
			151200000, 100, 100, 100, 100, 100,
		).
		WillReturnResult(sqlmock.NewResult(78, 1))
	mock.ExpectExec("INSERT INTO player_items").
		WithArgs(int64(78), 5).
		WillReturnResult(sqlmock.NewResult(0, 11))
	mock.ExpectExec("INSERT INTO player_spells").
		WithArgs(int64(78), 5).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("INSERT INTO player_storage").
		WithArgs(int64(78), 5).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	_, err = CreateCharacterFromSample(context.Background(), db, CharacterOpts{
		AccountID: 42, Name: "Palladyn", Sex: 0, Vocation: 3, TownID: 3,
		LookHead: 10, LookBody: 20, LookLegs: 30, LookFeet: 40,
		AllowedTowns: []int{1, 3, 9},
	})
	if err != nil {
		t.Fatalf("CreateCharacterFromSample: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCharacterDisallowedTownKeepsSampleTownAndPos(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, world_id, level, vocation, max_health").
		WithArgs("Paladin Sample").
		WillReturnRows(paladinSampleRow())
	mock.ExpectExec("INSERT INTO players").
		WithArgs(
			"Palladyn", 42, 5, 1, 3, 1, 434, 504, 7, // town 5 not allowed → sample town + pos
			8, 185, 185, 4200, 0,
			90, 90, 0, 100, 470,
			128, 78, 68, 58, 76, 0,
			151200000, 100, 100, 100, 100, 100,
		).
		WillReturnResult(sqlmock.NewResult(79, 1))
	mock.ExpectExec("INSERT INTO player_items").
		WithArgs(int64(79), 5).
		WillReturnResult(sqlmock.NewResult(0, 11))
	mock.ExpectExec("INSERT INTO player_spells").
		WithArgs(int64(79), 5).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("INSERT INTO player_storage").
		WithArgs(int64(79), 5).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	_, err = CreateCharacterFromSample(context.Background(), db, CharacterOpts{
		AccountID: 42, Name: "Palladyn", Sex: 1, Vocation: 3, TownID: 5,
		AllowedTowns: []int{1, 3, 9},
	})
	if err != nil {
		t.Fatalf("CreateCharacterFromSample: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCharacterUnknownVocationFallsBackToRook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, world_id, level, vocation, max_health").
		WithArgs("Rook Sample").
		WillReturnRows(sqlmock.NewRows(sampleCols)) // no row
	mock.ExpectRollback()

	_, err = CreateCharacterFromSample(context.Background(), db, CharacterOpts{
		AccountID: 42, Name: "Newbie", Vocation: 9,
	})
	if err == nil || !strings.Contains(err.Error(), "Rook Sample") {
		t.Fatalf("want missing-sample error naming Rook Sample, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
