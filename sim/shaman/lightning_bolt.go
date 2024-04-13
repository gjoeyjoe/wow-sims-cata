package shaman

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (shaman *Shaman) registerLightningBoltSpell() {
	shaman.LightningBolt = shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(false))
	shaman.LightningBoltOverload = shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(true))
}

func (shaman *Shaman) newLightningBoltSpellConfig(isElementalOverload bool) core.SpellConfig {
	castTime := time.Millisecond * 2500
	bonusCoefficient := 0.714
	canOverload := false
	overloadChance := shaman.GetOverloadChance()
	if shaman.Spec == proto.Spec_SpecElementalShaman {
		castTime -= 500
		// 0.36 is shamanism bonus
		bonusCoefficient += 0.36
		canOverload = true
	}

	spellConfig := shaman.newElectricSpellConfig(
		core.ActionID{SpellID: 403},
		0.1,
		castTime,
		isElementalOverload,
		bonusCoefficient)

	spellConfig.ClassSpellMask = SpellMaskLightningBolt

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, 770, spell.OutcomeMagicHitAndCrit)

		if canOverload && result.Landed() && sim.RandomFloat("Lightning Bolt Elemental Overload") < overloadChance {
			shaman.LightningBoltOverload.Cast(sim, target)
		}

		spell.DealDamage(sim, result)
	}

	return spellConfig
}
