package ai

import (
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// ─── Test helpers ────────────────────────────────────────────────────────────

func makeUnit(id, ownerID int, x, y float64, opts ...func(*game.UnitView)) game.UnitView {
	u := game.UnitView{
		ID:            id,
		OwnerID:       ownerID,
		Position:      [2]float64{x, y},
		AliveModels:   3,
		TotalModels:   3,
		CurrentWounds: 9,
		MaxWounds:     9,
		MoveSpeed:     5,
		Save:          4,
	}
	for _, fn := range opts {
		fn(&u)
	}
	return u
}

func withMelee(name string, attacks, toHit, toWound, rend, damage int) func(*game.UnitView) {
	return func(u *game.UnitView) {
		u.Weapons = append(u.Weapons, game.WeaponView{
			Name: name, Range: 0, Attacks: attacks,
			ToHit: toHit, ToWound: toWound, Rend: rend, Damage: damage,
		})
	}
}

func withRanged(name string, rng, attacks, toHit, toWound, rend, damage int) func(*game.UnitView) {
	return func(u *game.UnitView) {
		u.Weapons = append(u.Weapons, game.WeaponView{
			Name: name, Range: rng, Attacks: attacks,
			ToHit: toHit, ToWound: toWound, Rend: rend, Damage: damage,
		})
	}
}

func withSpell(name string, cv, rng int) func(*game.UnitView) {
	return func(u *game.UnitView) {
		u.CanCast = true
		u.Spells = append(u.Spells, game.SpellView{Name: name, CastingValue: cv, Range: rng})
	}
}

func withPrayer(name string, chant, rng int) func(*game.UnitView) {
	return func(u *game.UnitView) {
		u.CanChant = true
		u.Prayers = append(u.Prayers, game.PrayerView{Name: name, ChantingValue: chant, Range: rng})
	}
}

func withWounded(current, max int) func(*game.UnitView) {
	return func(u *game.UnitView) {
		u.CurrentWounds = current
		u.MaxWounds = max
	}
}

func withEngaged() func(*game.UnitView) {
	return func(u *game.UnitView) { u.IsEngaged = true }
}

func withMoved() func(*game.UnitView) {
	return func(u *game.UnitView) { u.HasMoved = true }
}

func withShot() func(*game.UnitView) {
	return func(u *game.UnitView) { u.HasShot = true }
}

func withCharged() func(*game.UnitView) {
	return func(u *game.UnitView) { u.HasCharged = true }
}

func withFought() func(*game.UnitView) {
	return func(u *game.UnitView) { u.HasFought = true }
}

func withRun() func(*game.UnitView) {
	return func(u *game.UnitView) { u.HasRun = true }
}

func baseView(myUnits []game.UnitView, enemyUnits []game.UnitView) *game.GameView {
	return &game.GameView{
		Units: map[int][]game.UnitView{
			1: myUnits,
			2: enemyUnits,
		},
		BoardWidth:      48,
		BoardHeight:     24,
		CommandPoints:   map[int]int{1: 4, 2: 4},
		VictoryPoints:   map[int]int{1: 0, 2: 0},
		BattleTactics:   map[int]*game.BattleTacticsView{},
	}
}

func heroPhase() phase.Phase {
	return phase.Phase{Type: phase.PhaseHero}
}

func movementPhase() phase.Phase {
	return phase.Phase{Type: phase.PhaseMovement}
}

func shootingPhase() phase.Phase {
	return phase.Phase{Type: phase.PhaseShooting}
}

func chargePhase() phase.Phase {
	return phase.Phase{Type: phase.PhaseCharging}
}

func combatPhase() phase.Phase {
	return phase.Phase{Type: phase.PhaseCombat, Alternating: true}
}

func endPhase() phase.Phase {
	return phase.Phase{Type: phase.PhaseEndOfTurn}
}

// ─── Hero Phase Tests ────────────────────────────────────────────────────────

func TestHeroPhase_CastsSpell(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	wizard := makeUnit(1, 1, 10, 12, withSpell("Lightning", 7, 18))
	enemy := makeUnit(10, 2, 20, 12)
	view := baseView([]game.UnitView{wizard}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, heroPhase())
	cast, ok := cmd.(*command.CastCommand)
	if !ok {
		t.Fatalf("expected CastCommand, got %T", cmd)
	}
	if cast.CasterID != 1 {
		t.Errorf("expected caster 1, got %d", cast.CasterID)
	}
	if cast.SpellIndex != 0 {
		t.Errorf("expected spell index 0, got %d", cast.SpellIndex)
	}
	if cast.TargetID != 10 {
		t.Errorf("expected target 10, got %d", cast.TargetID)
	}
}

func TestHeroPhase_ChantsPrayer(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	priest := makeUnit(1, 1, 10, 12, withPrayer("Smite", 5, 12))
	enemy := makeUnit(10, 2, 18, 12)
	view := baseView([]game.UnitView{priest}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, heroPhase())
	chant, ok := cmd.(*command.ChantCommand)
	if !ok {
		t.Fatalf("expected ChantCommand, got %T", cmd)
	}
	if chant.ChanterID != 1 {
		t.Errorf("expected chanter 1, got %d", chant.ChanterID)
	}
	if chant.PrayerIndex != 0 {
		t.Errorf("expected prayer index 0, got %d", chant.PrayerIndex)
	}
}

func TestHeroPhase_RalliesWoundedUnit(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	wounded := makeUnit(1, 1, 10, 12, withWounded(3, 9))
	enemy := makeUnit(10, 2, 30, 12) // Far away
	view := baseView([]game.UnitView{wounded}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, heroPhase())
	rally, ok := cmd.(*command.RallyCommand)
	if !ok {
		t.Fatalf("expected RallyCommand, got %T", cmd)
	}
	if rally.UnitID != 1 {
		t.Errorf("expected unit 1, got %d", rally.UnitID)
	}
}

func TestHeroPhase_NoRallyWhenEngaged(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	wounded := makeUnit(1, 1, 10, 12, withWounded(3, 9), withEngaged())
	enemy := makeUnit(10, 2, 11, 12) // Close
	view := baseView([]game.UnitView{wounded}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, heroPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when engaged, got %T", cmd)
	}
}

func TestHeroPhase_NoRallyWhenNoCPReserve(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	wounded := makeUnit(1, 1, 10, 12, withWounded(3, 9))
	enemy := makeUnit(10, 2, 30, 12)
	view := baseView([]game.UnitView{wounded}, []game.UnitView{enemy})
	view.CommandPoints[1] = 1 // Only 1 CP, reserve=1 so won't spend

	cmd := ai.GetNextCommand(view, heroPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand with low CP, got %T", cmd)
	}
}

func TestHeroPhase_SpellOutOfRange(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	wizard := makeUnit(1, 1, 0, 0, withSpell("Lightning", 7, 12))
	enemy := makeUnit(10, 2, 40, 0) // 40" away, spell is 12"
	view := baseView([]game.UnitView{wizard}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, heroPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when spell out of range, got %T", cmd)
	}
}

// ─── Movement Phase Tests ────────────────────────────────────────────────────

func TestMovement_MovesTowardEnemy(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 5, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 20, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, movementPhase())
	move, ok := cmd.(*command.MoveCommand)
	if !ok {
		t.Fatalf("expected MoveCommand, got %T", cmd)
	}
	if move.UnitID != 1 {
		t.Errorf("expected unit 1, got %d", move.UnitID)
	}
	// Should move closer to enemy
	if move.Destination.X <= 5.0 {
		t.Errorf("expected to move right toward enemy, got X=%.1f", move.Destination.X)
	}
}

