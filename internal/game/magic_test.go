package game

import (
	"strings"
	"testing"

	"github.com/jruiznavarro/wargamestactics/internal/game/command"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
	"github.com/jruiznavarro/wargamestactics/internal/game/phase"
)

// Helper: create a game with a Wizard and an enemy unit for spell tests.
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
	wizard.Spells = core.DefaultWizardSpells()

	enemy := g.CreateUnit("Target Squad", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 3, core.Position{X: 20, Y: 12}, 1.0)

	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero
	return g, wizard, enemy
}

// Helper: create a game with a Priest and units for prayer tests.
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
	priest.Prayers = core.DefaultPriestPrayers()

	enemy := g.CreateUnit("Enemy Squad", 2,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 3},
		nil, 3, core.Position{X: 20, Y: 12}, 1.0)

	g.Commands.InitRound([]int{1, 2}, 4, -1)
	g.CurrentPhase = phase.PhaseHero
	return g, priest, enemy
}

func TestCastSpell_ArcaneBolt(t *testing.T) {
	// Try many seeds to find one where casting succeeds
	for seed := int64(1); seed < 100; seed++ {
		g, wizard, enemy := setupWizardGame(seed)
		woundsBefore := enemy.TotalCurrentWounds()

		cmd := &command.CastCommand{
			OwnerID:    1,
			CasterID:   wizard.ID,
			SpellIndex: 0, // Arcane Bolt
			TargetID:   enemy.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			// Spell succeeded - enemy should have taken mortal damage
			woundsAfter := enemy.TotalCurrentWounds()
			if woundsAfter >= woundsBefore {
				t.Errorf("seed %d: Arcane Bolt succeeded but no damage dealt", seed)
			}
			return // Found a working seed
		}
		// Casting failed or was unbound - try next seed
	}
	t.Fatal("no seed found where Arcane Bolt succeeds within 100 attempts")
}

func TestCastSpell_MysticShield(t *testing.T) {
	// Find a seed where casting Mystic Shield succeeds on a friendly unit
	for seed := int64(1); seed < 100; seed++ {
		g, wizard, _ := setupWizardGame(seed)

		// Create a friendly target
		friendly := g.CreateUnit("Friendly Warriors", 1,
			core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
			nil, 3, core.Position{X: 15, Y: 12}, 1.0)

		cmd := &command.CastCommand{
			OwnerID:    1,
			CasterID:   wizard.ID,
			SpellIndex: 1, // Mystic Shield
			TargetID:   friendly.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			// Mystic Shield adds a +1 save rule
			foundLog := false
			for _, msg := range g.Log {
				if strings.Contains(msg, "Mystic Shield") && strings.Contains(msg, "+1 to save") {
					foundLog = true
					break
				}
			}
			if !foundLog {
				t.Error("expected log message about Mystic Shield +1 save")
			}
			return
		}
	}
	t.Fatal("no seed found where Mystic Shield succeeds within 100 attempts")
}

