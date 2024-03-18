package rogue

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (rogue *Rogue) registerFanOfKnives() {
	fokSpell := rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 51723},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeOrRangedSpecial,
		Flags:       core.SpellFlagMeleeMetrics | SpellFlagColdBlooded,

		DamageMultiplier: 0.8 * (1 +
			core.TernaryFloat64(rogue.Spec == proto.Spec_SpecCombatRogue, 0.75, 0.0)),
		CritMultiplier:   rogue.MeleeCritMultiplier(false), // TODO (TheBackstabi, 3/16/2024) - Verify what crit table FoK is on
		ThreatMultiplier: 1,
	})

	results := make([]*core.SpellResult, len(rogue.Env.Encounter.TargetUnits))

	rogue.FanOfKnives = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 51723},
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: 35,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return rogue.HasThrown()
		},

		ApplyEffects: func(sim *core.Simulation, unit *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			for i, aoeTarget := range sim.Encounter.TargetUnits {
				baseDamage := fokSpell.Unit.RangedWeaponDamage(sim, fokSpell.RangedAttackPower(aoeTarget))
				baseDamage *= sim.Encounter.AOECapMultiplier()

				results[i] = fokSpell.CalcDamage(sim, aoeTarget, baseDamage, fokSpell.OutcomeRangedHitAndCrit)
			}
			for i, aoeTarget := range sim.Encounter.TargetUnits {
				fokSpell.DealDamage(sim, results[i])

				if rogue.Talents.VilePoisons > 0 {
					poisonProcModifier := 0.33 * float64(rogue.Talents.VilePoisons)
					mhProcChance := poisonProcModifier * getPoisonsProcChance(core.ProcMaskMeleeMH, rogue.Options.MhImbue, rogue)
					ohProcChance := poisonProcModifier * getPoisonsProcChance(core.ProcMaskMeleeOH, rogue.Options.OhImbue, rogue)

					if sim.Proc(mhProcChance, "Vile Poisons FoK MH") {
						switch rogue.Options.MhImbue {
						case proto.RogueOptions_InstantPoison:
							rogue.InstantPoison[VilePoisonsProc].Cast(sim, aoeTarget)
						case proto.RogueOptions_WoundPoison:
							rogue.WoundPoison[VilePoisonsProc].Cast(sim, aoeTarget)
						case proto.RogueOptions_DeadlyPoison:
							rogue.DeadlyPoison.Cast(sim, aoeTarget)
						}
					}
					if sim.Proc(ohProcChance, "Vile Poisons FoK OH") {
						switch rogue.Options.OhImbue {
						case proto.RogueOptions_InstantPoison:
							rogue.InstantPoison[VilePoisonsProc].Cast(sim, aoeTarget)
						case proto.RogueOptions_WoundPoison:
							rogue.WoundPoison[VilePoisonsProc].Cast(sim, aoeTarget)
						case proto.RogueOptions_DeadlyPoison:
							rogue.DeadlyPoison.Cast(sim, aoeTarget)
						}
					}
				}
			}
		},
	})
}

func getPoisonsProcChance(procMask core.ProcMask, imbue proto.RogueOptions_PoisonImbue, rogue *Rogue) float64 {
	switch imbue {
	case proto.RogueOptions_InstantPoison:
		return rogue.instantPoisonPPMM.Chance(procMask)
	case proto.RogueOptions_WoundPoison:
		return rogue.woundPoisonPPMM.Chance(procMask)
	case proto.RogueOptions_DeadlyPoison:
		return rogue.GetDeadlyPoisonProcChance()
	}
	return 0.0
}
