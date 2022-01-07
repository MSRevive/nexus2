// Code generated by entc, DO NOT EDIT.

package character

import (
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the character type in the database.
	Label = "character"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldSteamid holds the string denoting the steamid field in the database.
	FieldSteamid = "steamid"
	// FieldSlot holds the string denoting the slot field in the database.
	FieldSlot = "slot"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldGender holds the string denoting the gender field in the database.
	FieldGender = "gender"
	// FieldRace holds the string denoting the race field in the database.
	FieldRace = "race"
	// FieldFlags holds the string denoting the flags field in the database.
	FieldFlags = "flags"
	// FieldQuickslots holds the string denoting the quickslots field in the database.
	FieldQuickslots = "quickslots"
	// FieldQuests holds the string denoting the quests field in the database.
	FieldQuests = "quests"
	// FieldGuild holds the string denoting the guild field in the database.
	FieldGuild = "guild"
	// FieldKills holds the string denoting the kills field in the database.
	FieldKills = "kills"
	// FieldGold holds the string denoting the gold field in the database.
	FieldGold = "gold"
	// FieldSkills holds the string denoting the skills field in the database.
	FieldSkills = "skills"
	// FieldPets holds the string denoting the pets field in the database.
	FieldPets = "pets"
	// FieldHealth holds the string denoting the health field in the database.
	FieldHealth = "health"
	// FieldMana holds the string denoting the mana field in the database.
	FieldMana = "mana"
	// FieldEquipped holds the string denoting the equipped field in the database.
	FieldEquipped = "equipped"
	// FieldLefthand holds the string denoting the lefthand field in the database.
	FieldLefthand = "lefthand"
	// FieldRighthand holds the string denoting the righthand field in the database.
	FieldRighthand = "righthand"
	// FieldSpells holds the string denoting the spells field in the database.
	FieldSpells = "spells"
	// FieldSpellbook holds the string denoting the spellbook field in the database.
	FieldSpellbook = "spellbook"
	// FieldBags holds the string denoting the bags field in the database.
	FieldBags = "bags"
	// FieldSheaths holds the string denoting the sheaths field in the database.
	FieldSheaths = "sheaths"
	// Table holds the table name of the character in the database.
	Table = "characters"
)

// Columns holds all SQL columns for character fields.
var Columns = []string{
	FieldID,
	FieldSteamid,
	FieldSlot,
	FieldName,
	FieldGender,
	FieldRace,
	FieldFlags,
	FieldQuickslots,
	FieldQuests,
	FieldGuild,
	FieldKills,
	FieldGold,
	FieldSkills,
	FieldPets,
	FieldHealth,
	FieldMana,
	FieldEquipped,
	FieldLefthand,
	FieldRighthand,
	FieldSpells,
	FieldSpellbook,
	FieldBags,
	FieldSheaths,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// SteamidValidator is a validator for the "steamid" field. It is called by the builders before save.
	SteamidValidator func(uint64) error
	// SlotValidator is a validator for the "slot" field. It is called by the builders before save.
	SlotValidator func(uint8) error
	// NameValidator is a validator for the "name" field. It is called by the builders before save.
	NameValidator func(string) error
	// GenderValidator is a validator for the "gender" field. It is called by the builders before save.
	GenderValidator func(uint8) error
	// RaceValidator is a validator for the "race" field. It is called by the builders before save.
	RaceValidator func(uint8) error
	// DefaultFlags holds the default value on creation for the "flags" field.
	DefaultFlags string
	// FlagsValidator is a validator for the "flags" field. It is called by the builders before save.
	FlagsValidator func(string) error
	// DefaultQuickslots holds the default value on creation for the "quickslots" field.
	DefaultQuickslots string
	// QuickslotsValidator is a validator for the "quickslots" field. It is called by the builders before save.
	QuickslotsValidator func(string) error
	// DefaultQuests holds the default value on creation for the "quests" field.
	DefaultQuests string
	// QuestsValidator is a validator for the "quests" field. It is called by the builders before save.
	QuestsValidator func(string) error
	// GuildValidator is a validator for the "guild" field. It is called by the builders before save.
	GuildValidator func(string) error
	// KillsValidator is a validator for the "kills" field. It is called by the builders before save.
	KillsValidator func(int16) error
	// GoldValidator is a validator for the "gold" field. It is called by the builders before save.
	GoldValidator func(uint32) error
	// DefaultSkills holds the default value on creation for the "skills" field.
	DefaultSkills string
	// SkillsValidator is a validator for the "skills" field. It is called by the builders before save.
	SkillsValidator func(string) error
	// DefaultPets holds the default value on creation for the "pets" field.
	DefaultPets string
	// PetsValidator is a validator for the "pets" field. It is called by the builders before save.
	PetsValidator func(string) error
	// HealthValidator is a validator for the "health" field. It is called by the builders before save.
	HealthValidator func(int) error
	// ManaValidator is a validator for the "mana" field. It is called by the builders before save.
	ManaValidator func(int) error
	// DefaultEquipped holds the default value on creation for the "equipped" field.
	DefaultEquipped string
	// EquippedValidator is a validator for the "equipped" field. It is called by the builders before save.
	EquippedValidator func(string) error
	// LefthandValidator is a validator for the "lefthand" field. It is called by the builders before save.
	LefthandValidator func(string) error
	// RighthandValidator is a validator for the "righthand" field. It is called by the builders before save.
	RighthandValidator func(string) error
	// DefaultSpells holds the default value on creation for the "spells" field.
	DefaultSpells string
	// SpellsValidator is a validator for the "spells" field. It is called by the builders before save.
	SpellsValidator func(string) error
	// DefaultSpellbook holds the default value on creation for the "spellbook" field.
	DefaultSpellbook string
	// SpellbookValidator is a validator for the "spellbook" field. It is called by the builders before save.
	SpellbookValidator func(string) error
	// DefaultBags holds the default value on creation for the "bags" field.
	DefaultBags string
	// BagsValidator is a validator for the "bags" field. It is called by the builders before save.
	BagsValidator func(string) error
	// DefaultSheaths holds the default value on creation for the "sheaths" field.
	DefaultSheaths string
	// SheathsValidator is a validator for the "sheaths" field. It is called by the builders before save.
	SheathsValidator func(string) error
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)