func TestMovement_PrefersObjective(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 24, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 45, 12) // Far away

	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})
	// Add an uncontrolled objective nearby
	view.Objectives = []game.ObjectiveView{
		{ID: 1, Position: [2]float64{20, 12}, ControlledBy: -1},
	}

	cmd := ai.GetNextCommand(view, movementPhase())
	move, ok := cmd.(*command.MoveCommand)
	if !ok {
		t.Fatalf("expected MoveCommand, got %T", cmd)
	}
	// Should move toward objective (X=20) not enemy (X=45)
	if move.Destination.X > 24.0 {
		t.Errorf("expected to move left toward objective, got X=%.1f", move.Destination.X)
	}
}

func TestMovement_RunsTowardDistantObjective(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 5, 12)
	view := baseView([]game.UnitView{warrior}, nil) // No enemies

	// Distant uncontrolled objective
	view.Objectives = []game.ObjectiveView{
		{ID: 1, Position: [2]float64{30, 12}, ControlledBy: -1},
	}

	cmd := ai.GetNextCommand(view, movementPhase())
	_, isRun := cmd.(*command.RunCommand)
	_, isMove := cmd.(*command.MoveCommand)

	if !isRun && !isMove {
		t.Fatalf("expected RunCommand or MoveCommand, got %T", cmd)
	}
}