func TestCastSpell_FailsBelowCastingValue(t *testing.T) {
	// Find a seed where casting roll < 5
	for seed := int64(1); seed < 200; seed++ {
		g, wizard, enemy := setupWizardGame(seed)

		cmd := &command.CastCommand{
			OwnerID:    1,
			CasterID:   wizard.ID,
			SpellIndex: 0, // Arcane Bolt (CV 5)
			TargetID:   enemy.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Success && strings.Contains(result.Description, "failed to cast") {
			return // Found a failing cast
		}
	}
	t.Fatal("no seed found where casting fails within 200 attempts")
}

func TestCastSpell_NotAWizard(t *testing.T) {
	g, _, enemy := setupWizardGame(42)

	// Create a non-wizard unit
	warrior := g.CreateUnit("Warrior", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.CastCommand{
		OwnerID:    1,
		CasterID:   warrior.ID,
		SpellIndex: 0,
		TargetID:   enemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when non-wizard tries to cast")
	}
}

func TestCastSpell_CantCastTwice(t *testing.T) {
	// Default CastsPerTurn = 1, so second cast should fail
	for seed := int64(1); seed < 100; seed++ {
		g, wizard, enemy := setupWizardGame(seed)

		cmd := &command.CastCommand{
			OwnerID:    1,
			CasterID:   wizard.ID,
			SpellIndex: 0,
			TargetID:   enemy.ID,
		}

		// First cast
		_, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("unexpected error on first cast: %v", err)
		}

		// Second cast should fail
		_, err = g.ExecuteCommand(cmd)
		if err == nil {
			t.Error("expected error on second cast (max 1 per turn)")
		}
		return // Only need to test once
	}
}

func TestCastSpell_OutOfRange(t *testing.T) {
	g, wizard, _ := setupWizardGame(42)

	// Create a far-away enemy
	farEnemy := g.CreateUnit("Far Enemy", 2,
		core.Stats{Move: 5, Save: 4, Health: 3},
		nil, 1, core.Position{X: 40, Y: 12}, 1.0)

	cmd := &command.CastCommand{
		OwnerID:    1,
		CasterID:   wizard.ID,
		SpellIndex: 0, // Arcane Bolt range 18"
		TargetID:   farEnemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error for target out of spell range")
	}
}

func TestCastSpell_WrongTargetOwnership(t *testing.T) {
	g, wizard, enemy := setupWizardGame(42)

	// Mystic Shield targets friendly, so targeting enemy should fail
	cmd := &command.CastCommand{
		OwnerID:    1,
		CasterID:   wizard.ID,
		SpellIndex: 1, // Mystic Shield (friendly target)
		TargetID:   enemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when casting friendly spell on enemy")
	}
}

func TestCastSpell_Unbind(t *testing.T) {
	// Place an enemy wizard to attempt unbind
	for seed := int64(1); seed < 500; seed++ {
		g, wizard, enemy := setupWizardGame(seed)

		// Add enemy wizard within 30"
		enemyWizard := g.CreateUnit("Enemy Wizard", 2,
			core.Stats{Move: 5, Save: 5, Control: 1, Health: 5},
			nil, 1, core.Position{X: 25, Y: 12}, 1.0)
		enemyWizard.Keywords = []core.Keyword{core.KeywordHero, core.KeywordWizard}
		enemyWizard.Spells = core.DefaultWizardSpells()

		cmd := &command.CastCommand{
			OwnerID:    1,
			CasterID:   wizard.ID,
			SpellIndex: 0,
			TargetID:   enemy.ID,
		}

		result, _ := g.ExecuteCommand(cmd)

		if !result.Success && strings.Contains(result.Description, "unbound") {
			// Verify enemy wizard used their unbind
			if enemyWizard.UnbindCount != 1 {
				t.Error("expected enemy wizard unbind count to be 1")
			}
			return
		}
	}
	t.Fatal("no seed found where unbind succeeds within 500 attempts")
}

func TestCastSpell_EmpoweredCannotBeUnbound(t *testing.T) {
	// Empowered (doubles) cannot be unbound
	for seed := int64(1); seed < 1000; seed++ {
		g, wizard, enemy := setupWizardGame(seed)

		// Add enemy wizard
		enemyWizard := g.CreateUnit("Enemy Wizard", 2,
			core.Stats{Move: 5, Save: 5, Control: 1, Health: 5},
			nil, 1, core.Position{X: 25, Y: 12}, 1.0)
		enemyWizard.Keywords = []core.Keyword{core.KeywordHero, core.KeywordWizard}
		enemyWizard.Spells = core.DefaultWizardSpells()

		cmd := &command.CastCommand{
			OwnerID:    1,
			CasterID:   wizard.ID,
			SpellIndex: 0,
			TargetID:   enemy.ID,
		}

		g.ExecuteCommand(cmd)

		// Check if it was empowered
		empowered := false
		for _, msg := range g.Log {
			if strings.Contains(msg, "EMPOWERED") {
				empowered = true
				break
			}
		}

		if empowered {
			// Enemy wizard should NOT have used unbind
			if enemyWizard.UnbindCount != 0 {
				t.Error("empowered spell should not be unbindable, but enemy wizard attempted unbind")
			}
			return
		}
	}
	t.Fatal("no seed found where empowered cast occurs within 1000 attempts")
}

func TestChantPrayer_Smite(t *testing.T) {
	for seed := int64(1); seed < 100; seed++ {
		g, priest, enemy := setupPriestGame(seed)
		woundsBefore := enemy.TotalCurrentWounds()

		cmd := &command.ChantCommand{
			OwnerID:     1,
			ChanterID:   priest.ID,
			PrayerIndex: 1, // Smite
			TargetID:    enemy.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			woundsAfter := enemy.TotalCurrentWounds()
			if woundsAfter >= woundsBefore {
				t.Errorf("seed %d: Smite succeeded but no damage dealt", seed)
			}
			return
		}
	}
	t.Fatal("no seed found where Smite succeeds within 100 attempts")
}

func TestChantPrayer_Heal(t *testing.T) {
	for seed := int64(1); seed < 100; seed++ {
		g, priest, _ := setupPriestGame(seed)

		// Create a wounded friendly unit
		friendly := g.CreateUnit("Wounded Warriors", 1,
			core.Stats{Move: 5, Save: 4, Control: 1, Health: 5},
			nil, 1, core.Position{X: 12, Y: 12}, 1.0)
		friendly.Models[0].CurrentWounds = 1 // 4 damage taken

		cmd := &command.ChantCommand{
			OwnerID:     1,
			ChanterID:   priest.ID,
			PrayerIndex: 0, // Heal
			TargetID:    friendly.ID,
		}

		result, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("seed %d: unexpected error: %v", seed, err)
		}

		if result.Success {
			if friendly.Models[0].CurrentWounds <= 1 {
				t.Errorf("seed %d: Heal succeeded but wounds not restored", seed)
			}
			return
		}
	}
	t.Fatal("no seed found where Heal succeeds within 100 attempts")
}

