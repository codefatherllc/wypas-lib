# wypas-lib — Shared Go Module

## Build

```bash
go vet ./...
go test ./...
```

No binary output. This is a library module consumed via `go get`.

## Packages

| Package | Purpose |
|---------|---------|
| `config` | `GetEnv`, `GetEnvInt`, `SplitCSV`, `ConvertDSN` (mysql:// URI to Go DSN) |
| `db` | `Open(dsn, maxOpen, maxIdle)` — MySQL pool with 5min conn lifetime |
| `jwt` | `Claims`, `ParseToken`, `SignToken`, `Middleware`, `RequireAdmin`, `GetClaims` |
| `middleware` | `CORS(origins...)` — Access-Control headers + OPTIONS preflight |
| `gamedata` | DB-backed item types + world data (`ItemType`, `MapTile`, `Spawn`, `Town`, etc. + `LoadItemTypes`, `LoadTiles`, etc.) |
| `otb` | Legacy file loading (OTB+XML→`gamedata.ItemType`, OTBM+XMLs→gamedata world types). Used by otconv, creator demo, api demo. |
| `otbm` | Raw binary OTBM/OTB parser (internal — `otb/` is the high-level loader returning `gamedata.*` types) |
| `maptile` | Map tile renderer (`Renderer`, `RenderTile`, `RenderMinimapImage`, `EncodePNG`, `From8Bit`). Sprite + minimap rendering with floor stacking, displacement, elevation, patterns. |
| `sprite` | `Cache` (dat+spr loader, PNG render), `DatFile`, `SpriteFile`, outfit color palette |
| `taxonomy` | Item classification (`Role`, `ClassifyItem`), semantic grouping (`GroupByMinimapColor`, `BuildFromItems`), taxonomy JSON schema types (`SemanticGroup`, `AdjacencyRule`, `WallPattern`, `MonsterAffinity`, `WFCAdjacencyData`), JSON loader (`LoadTaxonomy`) + 15 lookup methods. Shared between brain (consumer) and scrapper (producer). |
| `worlds` | `Load(path)` → `WorldList` from JSON, `All()`, `ByID()`, `Default()` |

## OTBM Types

`GameMap` holds `Tiles` (keyed by `PackPos(x,y,z)`), `Towns`, `Waypoints`, `FloorBounds`.

`MapTile` has two item representations:
- `Items []uint16` — flat server-ID list (backward compat)
- `RichItems []MapItem` — full items with attributes

`MapItem` fields: `ID`, `ActionID`, `UniqueID`, `TeleDest *TeleportDest`, `DoorID`, `DepotID`, `Text`, `Description`, `Charges`, `RuneCharges`, `Count`.

`TeleportDest` — `X, Y uint16`, `Z uint8`.

`Waypoint` — `Name string`, `X, Y uint16`, `Z uint8`.

Attribute constants: `AttrActionID` (4), `AttrUniqueID` (5), `AttrText` (6), `AttrDesc` (7), `AttrTeleDest` (8), `AttrItem` (9), `AttrDepotID` (10), `AttrRuneCharges` (12), `AttrDoorID` (14), `AttrCount` (15), `AttrCharges` (22), `AttrAttrMap` (128).

Tile flag constants: `TileFlagProtectionZone` (0x0001), `TileFlagNoPVP` (0x0004), `TileFlagNoLogout` (0x0008), `TileFlagPVPZone` (0x0010), `TileFlagRefresh` (0x0020).

## Conventions

- Go 1.22, module path `github.com/codefatherllc/wypas-lib`
- No CLI, no main package
- Consumers use `replace` directive for local dev
- Release via git tags (`v*`), CI runs vet + test + GitHub Release

## CI

`.github/workflows/release.yml` — triggered on `v*` tags. Runs `go vet`, `go test`, creates GitHub Release.