func TestMovement_SkipsMovedUnits(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	moved := makeUnit(1, 1, 5, 12, withMoved())
	enemy := makeUnit(10, 2, 20, 12)
	view := baseView([]game.UnitView{moved}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, movementPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand for already moved unit, got %T", cmd)
	}
}

func TestMovement_SkipsEngagedUnits(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	engaged := makeUnit(1, 1, 10, 12, withEngaged(), withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 11, 12)
	view := baseView([]game.UnitView{engaged}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, movementPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand for engaged unit, got %T", cmd)
	}
}

func TestMovement_RespectsThreeInchRule(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 13, 12) // Only 3" away
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, movementPhase())
	// Unit is within 3.1", should skip (already within safe distance)
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when already within 3\", got %T", cmd)
	}
}

// ─── Shooting Phase Tests ────────────────────────────────────────────────────

func TestShooting_ShootsNearestEnemy(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	archer := makeUnit(1, 1, 10, 12, withRanged("Bow", 18, 2, 3, 4, 0, 1))
	enemy := makeUnit(10, 2, 20, 12)
	view := baseView([]game.UnitView{archer}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, shootingPhase())
	shoot, ok := cmd.(*command.ShootCommand)
	if !ok {
		t.Fatalf("expected ShootCommand, got %T", cmd)
	}
	if shoot.ShooterID != 1 {
		t.Errorf("expected shooter 1, got %d", shoot.ShooterID)
	}
	if shoot.TargetID != 10 {
		t.Errorf("expected target 10, got %d", shoot.TargetID)
	}
}

func TestShooting_PrefersWoundedEnemy(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	archer := makeUnit(1, 1, 10, 12, withRanged("Bow", 24, 2, 3, 4, 0, 1))
	healthy := makeUnit(10, 2, 20, 12, withWounded(9, 9))
	wounded := makeUnit(11, 2, 22, 12, withWounded(2, 9))
	view := baseView([]game.UnitView{archer}, []game.UnitView{healthy, wounded})

	cmd := ai.GetNextCommand(view, shootingPhase())
	shoot, ok := cmd.(*command.ShootCommand)
	if !ok {
		t.Fatalf("expected ShootCommand, got %T", cmd)
	}
	// Should prefer the wounded enemy
	if shoot.TargetID != 11 {
		t.Errorf("expected target 11 (wounded), got %d", shoot.TargetID)
	}
}

func TestShooting_SkipsAfterRun(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	archer := makeUnit(1, 1, 10, 12, withRanged("Bow", 18, 2, 3, 4, 0, 1), withRun())
	enemy := makeUnit(10, 2, 20, 12)
	view := baseView([]game.UnitView{archer}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, shootingPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand after run, got %T", cmd)
	}
}

func TestShooting_SkipsOutOfRange(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	archer := makeUnit(1, 1, 0, 0, withRanged("Bow", 12, 2, 3, 4, 0, 1))
	enemy := makeUnit(10, 2, 40, 0) // 40" away
	view := baseView([]game.UnitView{archer}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, shootingPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when out of range, got %T", cmd)
	}
}

// ─── Charge Phase Tests ─────────────────────────────────────────────────────

func TestCharge_ChargesNearbyEnemy(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 18, 12) // 8" away - within 12"
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, chargePhase())
	charge, ok := cmd.(*command.ChargeCommand)
	if !ok {
		t.Fatalf("expected ChargeCommand, got %T", cmd)
	}
	if charge.ChargerID != 1 {
		t.Errorf("expected charger 1, got %d", charge.ChargerID)
	}
	if charge.TargetID != 10 {
		t.Errorf("expected target 10, got %d", charge.TargetID)
	}
}