func TestChantPrayer_NotAPriest(t *testing.T) {
	g, _, enemy := setupPriestGame(42)

	warrior := g.CreateUnit("Warrior", 1,
		core.Stats{Move: 5, Save: 4, Control: 1, Health: 2},
		nil, 1, core.Position{X: 10, Y: 10}, 1.0)

	cmd := &command.ChantCommand{
		OwnerID:     1,
		ChanterID:   warrior.ID,
		PrayerIndex: 0,
		TargetID:    enemy.ID,
	}

	_, err := g.ExecuteCommand(cmd)
	if err == nil {
		t.Error("expected error when non-priest tries to chant")
	}
}

func TestChantPrayer_CantChantTwice(t *testing.T) {
	for seed := int64(1); seed < 100; seed++ {
		g, priest, enemy := setupPriestGame(seed)

		cmd := &command.ChantCommand{
			OwnerID:     1,
			ChanterID:   priest.ID,
			PrayerIndex: 1, // Smite
			TargetID:    enemy.ID,
		}

		_, err := g.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("unexpected error on first chant: %v", err)
		}

		_, err = g.ExecuteCommand(cmd)
		if err == nil {
			t.Error("expected error on second chant (max 1 per turn)")
		}
		return
	}
}

func TestCanCast_DefaultsToOneCast(t *testing.T) {
	wizard := &core.Unit{
		Keywords: []core.Keyword{core.KeywordWizard},
		Spells:   core.DefaultWizardSpells(),
	}

	if !wizard.CanCast() {
		t.Error("wizard with spells should be able to cast")
	}

	wizard.CastCount = 1
	if wizard.CanCast() {
		t.Error("wizard should not cast after using default 1 cast")
	}
}

func TestCanCast_MultipleCasts(t *testing.T) {
	wizard := &core.Unit{
		Keywords:     []core.Keyword{core.KeywordWizard},
		Spells:       core.DefaultWizardSpells(),
		CastsPerTurn: 2,
	}

	wizard.CastCount = 1
	if !wizard.CanCast() {
		t.Error("wizard with CastsPerTurn=2 should still be able to cast after 1")
	}

	wizard.CastCount = 2
	if wizard.CanCast() {
		t.Error("wizard should not cast after 2 of 2 casts")
	}
}

func TestCanChant_DefaultsToOneChant(t *testing.T) {
	priest := &core.Unit{
		Keywords: []core.Keyword{core.KeywordPriest},
		Prayers:  core.DefaultPriestPrayers(),
	}

	if !priest.CanChant() {
		t.Error("priest with prayers should be able to chant")
	}

	priest.ChantCount = 1
	if priest.CanChant() {
		t.Error("priest should not chant after using default 1 chant")
	}
}

