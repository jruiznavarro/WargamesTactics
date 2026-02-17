package game

import (
	"strings"
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// Example warscroll spells (no generic spells in AoS4).
func testDamageSpell() core.Spell {
	return core.Spell{
		Name: "Chain Lightning", CastingValue: 7, Range: 18,
		Effect: core.SpellEffectDamage, TargetFriendly: false,
	}
}

func testBuffSpell() core.Spell {
	return core.Spell{
		Name: "Shield of Faith", CastingValue: 5, Range: 12,
		Effect: core.SpellEffectBuff, EffectValue: 1, TargetFriendly: true,
	}
}

func testHealSpell() core.Spell {
	return core.Spell{
		Name: "Lifebloom", CastingValue: 6, Range: 12,
		Effect: core.SpellEffectHeal, TargetFriendly: true,
	}
}

// Example warscroll prayers.
func testDamagePrayer() core.Prayer {
	return core.Prayer{
		Name: "Divine Wrath", ChantingValue: 6, Range: 12,
		Effect: core.SpellEffectDamage, TargetFriendly: false,
	}
}

func testHealPrayer() core.Prayer {
	return core.Prayer{
		Name: "Healing Light", ChantingValue: 4, Range: 12,
		Effect: core.SpellEffectHeal, TargetFriendly: true,
	}
}

// Helper: create a game with a Wizard and an enemy unit.
func setupWizardGame(seed int64) (*Game, *core.Unit, *core.Unit) {
	g := NewGame(seed, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	wizard := g.CreateUnit("Battlemage", 1,
		core.Stats{Move: 6, Save: 5, Control: 1, Health: 5},
		nil, 1, core.Position{X: 10, Y: 12}, 1.0)
	wizard.Keywords = []core.Keyword{core.KeywordHero, core.KeywordWizard}
	wizard.Spells = []core.Spell{testDamageSpell(), testBuffSpell(), testHealSpell()}

	enemy := g.CreateUnit("Target Squad", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 3, core.Position{X: 20, Y: 12}, 1.0)

	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero
	return g, wizard, enemy
}

// Helper: create a game with a Priest and units.
func setupPriestGame(seed int64) (*Game, *core.Unit, *core.Unit) {
	g := NewGame(seed, 48, 24)
	p1 := &stubPlayer{id: 1, name: "P1"}
	p2 := &stubPlayer{id: 2, name: "P2"}
	g.AddPlayer(p1)
	g.AddPlayer(p2)

	priest := g.CreateUnit("War Priest", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 5},
		nil, 1, core.Position{X: 10, Y: 12}, 1.0)
	priest.Keywords = []core.Keyword{core.KeywordHero, core.KeywordPriest}
	priest.Prayers = []core.Prayer{testHealPrayer(), testDamagePrayer()}

	enemy := g.CreateUnit("Enemy Squad", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 3, core.Position{X: 20, Y: 12}, 1.0)

	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero
	return g, priest, enemy
}

// --- SPELL TESTS ---

func TestCastSpell_DamageSpell(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, wizard, enemy := setupWizardGame(seed)
		woundsBefore := enemy.TotalCurrentWounds()

		cmd := &command.CastCommand{
			OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			if enemy.TotalCurrentWounds() >= woundsBefore {
				t.Errorf("seed %d: spell succeeded but no damage dealt", seed)
			}
			return
		}
	}
	t.Fatal("no seed found where damage spell succeeds within 200 attempts")
}

func TestCastSpell_BuffSpell(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, wizard, _ := setupWizardGame(seed)

		friendly := g.CreateUnit("Friendly Warriors", 1,
			core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
			nil, 3, core.Position{X: 15, Y: 12}, 1.0)

		cmd := &command.CastCommand{
			OwnerID: 1, CasterID: wizard.ID, SpellIndex: 1, TargetID: friendly.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			foundLog := false
			for _, msg := range g.Log {
				if strings.Contains(msg, "+1 to save") {
					foundLog = true
					break
				}
			}
			if !foundLog {
				t.Error("expected log message about +1 save buff")
			}
			return
		}
	}
	t.Fatal("no seed found where buff spell succeeds within 200 attempts")
}

