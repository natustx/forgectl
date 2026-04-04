
## Batch 1 — L0 types.config (2026-04-03)

**Item:** `[types.config]` ForgeConfig struct hierarchy  
**Commit:** 1fd7e3c  
**Result:** PASS (round 2)

Added the full ForgeConfig struct hierarchy to `state/types.go` and wired it into ForgeState. Round 1 FAIL flagged that ForgeState was missing Config ForgeConfig and still had old flat fields — both fixed in round 2 along with updating all callers (advance.go, output.go, cmd/init.go, all tests).

**Notable:** AgentConfig uses Go struct embedding to produce flat JSON promotion (model/type/count at same level as parent fields), matching state-persistence.md schema. DefaultForgeConfig() provides spec-defined defaults.