func TestCanUnbind(t *testing.T) {
	wizard := &core.Unit{
		Keywords: []core.Keyword{core.KeywordWizard},
		Models: []core.Model{
			{ID: 0, IsAlive: true, CurrentWounds: 5, MaxWounds: 5},
		},
	}

	if !wizard.CanUnbind() {
		t.Error("wizard should be able to unbind")
	}

	wizard.UnbindCount = 1
	if wizard.CanUnbind() {
		t.Error("wizard should not unbind twice")
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

func TestDefaultWizardSpells(t *testing.T) {
	spells := core.DefaultWizardSpells()
	if len(spells) != 2 {
		t.Fatalf("expected 2 default spells, got %d", len(spells))
	}

	bolt := spells[0]
	if bolt.Name != "Arcane Bolt" {
		t.Errorf("expected Arcane Bolt, got %s", bolt.Name)
	}
	if bolt.CastingValue != 5 {
		t.Errorf("expected CV 5, got %d", bolt.CastingValue)
	}
	if bolt.Range != 18 {
		t.Errorf("expected range 18, got %d", bolt.Range)
	}
	if bolt.TargetFriendly {
		t.Error("Arcane Bolt should target enemies")
	}

	shield := spells[1]
	if shield.Name != "Mystic Shield" {
		t.Errorf("expected Mystic Shield, got %s", shield.Name)
	}
	if !shield.TargetFriendly {
		t.Error("Mystic Shield should target friendlies")
	}
}

func TestDefaultPriestPrayers(t *testing.T) {
	prayers := core.DefaultPriestPrayers()
	if len(prayers) != 2 {
		t.Fatalf("expected 2 default prayers, got %d", len(prayers))
	}

	heal := prayers[0]
	if heal.Name != "Heal" {
		t.Errorf("expected Heal, got %s", heal.Name)
	}
	if heal.ChantingValue != 4 {
		t.Errorf("expected chanting value 4, got %d", heal.ChantingValue)
	}
	if !heal.TargetFriendly {
		t.Error("Heal should target friendlies")
	}

	smite := prayers[1]
	if smite.Name != "Smite" {
		t.Errorf("expected Smite, got %s", smite.Name)
	}
	if smite.TargetFriendly {
		t.Error("Smite should target enemies")
	}
}

func TestResetPhaseFlags_ResetsMagicCounters(t *testing.T) {
	u := &core.Unit{
		CastCount:   1,
		ChantCount:  1,
		UnbindCount: 1,
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
}

func TestGameView_IncludesSpellInfo(t *testing.T) {
	g, wizard, _ := setupWizardGame(42)
	_ = wizard

	view := g.View(1)
	units := view.Units[1]
	if len(units) == 0 {
		t.Fatal("expected units for player 1")
	}

	wizardView := units[0]
	if len(wizardView.Spells) != 2 {
		t.Errorf("expected 2 spells in view, got %d", len(wizardView.Spells))
	}
	if !wizardView.CanCast {
		t.Error("wizard should show CanCast = true")
	}
}

func TestAnswerPrayer(t *testing.T) {
	// Place an enemy priest to attempt answer
	for seed := int64(1); seed < 500; seed++ {
		g, priest, enemy := setupPriestGame(seed)

		// Add enemy priest within 30"
		enemyPriest := g.CreateUnit("Enemy Priest", 2,
			core.Stats{Move: 5, Save: 4, Control: 1, Health: 5},
			nil, 1, core.Position{X: 25, Y: 12}, 1.0)
		enemyPriest.Keywords = []core.Keyword{core.KeywordHero, core.KeywordPriest}
		enemyPriest.Prayers = core.DefaultPriestPrayers()

		cmd := &command.ChantCommand{
			OwnerID:     1,
			ChanterID:   priest.ID,
			PrayerIndex: 1, // Smite
			TargetID:    enemy.ID,
		}

		result, _ := g.ExecuteCommand(cmd)

		if !result.Success && strings.Contains(result.Description, "answered") {
			return // Found a prayer that was denied
		}
	}
	t.Fatal("no seed found where prayer is answered within 500 attempts")
}