func TestCastSpell_FailsBelowCastingValue(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, wizard, enemy := setupWizardGame(seed)

		cmd := &command.CastCommand{
			OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Success && strings.Contains(result.Description, "failed to cast") {
			return
		}
	}
	t.Fatal("no seed found where casting fails within 200 attempts")
}

func TestCastSpell_Miscast(t *testing.T) {
	// Double 1s = miscast: D3 mortal damage + no more spells
	for seed := int64(1); seed < 2000; seed++ {
		g, wizard, enemy := setupWizardGame(seed)
		woundsBefore := wizard.TotalCurrentWounds()

		cmd := &command.CastCommand{
			OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
		}

		result, _ := g.ExecuteCommand(cmd)

		if !result.Success && strings.Contains(result.Description, "miscast") {
			// Wizard should have taken mortal damage
			if wizard.TotalCurrentWounds() >= woundsBefore && !wizard.IsDestroyed() {
				t.Error("miscast should deal mortal damage to caster")
			}
			// HasMiscast should be set
			if !wizard.HasMiscast {
				t.Error("HasMiscast should be true after miscast")
			}
			// Should not be able to cast again
			if wizard.CanCast() {
				t.Error("wizard should not be able to cast after miscast")
			}
			return
		}
	}
	t.Fatal("no seed found where miscast occurs within 2000 attempts")
}

func TestCastSpell_NotAWizard(t *testing.T) {
	g, _, enemy := setupWizardGame(42)

	warrior := g.CreateUnit("Warrior", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.CastCommand{
		OwnerID: 1, CasterID: warrior.ID, SpellIndex: 0, TargetID: enemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when non-wizard tries to cast")
	}
}

func TestCastSpell_PowerLevel(t *testing.T) {
	g, wizard, enemy := setupWizardGame(42)
	wizard.PowerLevel = 2 // Wizard(2) can cast twice

	cmd := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
	}

	// First cast should work
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error on first cast: %v", err)
	}

	// Second cast should also work (power level 2)
	// Need a different spell name since same-spell-once
	wizard.Spells = append(wizard.Spells, core.Spell{
		Name: "Fireball", CastingValue: 5, Range: 18,
		Effect: core.SpellEffectDamage, TargetFriendly: false,
	})
	cmd2 := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 3, TargetID: enemy.ID,
	}
	_, err = g.ExecuteCommand(cmd2)
	if err != nil {
		t.Fatalf("unexpected error on second cast with power level 2: %v", err)
	}

	// Third cast should fail
	wizard.Spells = append(wizard.Spells, core.Spell{
		Name: "Ice Storm", CastingValue: 5, Range: 18,
		Effect: core.SpellEffectDamage, TargetFriendly: false,
	})
	cmd3 := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 4, TargetID: enemy.ID,
	}
	_, err = g.ExecuteCommand(cmd3)
	if err == nil {
		t.Error("expected error on third cast (power level 2)")
	}
}

func TestCastSpell_SameSpellOncePerTurn(t *testing.T) {
	g, wizard, enemy := setupWizardGame(42)
	wizard.PowerLevel = 3 // Enough power level

	cmd := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
	}

	// First cast of Chain Lightning
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error on first cast: %v", err)
	}

	// Second cast of same spell should fail (same spell once per turn)
	_, err = g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error: same spell cannot be cast twice per turn")
	}
	if err != nil && !strings.Contains(err.Error(), "already been cast this turn") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCastSpell_UnlimitedSpell(t *testing.T) {
	g, wizard, enemy := setupWizardGame(42)
	wizard.PowerLevel = 3
	wizard.Spells[0] = core.Spell{
		Name: "Minor Bolt", CastingValue: 5, Range: 18,
		Effect: core.SpellEffectDamage, TargetFriendly: false,
		Unlimited: true,
	}

	cmd := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
	}

	// First cast
	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error on first cast: %v", err)
	}

	// Second cast of same Unlimited spell should work
	_, err = g.ExecuteCommand(cmd)
	if err != nil {
		t.Errorf("Unlimited spell should be castable multiple times: %v", err)
	}
}

