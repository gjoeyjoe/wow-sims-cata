package subtlety

import (
	"time"

	"github.com/wowsims/cata/sim/core"
)

func (subRogue *SubtletyRogue) registerPreparationCD() {
	if !subRogue.Talents.Preparation {
		return
	}

	subRogue.Preparation = subRogue.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 14185},
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    subRogue.NewTimer(),
				Duration: time.Minute * 5,
			},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// Spells affected by Preparation are: Shadowstep, Vanish (Overkill/Master of Subtlety), Sprint
			// If Glyph of Preparation is applied, Smoke Bomb, Dismantle, and Kick are also affected
			var affectedSpells = []*core.Spell{subRogue.Shadowstep, subRogue.Vanish}
			// Reset Cooldown on affected spells
			for _, affectedSpell := range affectedSpells {
				if affectedSpell != nil {
					affectedSpell.CD.Reset()
				}
			}
		},
	})

	subRogue.AddMajorCooldown(core.MajorCooldown{
		Spell:    subRogue.Preparation,
		Type:     core.CooldownTypeDPS,
		Priority: core.CooldownPriorityDefault,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return !subRogue.Vanish.CD.IsReady(sim)
		},
	})
}
