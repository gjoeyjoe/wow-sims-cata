package rogue

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (rogue *Rogue) registerBackstabSpell() {
	hasGlyph := rogue.HasPrimeGlyph(proto.RoguePrimeGlyph_GlyphOfBackstab)
	baseDamage := RogueBaseDamageScalar * .307
	murderousIntentMetrics := rogue.NewEnergyMetrics(core.ActionID{SpellID: 79132})
	glyphOfBackstabMetrics := rogue.NewEnergyMetrics(core.ActionID{SpellID: 56800})

	rogue.Backstab = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 53},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagIncludeTargetBonusDamage | SpellFlagBuilder | SpellFlagColdBlooded | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost:   rogue.GetGeneratorCostModifier(60 - []float64{0, 7, 14, 20}[rogue.Talents.SlaughterFromTheShadows]),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand)
		},

		BonusCritRating: 10 * core.CritRatingPerCritChance * float64(rogue.Talents.PuncturingWounds),

		// Opportunity and Aggression are additive
		DamageMultiplierAdditive: 1 +
			0.1*float64(rogue.Talents.Opportunity) +
			[]float64{0.0, .07, .14, .20}[rogue.Talents.Aggression],
		// Sinister Calling (Subtlety Spec Passive) is Multiplicative
		DamageMultiplier: 2.07 *
			core.TernaryFloat64(rogue.Spec == proto.Spec_SpecSubtletyRogue, 1.4, 1),
		CritMultiplier:   rogue.MeleeCritMultiplier(true),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := baseDamage +
				spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower()) +
				spell.BonusWeaponDamage()

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
				if result.DidCrit() && hasGlyph {
					rogue.AddEnergy(sim, 5, glyphOfBackstabMetrics)
				}
				if sim.IsExecutePhase35() && rogue.Talents.MurderousIntent > 0 {
					totalRecovery := 15 * rogue.Talents.MurderousIntent
					rogue.AddEnergy(sim, float64(totalRecovery), murderousIntentMetrics)
				}
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