func TestCastSpell_OutOfRange(t *testing.T) {
	g, wizard, _ := setupWizardGame(42)

	farEnemy := g.CreateUnit("Far Enemy", 2,
		core.Stats{Move: 5, Save: 4, Health: 3},
		nil, 1, core.Position{X: 40, Y: 12}, 1.0)

	cmd := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: farEnemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error for target out of spell range")
	}
}

func TestCastSpell_WrongTargetOwnership(t *testing.T) {
	g, wizard, enemy := setupWizardGame(42)

	// Shield of Faith targets friendly, so targeting enemy should fail
	cmd := &command.CastCommand{
		OwnerID: 1, CasterID: wizard.ID, SpellIndex: 1, TargetID: enemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when casting friendly spell on enemy")
	}
}

func TestCastSpell_Unbind(t *testing.T) {
	for seed := int64(1); seed < 500; seed++ {
		g, wizard, enemy := setupWizardGame(seed)

		enemyWizard := g.CreateUnit("Enemy Wizard", 2,
			core.Stats{Move: 5, Save: 5, Control: 1, Health: 5},
			nil, 1, core.Position{X: 25, Y: 12}, 1.0)
		enemyWizard.Keywords = []core.Keyword{core.KeywordHero, core.KeywordWizard}
		enemyWizard.Spells = []core.Spell{testDamageSpell()}

		cmd := &command.CastCommand{
			OwnerID: 1, CasterID: wizard.ID, SpellIndex: 0, TargetID: enemy.ID,
		}

		result, _ := g.ExecuteCommand(cmd)

		if !result.Success && strings.Contains(result.Description, "unbound") {
			if enemyWizard.UnbindCount != 1 {
				t.Error("expected enemy wizard unbind count to be 1")
			}
			return
		}
	}
	t.Fatal("no seed found where unbind succeeds within 500 attempts")
}

// --- PRAYER / RITUAL POINTS TESTS ---

func TestChant_BankRitualPoints(t *testing.T) {
	for seed := int64(1); seed < 100; seed++ {
		g, priest, enemy := setupPriestGame(seed)

		cmd := &command.ChantCommand{
			OwnerID: 1, ChanterID: priest.ID, PrayerIndex: 0,
			TargetID: enemy.ID, BankPoints: true,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			if priest.RitualPoints <= 0 {
				t.Errorf("seed %d: ritual points should be > 0 after banking", seed)
			}
			return
		}
		// Roll of 1 = failure, try next seed
	}
	t.Fatal("no seed found where banking succeeds within 100 attempts")
}

func TestChant_FailOnRollOfOne(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, priest, enemy := setupPriestGame(seed)
		priest.RitualPoints = 5 // Give some points to lose

		cmd := &command.ChantCommand{
			OwnerID: 1, ChanterID: priest.ID, PrayerIndex: 0,
			TargetID: enemy.ID, BankPoints: true,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if !result.Success && strings.Contains(result.Description, "rolled 1") {
			if priest.RitualPoints >= 5 {
				t.Error("should have lost ritual points on roll of 1")
			}
			return
		}
	}
	t.Fatal("no seed found where chant roll of 1 occurs within 200 attempts")
}

