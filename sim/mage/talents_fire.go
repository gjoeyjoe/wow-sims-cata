package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/cata/sim/core"
)

func (mage *Mage) ApplyFireTalents() {
	fmt.Println("ApplyFireTalents() called")
	// Cooldowns/Special Implementations
	//mage.applyHotStreak()
	//mage.applyMoltenFury()
	//mage.applyMasterOfElements()
	//mage.applyPyromaniac()

	// Improved Fire Blast
	if mage.Talents.ImprovedFireBlast > 0 {
		mage.AddStaticMod(core.SpellModConfig{
			ClassMask:  MageSpellFireBlast,
			FloatValue: .04 * float64(mage.Talents.ImprovedFireBlast) * core.CritRatingPerCritChance,
			Kind:       core.SpellMod_BonusCrit_Rating,
		})
	}

	// Fire Power
	if mage.Talents.FirePower > 0 {
		mage.AddStaticMod(core.SpellModConfig{
			School:     core.SpellSchoolFire,
			FloatValue: 0.01 * float64(mage.Talents.FirePower),
			Kind:       core.SpellMod_DamageDone_Flat,
		})
	}

	// Improved Scorch
	if mage.Talents.ImprovedScorch > 0 {
		mage.AddStaticMod(core.SpellModConfig{
			ClassMask:  MageSpellScorch,
			FloatValue: -0.5 * float64(mage.Talents.ImprovedScorch),
			Kind:       core.SpellMod_PowerCost_Pct,
		})
	}

	// Improved Flamestrike
	if mage.Talents.ImprovedFlamestrike > 0 {
		mage.AddStaticMod(core.SpellModConfig{
			ClassMask:  MageSpellFlamestrike,
			FloatValue: -0.5 * float64(mage.Talents.ImprovedFlamestrike),
			Kind:       core.SpellMod_CastTime_Pct,
		})
	}

	// Pyromaniac

	// Critical Mass
	if mage.Talents.CriticalMass > 0 {
		mage.AddStaticMod(core.SpellModConfig{
			ClassMask:  MageSpellLivingBombDot | MageSpellLivingBombExplosion | MageSpellFlameOrb,
			FloatValue: 0.05 * float64(mage.Talents.CriticalMass),
			Kind:       core.SpellMod_DamageDone_Pct,
		})
	}
}

// Master of Elements
func (mage *Mage) applyMasterOfElements() {
	fmt.Println("Registering Master of Elements")
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundCoeff := 0.15 * float64(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29077})

	/* 	core.MakeProcTriggerAura(&mage.Unit, core.ProcTrigger{
		Name: "Master of Elements",
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskSpellDamage,
		Outcome: core.OutcomeLanded,
		Handler: func()
	}) */
	mage.RegisterAura(core.Aura{
		ActionID: core.ActionID{SpellID: 29076},
		Label:    "Master of Elements",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				return
			}
			if spell.CurCast.Cost == 0 {
				return
			}
			if result.DidCrit() {
				if refundCoeff < 0 {
					mage.SpendMana(sim, -1*spell.DefaultCast.Cost*refundCoeff, manaMetrics)
				} else {
					mage.AddMana(sim, spell.DefaultCast.Cost*refundCoeff, manaMetrics)
				}
			}
		},
	})
}

