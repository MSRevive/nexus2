// Code generated by entc, DO NOT EDIT.

package ent

import (
	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent/character"
	"github.com/msrevive/nexus2/ent/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	characterFields := schema.Character{}.Fields()
	_ = characterFields
	// characterDescSteamid is the schema descriptor for steamid field.
	characterDescSteamid := characterFields[1].Descriptor()
	// character.SteamidValidator is a validator for the "steamid" field. It is called by the builders before save.
	character.SteamidValidator = characterDescSteamid.Validators[0].(func(uint64) error)
	// characterDescSlot is the schema descriptor for slot field.
	characterDescSlot := characterFields[2].Descriptor()
	// character.SlotValidator is a validator for the "slot" field. It is called by the builders before save.
	character.SlotValidator = characterDescSlot.Validators[0].(func(int) error)
	// characterDescName is the schema descriptor for name field.
	characterDescName := characterFields[3].Descriptor()
	// character.NameValidator is a validator for the "name" field. It is called by the builders before save.
	character.NameValidator = characterDescName.Validators[0].(func(string) error)
	// characterDescGender is the schema descriptor for gender field.
	characterDescGender := characterFields[4].Descriptor()
	// character.GenderValidator is a validator for the "gender" field. It is called by the builders before save.
	character.GenderValidator = characterDescGender.Validators[0].(func(int) error)
	// characterDescRace is the schema descriptor for race field.
	characterDescRace := characterFields[5].Descriptor()
	// character.RaceValidator is a validator for the "race" field. It is called by the builders before save.
	character.RaceValidator = characterDescRace.Validators[0].(func(int) error)
	// characterDescFlags is the schema descriptor for flags field.
	characterDescFlags := characterFields[6].Descriptor()
	// character.DefaultFlags holds the default value on creation for the flags field.
	character.DefaultFlags = characterDescFlags.Default.(string)
	// character.FlagsValidator is a validator for the "flags" field. It is called by the builders before save.
	character.FlagsValidator = characterDescFlags.Validators[0].(func(string) error)
	// characterDescQuickslots is the schema descriptor for quickslots field.
	characterDescQuickslots := characterFields[7].Descriptor()
	// character.DefaultQuickslots holds the default value on creation for the quickslots field.
	character.DefaultQuickslots = characterDescQuickslots.Default.(string)
	// character.QuickslotsValidator is a validator for the "quickslots" field. It is called by the builders before save.
	character.QuickslotsValidator = characterDescQuickslots.Validators[0].(func(string) error)
	// characterDescQuests is the schema descriptor for quests field.
	characterDescQuests := characterFields[8].Descriptor()
	// character.DefaultQuests holds the default value on creation for the quests field.
	character.DefaultQuests = characterDescQuests.Default.(string)
	// character.QuestsValidator is a validator for the "quests" field. It is called by the builders before save.
	character.QuestsValidator = characterDescQuests.Validators[0].(func(string) error)
	// characterDescGuild is the schema descriptor for guild field.
	characterDescGuild := characterFields[9].Descriptor()
	// character.GuildValidator is a validator for the "guild" field. It is called by the builders before save.
	character.GuildValidator = characterDescGuild.Validators[0].(func(string) error)
	// characterDescKills is the schema descriptor for kills field.
	characterDescKills := characterFields[10].Descriptor()
	// character.KillsValidator is a validator for the "kills" field. It is called by the builders before save.
	character.KillsValidator = characterDescKills.Validators[0].(func(int) error)
	// characterDescGold is the schema descriptor for gold field.
	characterDescGold := characterFields[11].Descriptor()
	// character.GoldValidator is a validator for the "gold" field. It is called by the builders before save.
	character.GoldValidator = characterDescGold.Validators[0].(func(int) error)
	// characterDescSkills is the schema descriptor for skills field.
	characterDescSkills := characterFields[12].Descriptor()
	// character.DefaultSkills holds the default value on creation for the skills field.
	character.DefaultSkills = characterDescSkills.Default.(string)
	// character.SkillsValidator is a validator for the "skills" field. It is called by the builders before save.
	character.SkillsValidator = characterDescSkills.Validators[0].(func(string) error)
	// characterDescPets is the schema descriptor for pets field.
	characterDescPets := characterFields[13].Descriptor()
	// character.DefaultPets holds the default value on creation for the pets field.
	character.DefaultPets = characterDescPets.Default.(string)
	// character.PetsValidator is a validator for the "pets" field. It is called by the builders before save.
	character.PetsValidator = characterDescPets.Validators[0].(func(string) error)
	// characterDescHealth is the schema descriptor for health field.
	characterDescHealth := characterFields[14].Descriptor()
	// character.HealthValidator is a validator for the "health" field. It is called by the builders before save.
	character.HealthValidator = characterDescHealth.Validators[0].(func(int) error)
	// characterDescMana is the schema descriptor for mana field.
	characterDescMana := characterFields[15].Descriptor()
	// character.ManaValidator is a validator for the "mana" field. It is called by the builders before save.
	character.ManaValidator = characterDescMana.Validators[0].(func(int) error)
	// characterDescEquipped is the schema descriptor for equipped field.
	characterDescEquipped := characterFields[16].Descriptor()
	// character.DefaultEquipped holds the default value on creation for the equipped field.
	character.DefaultEquipped = characterDescEquipped.Default.(string)
	// character.EquippedValidator is a validator for the "equipped" field. It is called by the builders before save.
	character.EquippedValidator = characterDescEquipped.Validators[0].(func(string) error)
	// characterDescSpells is the schema descriptor for spells field.
	characterDescSpells := characterFields[19].Descriptor()
	// character.DefaultSpells holds the default value on creation for the spells field.
	character.DefaultSpells = characterDescSpells.Default.(string)
	// character.SpellsValidator is a validator for the "spells" field. It is called by the builders before save.
	character.SpellsValidator = characterDescSpells.Validators[0].(func(string) error)
	// characterDescSpellbook is the schema descriptor for spellbook field.
	characterDescSpellbook := characterFields[20].Descriptor()
	// character.DefaultSpellbook holds the default value on creation for the spellbook field.
	character.DefaultSpellbook = characterDescSpellbook.Default.(string)
	// character.SpellbookValidator is a validator for the "spellbook" field. It is called by the builders before save.
	character.SpellbookValidator = characterDescSpellbook.Validators[0].(func(string) error)
	// characterDescBags is the schema descriptor for bags field.
	characterDescBags := characterFields[21].Descriptor()
	// character.DefaultBags holds the default value on creation for the bags field.
	character.DefaultBags = characterDescBags.Default.(string)
	// character.BagsValidator is a validator for the "bags" field. It is called by the builders before save.
	character.BagsValidator = characterDescBags.Validators[0].(func(string) error)
	// characterDescSheaths is the schema descriptor for sheaths field.
	characterDescSheaths := characterFields[22].Descriptor()
	// character.DefaultSheaths holds the default value on creation for the sheaths field.
	character.DefaultSheaths = characterDescSheaths.Default.(string)
	// character.SheathsValidator is a validator for the "sheaths" field. It is called by the builders before save.
	character.SheathsValidator = characterDescSheaths.Validators[0].(func(string) error)
	// characterDescID is the schema descriptor for id field.
	characterDescID := characterFields[0].Descriptor()
	// character.DefaultID holds the default value on creation for the id field.
	character.DefaultID = characterDescID.Default.(func() uuid.UUID)
}