func TestCharge_PrefersCloserTarget(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	far := makeUnit(10, 2, 20, 12)  // 10" away
	close := makeUnit(11, 2, 14, 12) // 4" away
	view := baseView([]game.UnitView{warrior}, []game.UnitView{far, close})

	cmd := ai.GetNextCommand(view, chargePhase())
	charge, ok := cmd.(*command.ChargeCommand)
	if !ok {
		t.Fatalf("expected ChargeCommand, got %T", cmd)
	}
	if charge.TargetID != 11 {
		t.Errorf("expected closer target 11, got %d", charge.TargetID)
	}
}

func TestCharge_SkipsAlreadyCharged(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withCharged())
	enemy := makeUnit(10, 2, 18, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, chargePhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand for already charged, got %T", cmd)
	}
}

func TestCharge_SkipsAfterRun(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withRun())
	enemy := makeUnit(10, 2, 18, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, chargePhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand after run, got %T", cmd)
	}
}

func TestCharge_SkipsTooFar(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 0, 0)
	enemy := makeUnit(10, 2, 40, 0) // 40" away - way beyond 12"
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, chargePhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when too far, got %T", cmd)
	}
}

// ─── Combat Phase Tests ─────────────────────────────────────────────────────

func TestCombat_FightsEngagedEnemy(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withEngaged(), withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 11, 12) // Within 3"
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, combatPhase())
	fight, ok := cmd.(*command.FightCommand)
	if !ok {
		t.Fatalf("expected FightCommand, got %T", cmd)
	}
	if fight.AttackerID != 1 {
		t.Errorf("expected attacker 1, got %d", fight.AttackerID)
	}
	if fight.TargetID != 10 {
		t.Errorf("expected target 10, got %d", fight.TargetID)
	}
}

func TestCombat_AllOutAttackForStrongAttacker(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	// Unit with 6 attacks (>= 4 threshold)
	warrior := makeUnit(1, 1, 10, 12, withEngaged(),
		withMelee("Greatsword", 6, 3, 3, 1, 2))
	enemy := makeUnit(10, 2, 11, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	// First call should issue All-out Attack
	cmd := ai.GetNextCommand(view, combatPhase())
	aoa, ok := cmd.(*command.AllOutAttackCommand)
	if !ok {
		t.Fatalf("expected AllOutAttackCommand, got %T", cmd)
	}
	if aoa.UnitID != 1 {
		t.Errorf("expected unit 1, got %d", aoa.UnitID)
	}

	// Second call should issue FightCommand
	cmd = ai.GetNextCommand(view, combatPhase())
	_, ok = cmd.(*command.FightCommand)
	if !ok {
		t.Fatalf("expected FightCommand after boost, got %T", cmd)
	}
}

func TestCombat_AllOutDefenceForWoundedUnit(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	// Wounded unit with low attacks (won't trigger AllOutAttack)
	warrior := makeUnit(1, 1, 10, 12, withEngaged(),
		withWounded(2, 9), // Below 50% wounds
		withMelee("Dagger", 1, 4, 4, 0, 1))
	enemy := makeUnit(10, 2, 11, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	// First call should issue All-out Defence
	cmd := ai.GetNextCommand(view, combatPhase())
	aod, ok := cmd.(*command.AllOutDefenceCommand)
	if !ok {
		t.Fatalf("expected AllOutDefenceCommand, got %T", cmd)
	}
	if aod.UnitID != 1 {
		t.Errorf("expected unit 1, got %d", aod.UnitID)
	}
}

func TestCombat_NoBoostWithLowCP(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withEngaged(),
		withMelee("Greatsword", 6, 3, 3, 1, 2))
	enemy := makeUnit(10, 2, 11, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})
	view.CommandPoints[1] = 1 // Only 1 CP, reserve=1

	cmd := ai.GetNextCommand(view, combatPhase())
	// Should go straight to fight, no boost
	if _, ok := cmd.(*command.FightCommand); !ok {
		t.Fatalf("expected FightCommand with low CP, got %T", cmd)
	}
}

func TestCombat_SkipsNotEngaged(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 20, 12) // Not within 3"
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	cmd := ai.GetNextCommand(view, combatPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when not engaged, got %T", cmd)
	}
}