func TestChant_SpendRitualPoints_Answer(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, priest, enemy := setupPriestGame(seed)
		// Give enough ritual points so that even a low roll can answer the prayer
		priest.RitualPoints = 10 // ChantingValue for Divine Wrath is 6

		woundsBefore := enemy.TotalCurrentWounds()

		cmd := &command.ChantCommand{
			OwnerID: 1, ChanterID: priest.ID, PrayerIndex: 1, // Divine Wrath (damage)
			TargetID: enemy.ID, BankPoints: false,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success && strings.Contains(result.Description, "answered") {
			// Prayer should have dealt damage
			if enemy.TotalCurrentWounds() >= woundsBefore {
				t.Errorf("seed %d: prayer answered but no damage dealt", seed)
			}
			// Ritual points should be reset to 0
			if priest.RitualPoints != 0 {
				t.Errorf("ritual points should be 0 after answered prayer, got %d", priest.RitualPoints)
			}
			return
		}
	}
	t.Fatal("no seed found where prayer is answered within 200 attempts")
}

func TestChant_SpendRitualPoints_Fail(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, priest, enemy := setupPriestGame(seed)
		priest.RitualPoints = 0 // No ritual points, high chanting value needed

		cmd := &command.ChantCommand{
			OwnerID: 1, ChanterID: priest.ID, PrayerIndex: 1, // Divine Wrath CV 6
			TargetID: enemy.ID, BankPoints: false,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if !result.Success && strings.Contains(result.Description, "failed to answer") {
			// Ritual points should be 0 (reset on failed spend)
			if priest.RitualPoints != 0 {
				t.Errorf("ritual points should be 0 after failed spend, got %d", priest.RitualPoints)
			}
			return
		}
	}
	t.Fatal("no seed found where prayer spend fails within 200 attempts")
}

func TestChant_HealPrayer(t *testing.T) {
	for seed := int64(1); seed < 200; seed++ {
		g, priest, _ := setupPriestGame(seed)
		priest.RitualPoints = 10 // Enough to answer Healing Light (CV 4)

		friendly := g.CreateUnit("Wounded Warriors", 1,
			core.Stats{Move: 5, Save: 4, Control: 1, Health: 5},
			nil, 1, core.Position{X: 12, Y: 12}, 1.0)
		friendly.Models[0].CurrentWounds = 1

		cmd := &command.ChantCommand{
			OwnerID: 1, ChanterID: priest.ID, PrayerIndex: 0,
			TargetID: friendly.ID, BankPoints: false,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			if friendly.Models[0].CurrentWounds <= 1 {
				t.Errorf("seed %d: Heal prayer succeeded but wounds not restored", seed)
			}
			return
		}
	}
	t.Fatal("no seed found where heal prayer succeeds within 200 attempts")
}

func TestChant_NotAPriest(t *testing.T) {
	g, _, enemy := setupPriestGame(42)

	warrior := g.CreateUnit("Warrior", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.ChantCommand{
		OwnerID: 1, ChanterID: warrior.ID, PrayerIndex: 0,
		TargetID: enemy.ID, BankPoints: true,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when non-priest tries to chant")
	}
}

func TestChant_CantChantTwice(t *testing.T) {
	g, priest, enemy := setupPriestGame(42)

	cmd := &command.ChantCommand{
		OwnerID: 1, ChanterID: priest.ID, PrayerIndex: 0,
		TargetID: enemy.ID, BankPoints: true,
	}

	_, err := g.ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error on first chant: %v", err)
	}

	_, err = g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error on second chant (power level 1)")
	}
}

func TestChant_RitualPointsPersistAcrossTurns(t *testing.T) {
	priest := &core.Unit{
		RitualPoints: 5,
	}
	priest.ResetPhaseFlags()

	// Ritual points should NOT be reset
	if priest.RitualPoints != 5 {
		t.Errorf("ritual points should persist across turns, got %d", priest.RitualPoints)
	}
}

// --- UNIT METHOD TESTS ---

func TestCanCast_DefaultPowerLevel(t *testing.T) {
	wizard := &core.Unit{
		Keywords: []core.Keyword{core.KeywordWizard},
		Spells:   []core.Spell{testDamageSpell()},
	}

	if !wizard.CanCast() {
		t.Error("wizard with spells should be able to cast")
	}

	wizard.CastCount = 1
	if wizard.CanCast() {
		t.Error("wizard should not cast after using default power level 1")
	}
}

