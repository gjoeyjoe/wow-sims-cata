package mage

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
	"github.com/wowsims/cata/sim/core/stats"
)

const (
	SpellFlagMage       = core.SpellFlagAgentReserved1
	ArcaneMissileSpells = core.SpellFlagAgentReserved2
	HotStreakSpells     = core.SpellFlagAgentReserved3
	BrainFreezeSpells   = core.SpellFlagAgentReserved4
)

var TalentTreeSizes = [3]int{21, 21, 19}

type Mage struct {
	core.Character

	moltenArmorMod *core.SpellMod

	Talents       *proto.MageTalents
	Options       *proto.MageOptions
	ArcaneOptions *proto.ArcaneMage_Options
	FireOptions   *proto.FireMage_Options
	FrostOptions  *proto.FrostMage_Options

	//waterElemental *WaterElemental
	mirrorImage *MirrorImage
	flameOrb    *FlameOrb

	// Cached values for a few mechanics.
	bonusCritDamage float64

	ArcaneBarrage           *core.Spell
	ArcaneBlast             *core.Spell
	ArcaneExplosion         *core.Spell
	ArcaneMissiles          *core.Spell
	ArcaneMissilesTickSpell *core.Spell
	ArcanePower             *core.Spell
	Blizzard                *core.Spell
	Combustion              *core.Spell
	DeepFreeze              *core.Spell
	Ignite                  *core.Spell
	LivingBomb              *core.Spell
	Fireball                *core.Spell
	FireBlast               *core.Spell
	FlameOrb                *core.Spell
	FlameOrbExplode         *core.Spell
	Flamestrike             *core.Spell
	Freeze                  *core.Spell
	Frostbolt               *core.Spell
	FrostfireBolt           *core.Spell
	FrostfireOrb            *core.Spell
	FrostfireOrbTickSpell   *core.Spell
	IceLance                *core.Spell
	Pyroblast               *core.Spell
	PyroblastDot            *core.Spell
	Scorch                  *core.Spell
	MirrorImage             *core.Spell
	BlastWave               *core.Spell
	DragonsBreath           *core.Spell
	IcyVeins                *core.Spell
	SummonWaterElemental    *core.Spell

	ArcaneBlastAura        *core.Aura
	ArcaneMissilesProcAura *core.Aura
	ArcanePotencyAura      *core.Aura
	ArcanePowerAura        *core.Aura
	BrainFreezeAura        *core.Aura
	MageClearcastingAura   *core.Aura
	CriticalMassAuras      core.AuraArray
	FingersOfFrostAura     *core.Aura
	FlameOrbTimer          *core.Aura
	FrostArmorAura         *core.Aura
	GlyphedFrostArmorPA    *core.PendingAction
	hotStreakCritAura      *core.Aura
	HotStreakAura          *core.Aura
	hotStreakCritListener  *core.Aura
	MageArmorAura          *core.Aura
	MageArmorPA            *core.PendingAction
	PyromaniacAura         *core.Aura

	ScalingBaseDamage float64

	CritDebuffCategories core.ExclusiveCategoryArray
}

func (mage *Mage) GetCharacter() *core.Character {
	return &mage.Character
}

func (mage *Mage) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
}

func (mage *Mage) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	raidBuffs.ArcaneBrilliance = true

	if mage.Talents.ArcaneTactics {
		raidBuffs.ArcaneTactics = true
	}
}

func (mage *Mage) ApplyTalents() {

	mage.ApplyArcaneTalents()
	mage.ApplyFireTalents()
	mage.ApplyFrostTalents()
}

func (mage *Mage) Initialize() {
	mage.registerLivingBombSpell()
	mage.applyArmor()
	mage.applyGlyphs()
	mage.registerArcaneBlastSpell()
	mage.registerArcaneExplosionSpell()
	mage.registerArcaneMissilesSpell()
	mage.registerBlizzardSpell()
	mage.registerDeepFreezeSpell()
	mage.registerFireballSpell()
	mage.registerFireBlastSpell()
	mage.registerFlameOrbSpell()
	mage.registerFlameOrbExplodeSpell()
	mage.registerFlamestrikeSpell()
	mage.registerFrostboltSpell()
	mage.registerFrostfireOrbSpell()
	mage.registerIceLanceSpell()
	mage.registerScorchSpell()
	mage.registerLivingBombSpell()
	mage.registerFrostfireBoltSpell()
	mage.registerEvocation()
	mage.registerManaGemsCD()
	mage.registerMirrorImageCD()
	mage.registerCombustionSpell()
	mage.registerBlastWaveSpell()
	mage.registerDragonsBreathSpell()
	// mage.registerSummonWaterElementalCD()
	mage.registerIgnite()
	mage.applyArcaneMissileProc()
	mage.ScalingBaseDamage = 937.330078125
}

func (mage *Mage) Reset(sim *core.Simulation) {
}

func (mage *Mage) HasPrimeGlyph(glyph proto.MagePrimeGlyph) bool {
	return mage.HasGlyph(int32(glyph))
}

func (mage *Mage) HasMajorGlyph(glyph proto.MageMajorGlyph) bool {
	return mage.HasGlyph(int32(glyph))
}
func (mage *Mage) HasMinorGlyph(glyph proto.MageMinorGlyph) bool {
	return mage.HasGlyph(int32(glyph))
}