// ─── End of Turn / Battle Tactic Tests ──────────────────────────────────────

func TestEndOfTurn_SelectsBattleTactic(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12)
	enemy := makeUnit(10, 2, 30, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	// Add available battle tactics
	view.BattleTactics[1] = &game.BattleTacticsView{
		AvailableCards: []game.BattleTacticCardView{
			{
				CardID:   2, // Broken Ranks
				CardName: "Broken Ranks",
				Tiers: [3]game.BattleTacticOptionView{
					{Tier: "Affray", Name: "Eliminate Threat", VP: 4},
					{Tier: "Strike", Name: "Crush Opposition", VP: 4},
					{Tier: "Domination", Name: "Total Annihilation", VP: 4},
				},
			},
		},
	}

	cmd := ai.GetNextCommand(view, endPhase())
	sel, ok := cmd.(*command.SelectBattleTacticCommand)
	if !ok {
		t.Fatalf("expected SelectBattleTacticCommand, got %T", cmd)
	}
	if sel.CardID != 2 {
		t.Errorf("expected card 2, got %d", sel.CardID)
	}
}

func TestEndOfTurn_SkipsWhenAlreadySelected(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12)
	view := baseView([]game.UnitView{warrior}, nil)
	view.BattleTactics[1] = &game.BattleTacticsView{
		ActiveTactic: &game.ActiveBattleTacticView{
			CardName: "Broken Ranks",
			TacticName: "Eliminate Threat",
		},
		AvailableCards: []game.BattleTacticCardView{
			{CardID: 1, CardName: "Savage Spearhead"},
		},
	}

	cmd := ai.GetNextCommand(view, endPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when tactic already selected, got %T", cmd)
	}
}

func TestEndOfTurn_SkipsWhenNoTactics(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12)
	view := baseView([]game.UnitView{warrior}, nil)

	cmd := ai.GetNextCommand(view, endPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("expected EndPhaseCommand when no tactics available, got %T", cmd)
	}
}

// ─── Tactic Scoring Tests ───────────────────────────────────────────────────

func TestChooseTactic_PrefersBrokenRanksWhenEnemiesWounded(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12)
	wounded := makeUnit(10, 2, 20, 12, withWounded(1, 9))
	view := baseView([]game.UnitView{warrior}, []game.UnitView{wounded})

	tactics := &game.BattleTacticsView{
		AvailableCards: []game.BattleTacticCardView{
			{CardID: 1, CardName: "Savage Spearhead"},
			{CardID: 2, CardName: "Broken Ranks"},
		},
	}

	cardID, tier := ai.chooseBestTactic(view, tactics)
	if cardID != 2 {
		t.Errorf("expected Broken Ranks (2), got card %d", cardID)
	}
	if tier != 0 {
		t.Errorf("expected Affray tier (0), got %d", tier)
	}
}

func TestChooseTactic_PrefersSavageSpearheadWhenControllingObjectives(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12)
	enemy := makeUnit(10, 2, 30, 12)
	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})
	view.Objectives = []game.ObjectiveView{
		{ID: 1, Position: [2]float64{10, 12}, ControlledBy: 1},
		{ID: 2, Position: [2]float64{20, 12}, ControlledBy: 1},
		{ID: 3, Position: [2]float64{38, 12}, ControlledBy: 2},
	}

	tactics := &game.BattleTacticsView{
		AvailableCards: []game.BattleTacticCardView{
			{CardID: 1, CardName: "Savage Spearhead"},
			{CardID: 4, CardName: "Ferocious Advance"},
		},
	}

	cardID, _ := ai.chooseBestTactic(view, tactics)
	if cardID != 1 {
		t.Errorf("expected Savage Spearhead (1) when controlling more objectives, got card %d", cardID)
	}
}

// ─── CP Budgeting Tests ─────────────────────────────────────────────────────

func TestCPBudgeting_ReservesOneCP(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	view := baseView(nil, nil)
	view.CommandPoints[1] = 2 // 2 CP: can spend 1, reserve 1

	if !ai.canSpendCP(view, 1, false) {
		t.Error("should be able to spend 1 CP with 2 remaining (1 reserve)")
	}

	view.CommandPoints[1] = 1 // 1 CP: can't spend due to reserve
	if ai.canSpendCP(view, 1, false) {
		t.Error("should not spend when it would deplete reserve")
	}
}

