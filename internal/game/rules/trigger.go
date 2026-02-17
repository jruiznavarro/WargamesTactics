package rules

// Trigger identifies the point in the game loop where a rule is evaluated.
type Trigger int

const (
	// Combat pipeline triggers (applied per weapon resolution)
	BeforeAttackCount  Trigger = iota // Modify number of attacks
	BeforeHitRoll                     // Modify hit roll threshold
	BeforeWoundRoll                   // Modify wound roll threshold
	BeforeSaveRoll                    // Modify save roll threshold (cover, etc.)
	BeforeDamage                      // Modify damage per unsaved wound
	AfterCombatResolve                // After full combat resolution (triggers on kill, etc.)

	// Movement triggers
	BeforeMove     // Modify movement distance or block movement
	AfterMove      // After movement completes (deadly terrain, etc.)
	BeforePileIn   // Modify pile-in distance
	BeforeCharge   // Modify charge roll or block charge

	// Shooting triggers
	BeforeShoot // Check visibility, modify range, block shooting

	// Phase triggers
	OnPhaseStart // Start of any phase
	OnPhaseEnd   // End of any phase

	// Ward save trigger
	BeforeWardSave // Modify ward save value dynamically

	// Game event triggers
	OnModelSlain    // When a model is killed
	OnUnitDestroyed // When a unit is fully destroyed
	OnBattleRoundStart // When a new battle round begins
)