func NewMage(character *core.Character, options *proto.Player, mageOptions *proto.MageOptions) *Mage {
	mage := &Mage{
		Character: *character,
		Talents:   &proto.MageTalents{},
		Options:   mageOptions,
	}

	core.FillTalentsProto(mage.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)

	// Molten Armor
	if mage.Options.Armor == proto.MageOptions_MoltenArmor {
		var critToAdd float64
		if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfMoltenArmor) {
			critToAdd = 5 * core.CritRatingPerCritChance
		} else {
			critToAdd = 3 * core.CritRatingPerCritChance
		}
		mage.AddStat(stats.SpellCrit, critToAdd)
		core.MakePermanent(mage.RegisterAura(core.Aura{
			Label:    "Molten Armor",
			ActionID: core.ActionID{SpellID: 30482},
		}))
	}

	// Mage Armor
	if mage.Options.Armor == proto.MageOptions_MageArmor {
		hasGlyph := mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfMageArmor)
		manaRegenPerSecond := mage.MaxMana() * core.TernaryFloat64(hasGlyph, .072, 0.06)
		mage.MageArmorAura = core.MakePermanent(mage.RegisterAura(core.Aura{
			ActionID: core.ActionID{SpellID: 6117},
			Label:    "Mage Armor",
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				mage.MageArmorPA = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period: time.Second * 1,
					OnAction: func(sim *core.Simulation) {
						mage.AddMana(sim, manaRegenPerSecond, mage.NewManaMetrics(core.ActionID{SpellID: 6117}))
					},
				})
			},
		}))
	}
	mage.mirrorImage = mage.NewMirrorImage()
	mage.flameOrb = mage.NewFlameOrb()

	if mage.Spec == proto.Spec_SpecFrostMage {
		//mage.WaterElemental = mage.NewWaterElemental()
	}
	mage.EnableManaBar()
	mage.applyArcaneMissileProc()

	return mage
}

func (mage *Mage) GetMage() *Mage {
	return mage
}

func (mage *Mage) applyArmor() {

}

// Agent is a generic way to access underlying mage on any of the agents.
type MageAgent interface {
	GetMage() *Mage
}

func (mage *Mage) applyArcaneMissileProc() {
	if mage.Talents.HotStreak || mage.Talents.BrainFreeze > 0 {
		return
	}

	t10ProcAura := mage.BloodmagesRegalia2pcAura()

	// Aura for when proc is successful
	mage.ArcaneMissilesProcAura = mage.RegisterAura(core.Aura{
		Label:    "Arcane Missiles Proc",
		ActionID: core.ActionID{SpellID: 79683},
		Duration: time.Second * 20,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell == mage.ArcaneMissiles {
				aura.Deactivate(sim)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if t10ProcAura != nil {
				t10ProcAura.Activate(sim)
			}
		},
	})

	procChance := 0.4

	// Listener for procs
	mage.RegisterAura(core.Aura{
		Label:    "Arcane Missiles Activation",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Flags.Matches(ArcaneMissileSpells) {
				return
			}
			if sim.Proc(procChance, "Arcane Missiles") {
				mage.ArcaneMissilesProcAura.Activate(sim)
			}
		},
	})
}

func (mage *Mage) hasChillEffect(spell *core.Spell) bool {
	return spell == mage.Frostbolt || spell == mage.FrostfireBolt || (spell == mage.Blizzard && mage.Talents.IceShards > 0)
}

const (
	MageSpellFlagNone      int64 = 0
	MageSpellArcaneBarrage int64 = 1 << iota
	MageSpellArcaneBlast
	MageSpellArcaneExplosion
	MageSpellArcanePower
	MageSpellArcaneMissilesCast
	MageSpellArcaneMissilesTick
	MageSpellBlastWave
	MageSpellBlizzard
	MageSpellCombustion
	MageSpellConeOfCold
	MageSpellDeepFreeze
	MageSpellDragonsBreath
	MageSpellEvocation
	MageSpellFireBlast
	MageSpellFireball
	MageSpellFlamestrike
	MageSpellFlameOrb
	MageSpellFocusMagic
	MageSpellFrostbolt
	MageSpellFrostfireBolt
	MageSpellFrostfireOrb
	MageSpellIceLance
	MageSpellIcyVeins
	MageSpellIgnite
	MageSpellLivingBombExplosion
	MageSpellLivingBombDot
	MageSpellManaGems
	MageSpellMirrorImage
	MageSpellPyroblast
	MageSpellPyroblastDot
	MageSpellScorch

	MageSpellLast
	MageSpellsAll       = MageSpellLast<<1 - 1
	MageSpellLivingBomb = MageSpellLivingBombDot | MageSpellLivingBombExplosion
	MageSpellFireDoT    = MageSpellLivingBombDot | MageSpellPyroblastDot | MageSpellIgnite | MageSpellCombustion
	MageSpellFire       = MageSpellBlastWave | MageSpellCombustion | MageSpellDragonsBreath | MageSpellFireball |
		MageSpellFireBlast | MageSpellFlameOrb | MageSpellFlamestrike | MageSpellFrostfireBolt | MageSpellIgnite |
		MageSpellLivingBomb | MageSpellPyroblast | MageSpellScorch
	MageSpellChill        = MageSpellFrostbolt | MageSpellFrostfireBolt
	MageSpellBrainFreeze  = MageSpellFireball | MageSpellFrostfireBolt
	MageSpellsAllDamaging = MageSpellArcaneBarrage | MageSpellArcaneBlast | MageSpellArcaneExplosion | /*MageSpellArcaneMissiles | */ MageSpellBlastWave | MageSpellBlizzard | MageSpellDeepFreeze |
		MageSpellDragonsBreath | MageSpellFireBlast | MageSpellFireball | MageSpellFlamestrike | MageSpellFlameOrb | MageSpellFrostbolt | MageSpellFrostfireBolt |
		MageSpellFrostfireOrb | MageSpellIceLance | MageSpellLivingBombExplosion | MageSpellLivingBombDot | MageSpellPyroblast | MageSpellPyroblastDot | MageSpellScorch
)