func TestCPBudgeting_ForceOverridesReserve(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	view := baseView(nil, nil)
	view.CommandPoints[1] = 1

	if !ai.canSpendCP(view, 1, true) {
		t.Error("force=true should override reserve")
	}
}

// ─── Integration: Multiple Phases ────────────────────────────────────────────

func TestAI_FullTurnSequence(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 5, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	enemy := makeUnit(10, 2, 30, 12)

	view := baseView([]game.UnitView{warrior}, []game.UnitView{enemy})

	// Hero phase: no spells, no wounds -> end
	cmd := ai.GetNextCommand(view, heroPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("hero phase: expected EndPhaseCommand, got %T", cmd)
	}

	// Movement phase: should move toward enemy
	cmd = ai.GetNextCommand(view, movementPhase())
	if _, ok := cmd.(*command.MoveCommand); !ok {
		t.Fatalf("movement phase: expected MoveCommand, got %T", cmd)
	}

	// Shooting phase: no ranged -> end
	cmd = ai.GetNextCommand(view, shootingPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("shooting phase: expected EndPhaseCommand, got %T", cmd)
	}

	// Charge phase: enemy at 25" -> too far -> end
	cmd = ai.GetNextCommand(view, chargePhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("charge phase: expected EndPhaseCommand (too far), got %T", cmd)
	}

	// Combat phase: not engaged -> end
	cmd = ai.GetNextCommand(view, combatPhase())
	if _, ok := cmd.(*command.EndPhaseCommand); !ok {
		t.Fatalf("combat phase: expected EndPhaseCommand, got %T", cmd)
	}
}

func TestAI_NoEnemies(t *testing.T) {
	ai := NewAIPlayer(1, "AI")

	warrior := makeUnit(1, 1, 10, 12, withMelee("Sword", 2, 3, 3, 1, 1))
	view := baseView([]game.UnitView{warrior}, nil) // No enemies

	phases := []phase.Phase{heroPhase(), shootingPhase(), chargePhase(), combatPhase()}
	for _, p := range phases {
		cmd := ai.GetNextCommand(view, p)
		if _, ok := cmd.(*command.EndPhaseCommand); !ok {
			t.Errorf("phase %s: expected EndPhaseCommand with no enemies, got %T", p.Type, cmd)
		}
	}
}

// ─── Command Type Tests ─────────────────────────────────────────────────────

func TestAllOutAttackCommand(t *testing.T) {
	cmd := &command.AllOutAttackCommand{OwnerID: 1, UnitID: core.UnitID(5)}
	if cmd.Type() != command.CommandTypeAllOutAttack {
		t.Errorf("expected type %s, got %s", command.CommandTypeAllOutAttack, cmd.Type())
	}
	if cmd.PlayerID() != 1 {
		t.Errorf("expected player 1, got %d", cmd.PlayerID())
	}
	if cmd.GetUnitID() != 5 {
		t.Errorf("expected unit 5, got %d", cmd.GetUnitID())
	}
}

func TestAllOutDefenceCommand(t *testing.T) {
	cmd := &command.AllOutDefenceCommand{OwnerID: 2, UnitID: core.UnitID(3)}
	if cmd.Type() != command.CommandTypeAllOutDefence {
		t.Errorf("expected type %s, got %s", command.CommandTypeAllOutDefence, cmd.Type())
	}
	if cmd.PlayerID() != 2 {
		t.Errorf("expected player 2, got %d", cmd.PlayerID())
	}
}

func TestSelectBattleTacticCommand(t *testing.T) {
	cmd := &command.SelectBattleTacticCommand{OwnerID: 1, CardID: 2, Tier: 0}
	if cmd.Type() != command.CommandTypeSelectBattleTactic {
		t.Errorf("expected type %s, got %s", command.CommandTypeSelectBattleTactic, cmd.Type())
	}
	if cmd.GetCardID() != 2 {
		t.Errorf("expected card 2, got %d", cmd.GetCardID())
	}
	if cmd.GetTier() != 0 {
		t.Errorf("expected tier 0, got %d", cmd.GetTier())
	}
}
