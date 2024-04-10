package shaman

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (shaman *Shaman) registerThunderstormSpell() {
	if shaman.Spec != proto.Spec_SpecElementalShaman {
		return
	}

	actionID := core.ActionID{SpellID: 51490}
	manaMetrics := shaman.NewManaMetrics(actionID)

	manaRestore := 0.08
	if shaman.HasMinorGlyph(proto.ShamanMinorGlyph_GlyphOfThunderstorm) {
		manaRestore = 0.02
	}

	cooldown := time.Second * 45
	if shaman.HasMajorGlyph(proto.ShamanMajorGlyph_GlyphOfThunder) {
		cooldown = 35
	}

	shaman.Thunderstorm = shaman.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		Flags:       core.SpellFlagAPL,
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellDamage,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.0,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: cooldown,
			},
		},

		DamageMultiplier: 1 + 0.02*float64(shaman.Talents.Concussion),
		CritMultiplier:   shaman.ElementalFuryCritMultiplier(0),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			shaman.AddMana(sim, shaman.MaxMana()*manaRestore, manaMetrics)

			if shaman.thunderstormInRange {
				dmgFromSP := 0.571 * spell.SpellPower()
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					baseDamage := 1637 + dmgFromSP
					baseDamage *= sim.Encounter.AOECapMultiplier()
					spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
				}
			}
		},
	})
}