func TestCanCast_HigherPowerLevel(t *testing.T) {
	wizard := &core.Unit{
		Keywords:   []core.Keyword{core.KeywordWizard},
		Spells:     []core.Spell{testDamageSpell()},
		PowerLevel: 2,
	}

	wizard.CastCount = 1
	if !wizard.CanCast() {
		t.Error("wizard(2) should still be able to cast after 1")
	}

	wizard.CastCount = 2
	if wizard.CanCast() {
		t.Error("wizard(2) should not cast after 2 of 2")
	}
}

func TestCanCast_MiscastBlocksFurther(t *testing.T) {
	wizard := &core.Unit{
		Keywords: []core.Keyword{core.KeywordWizard},
		Spells:   []core.Spell{testDamageSpell()},
	}
	wizard.HasMiscast = true

	if wizard.CanCast() {
		t.Error("wizard should not cast after miscast")
	}
}

func TestCanChant_DefaultPowerLevel(t *testing.T) {
	priest := &core.Unit{
		Keywords: []core.Keyword{core.KeywordPriest},
		Prayers:  []core.Prayer{testDamagePrayer()},
	}

	if !priest.CanChant() {
		t.Error("priest with prayers should be able to chant")
	}

	priest.ChantCount = 1
	if priest.CanChant() {
		t.Error("priest should not chant after using default power level 1")
	}
}

func TestCanUnbind_PowerLevel(t *testing.T) {
	wizard := &core.Unit{
		Keywords:   []core.Keyword{core.KeywordWizard},
		PowerLevel: 2,
		Models:     []core.Model{{ID: 0, IsAlive: true, CurrentWounds: 5, MaxWounds: 5}},
	}

	if !wizard.CanUnbind() {
		t.Error("wizard should be able to unbind")
	}

	wizard.UnbindCount = 1
	if !wizard.CanUnbind() {
		t.Error("wizard(2) should still be able to unbind after 1")
	}

	wizard.UnbindCount = 2
	if wizard.CanUnbind() {
		t.Error("wizard(2) should not unbind after 2 of 2")
	}
}

func TestResetPhaseFlags_ResetsMagicCounters(t *testing.T) {
	u := &core.Unit{
		CastCount:    1,
		ChantCount:   1,
		UnbindCount:  1,
		HasMiscast:   true,
		RitualPoints: 5,
	}
	u.ResetPhaseFlags()

	if u.CastCount != 0 {
		t.Error("CastCount should be reset")
	}
	if u.ChantCount != 0 {
		t.Error("ChantCount should be reset")
	}
	if u.UnbindCount != 0 {
		t.Error("UnbindCount should be reset")
	}
	if u.HasMiscast {
		t.Error("HasMiscast should be reset")
	}
	if u.RitualPoints != 5 {
		t.Error("RitualPoints should NOT be reset (persist across turns)")
	}
}

func TestHeroPhase_AllowsCastAndChant(t *testing.T) {
	hp := phase.NewHeroPhase()

	if !hp.IsCommandAllowed(command.CommandTypeCast) {
		t.Error("hero phase should allow cast command")
	}
	if !hp.IsCommandAllowed(command.CommandTypeChant) {
		t.Error("hero phase should allow chant command")
	}
	if !hp.IsCommandAllowed(command.CommandTypeEndPhase) {
		t.Error("hero phase should allow end phase command")
	}
}

func TestGameView_IncludesSpellInfo(t *testing.T) {
	g, _, _ := setupWizardGame(42)

	view := g.View(1)
	units := view.Units[1]
	if len(units) == 0 {
		t.Fatal("expected units for player 1")
	}

	wizardView := units[0]
	if len(wizardView.Spells) != 3 {
		t.Errorf("expected 3 spells in view, got %d", len(wizardView.Spells))
	}
	if !wizardView.CanCast {
		t.Error("wizard should show CanCast = true")
	}
}
