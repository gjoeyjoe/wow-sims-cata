package mage

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
)

func (mage *Mage) applyGlyphs() {
	//Primes
	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfArcaneBarrage) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			ClassMask:  MageSpellArcaneBarrage,
			FloatValue: 0.04,
		})
	}

	// Arcane Blast handled in spell due to handling stacks

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfArcaneMissiles) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Rating,
			ClassMask:  MageSpellArcaneMissilesTick,
			FloatValue: 5 * core.CritRatingPerCritChance,
		})
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfConeOfCold) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			ClassMask:  MageSpellConeOfCold,
			FloatValue: 0.25,
		})
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfDeepFreeze) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			ClassMask:  MageSpellDeepFreeze,
			FloatValue: 0.2,
		})
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfFireball) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Rating,
			ClassMask:  MageSpellFireball,
			FloatValue: 5 * core.CritRatingPerCritChance,
		})
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfFrostbolt) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Rating,
			ClassMask:  MageSpellFrostbolt,
			FloatValue: 5 * core.CritRatingPerCritChance,
		})
	}

	//Frostfire bolt handled inside spell due to changing behavior

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfIceLance) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			ClassMask:  MageSpellIceLance,
			FloatValue: .05,
		})
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfLivingBomb) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			ClassMask:  MageSpellLivingBomb,
			FloatValue: .03,
		})
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfMageArmor) && mage.Options.Armor == proto.MageOptions_MoltenArmor {
		mage.moltenArmorMod.UpdateFloatValue(5 * core.CritRatingPerCritChance)
	}

	if mage.HasPrimeGlyph(proto.MagePrimeGlyph_GlyphOfPyroblast) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Rating,
			ClassMask:  MageSpellPyroblast | MageSpellPyroblastDot,
			FloatValue: 5 * core.CritRatingPerCritChance,
		})
	}

	// Majors

	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfArcanePower) {
		core.MakeProcTriggerAura(&mage.Unit, core.ProcTrigger{
			Name:           "Arcane Power Mirror Image GCD Reduction",
			Callback:       core.CallbackOnCastComplete,
			ClassSpellMask: MageSpellArcanePower,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.ClassSpellMask == MageSpellArcanePower {
					mage.MirrorImage.DefaultCast.GCD = 0
				}
				core.StartDelayedAction(sim, core.DelayedActionOptions{
					DoAt: sim.CurrentTime + mage.ArcanePowerAura.Duration,
					OnAction: func(*core.Simulation) {
						mage.MirrorImage.DefaultCast.GCD = core.GCDDefault
					},
				})
			},
		})
	}

	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfDragonSBreath) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:      core.SpellMod_Cooldown_Flat,
			ClassMask: MageSpellDragonsBreath,
			TimeValue: -3 * time.Second,
		})
	}

	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfDragonSBreath) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:      core.SpellMod_Cooldown_Flat,
			ClassMask: MageSpellDragonsBreath,
			TimeValue: -3 * time.Second,
		})
	}

	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfFrostArmor) && mage.Options.Armor == proto.MageOptions_FrostArmor {
		mage.FrostArmorAura = core.MakePermanent(mage.RegisterAura(core.Aura{
			ActionID: core.ActionID{SpellID: 7302},
			Label:    "Frost Armor",
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				mage.GlyphedFrostArmorPA = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period: time.Second * 1,
					OnAction: func(sim *core.Simulation) {
						mage.AddMana(sim, 0.02*mage.MaxMana(), mage.NewManaMetrics(core.ActionID{SpellID: 6117}))
					},
				})
			},
		}))
	}

	// Minors

	// Mirror Images added inside pet's rotation

}
