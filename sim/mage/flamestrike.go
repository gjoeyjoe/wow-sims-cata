package mage

import (
	"time"

	"github.com/wowsims/cata/sim/core"
)

func (mage *Mage) registerFlamestrikeSpell() {

	mage.Flamestrike = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 2120},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          SpellFlagMage | core.SpellFlagAPL,
		ClassSpellMask: MageSpellFlamestrike,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.30,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*2000 - 1000*time.Duration(mage.Talents.ImprovedFlamestrike),
			},
		},

		DamageMultiplierAdditive: 1 +
			.01*float64(mage.Talents.FirePower),
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: 0.146,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "Flamestrike",
			},
			NumberOfTicks: 4,
			TickLength:    time.Second * 2,
			OnSnapshot: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot, _ bool) {
				target := mage.CurrentTarget
				dot.SnapshotBaseDamage = 0.103 * mage.ScalingBaseDamage
				dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(dot.Spell.Unit.AttackTables[target.UnitIndex], true)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.TargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
				}
			},
			BonusCoefficient: 0.061,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.TargetUnits {
				baseDamage := 0.662 * mage.ScalingBaseDamage
				baseDamage *= sim.Encounter.AOECapMultiplier()
				spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
			spell.AOEDot().Apply(sim)
		},
	})
}