func (mage *Mage) applyHotStreak() {
	if !mage.Talents.HotStreak {
		return
	}

	ImprovedHotStreakProcChance := float64(mage.Talents.ImprovedHotStreak) * 0.5
	BaseHotStreakProcChance := float64(0.25) // Research needed
	t10ProcAura := mage.BloodmagesRegalia2pcAura()

	// Unimproved Hot Streak Proc Aura
	mage.HotStreakAura = mage.GetOrRegisterAura(core.Aura{
		Label:    "Hot Streak",
		ActionID: core.ActionID{SpellID: 48108},
		Duration: time.Second * 10,
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if t10ProcAura != nil {
				t10ProcAura.Activate(sim)
			}
		},
	})

	// Improved Hotstreak Crit Stacking Aura
	mage.hotStreakCritAura = mage.RegisterAura(core.Aura{
		Label:     "Hot Streak Proc Aura",
		ActionID:  core.ActionID{SpellID: 44448}, //, Tag: 1}, Removing Tag gets rid of the (??) in Timeline
		MaxStacks: 2,
		Duration:  time.Hour,
	})

	// Aura to allow the character to track crits
	mage.hotStreakCritListener = mage.RegisterAura(core.Aura{
		Label:    "Hot Streak Trigger",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.Flags.Matches(HotStreakSpells) {
				return
			}

			// Hot Streak Base Talent Proc
			if result.DidCrit() && spell.Flags.Matches(HotStreakSpells) {
				if sim.Proc(BaseHotStreakProcChance, "Hot Streak") {
					if mage.HotStreakAura.IsActive() {
						mage.HotStreakAura.Refresh(sim)
					} else {
						mage.HotStreakAura.Activate(sim)
					}
				}
			}

			// Improved Hot Streak
			if mage.Talents.ImprovedHotStreak > 0 {

				// If you didn't crit, reset your crit counter
				if !result.DidCrit() {
					mage.hotStreakCritAura.SetStacks(sim, 0)
					mage.hotStreakCritAura.Deactivate(sim)
					return
				}

				// If you did crit, check against talents to see if you proc
				// If you proc and had 1 stack, set crit counter to 0 and give hot streak.
				if mage.hotStreakCritAura.GetStacks() == 1 {
					if sim.Proc(ImprovedHotStreakProcChance, "Improved Hot Streak") {
						mage.hotStreakCritAura.SetStacks(sim, 0)
						mage.hotStreakCritAura.Deactivate(sim)

						mage.HotStreakAura.Activate(sim)
					}
				} else {
					mage.hotStreakCritAura.Activate(sim)
					mage.hotStreakCritAura.AddStack(sim)
				}
			}
		},
	})
}

/* func (mage *Mage) applyPyromaniac() {
if mage.Talents.Pyromaniac == 0 {
	return
} */
/*
		pyromaniacMod := mage.AddDynamicMod(core.SpellModConfig{
			ClassMask:  MageSpellsAll,
			FloatValue: -.05 * float64(mage.Talents.Pyromaniac),
			Kind:       core.SpellMod_CastTime_Pct,
		})
		var activeFireDots []*core.Spell

		mage.PyromaniacAura = mage.RegisterAura(core.Aura{
			Label:    "Pyromaniac Trackers",
			ActionID: core.ActionID{SpellID: 83582},
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				if len(sim.AllUnits) < 3 {
					return
				}
				aura.Activate(sim)
			},
			OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {

				 for _, aoeTarget := range sim.Encounter.TargetUnits {
					if mage.LivingBomb.Dot(aoeTarget).RemainingDuration(sim) > 0 {

					}
					if spell.ClassSpellMask == MageSpellFireDoT {
						activeFireDots = append(activeFireDots, spell)
						core.StartDelayedAction(sim, core.DelayedActionOptions{
							DoAt: sim.CurrentTime + spell.Dot(mage.CurrentTarget).RemainingDuration(sim),
							OnAction: func(sim *core.Simulation) {
								l := len(activeFireDots)
								activeFireDots = activeFireDots[:l-1]

							},
						})
					}
				}
				fmt.Println("activeFireDots: ", len(activeFireDots))

				if len(activeFireDots) >= 3 {
					pyromaniacMod.Activate()
				} else {
					pyromaniacMod.Deactivate()
				}
			},
		})
}*/

func (mage *Mage) applyMoltenFury() {
	if mage.Talents.MoltenFury == 0 {
		return
	}

	moltenFuryMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: .04 * float64(mage.Talents.MoltenFury),
		Kind:       core.SpellMod_DamageDone_Pct,
	})

	mage.RegisterResetEffect(func(sim *core.Simulation) {
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute == 35 {
				moltenFuryMod.Activate()

				// For some reason Molten Fury doesn't apply to living bomb DoT, so cancel it out.
				// 4/15/24 - Comment above was from before. Worth checking this out.
				/*if mage.LivingBomb != nil {
					mage.LivingBomb.DamageMultiplier /= multiplier
				}*/
			}
		})
	})
}
