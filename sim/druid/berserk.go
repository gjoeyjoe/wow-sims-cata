package druid

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (druid *Druid) registerBerserkCD() {
	if !druid.Talents.Berserk {
		return
	}

	actionId := core.ActionID{SpellID: 50334}
	glyphBonus := core.TernaryDuration(druid.HasPrimeGlyph(proto.DruidPrimeGlyph_GlyphOfBerserk), time.Second*10.0, 0.0)
	primalMadnessRage := 6.0 * float64(druid.Talents.PrimalMadness)
	var affectedSpells []*DruidSpell

	druid.BerserkAura = druid.RegisterAura(core.Aura{
		Label:    "Berserk",
		ActionID: actionId,
		Duration: (time.Second * 15) + glyphBonus,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice([]*DruidSpell{
				druid.MangleCat,
				druid.FerociousBite,
				druid.Rake,
				druid.Rip,
				druid.SavageRoar,
				druid.SwipeCat,
				druid.Shred,
			}, func(spell *DruidSpell) bool { return spell != nil })
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.CostMultiplier -= 0.5
			}

			if druid.PrimalMadnessAura != nil {
				druid.PrimalMadnessAura.Activate(sim)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.CostMultiplier += 0.5
			}

			if druid.PrimalMadnessAura.IsActive() && !druid.TigersFuryAura.IsActive() {
				druid.PrimalMadnessAura.Deactivate(sim)
			}
		},
	})

	druid.Berserk = druid.RegisterSpell(Cat|Bear, core.SpellConfig{
		ActionID: actionId,
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute * 3,
			},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.BerserkAura.Activate(sim)

			if (primalMadnessRage > 0) && druid.InForm(Bear) {
				druid.AddRage(sim, primalMadnessRage, druid.PrimalMadnessRageMetrics)
			}
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Berserk.Spell,
		Type:  core.CooldownTypeDPS,
	})
}
