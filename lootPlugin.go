package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

// type LootPlugin Plugin
type LootPlugin struct {
	Plugin
	LootMatch *regexp.Regexp
}

func init() {
	plug := new(LootPlugin)
	plug.Name = "Loot tracking"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = SPELLOUT
	Handlers = append(Handlers, plug)

	plug.LootMatch, _ = regexp.Compile(configuration.Everquest.RegexLoot)
}

// Handle for ParsePlugin sends a message if a parse was pasted to the parse channel
func (p *LootPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "system" {
		match := p.LootMatch.FindStringSubmatch(msg.Msg)
		if len(match) > 0 {
			player := match[1]
			loot := match[2]
			corpse := match[3]
			// fmt.Printf("%#+v\n", loot)
			class := "Unknown"
			if _, ok := roster[player]; ok {
				class = roster[player].Class
			}
			if loot != "" && strings.Contains(loot, "Spell: ") || strings.Contains(loot, "Ancient: ") || isSpellProvider(loot) || isAwardedLoot(loot) {
				// Lookup spell name, and what players need it
				id, _ := itemDB.FindIDByName(loot)
				item, _ := itemDB.GetItemByID(id)
				fmt.Fprintf(out, "> %s (%s) looted %s from %s\n```%s```\n", player, class, loot, corpse, getItemDesc(item)) // TODO: Make this a sexy print with item stats
			}
		}
	}
}

func (p *LootPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *LootPlugin) OutputChannel() int {
	return p.Output
}

func isSpellProvider(item string) bool { // TODO: Add spell replacement options
	for _, sitem := range configuration.Everquest.SpellProvider {
		if item == sitem {
			return true
		}
	}
	return false
}

func isAwardedLoot(item string) bool {
	for _, needs := range needsLooted { // Notify that someone looted a bid upon item
		if strings.EqualFold(needs, item) {
			removeLootFromLooted(needs)
			return true // We only want to remove 1 item per loot (multi bid items we want to see all winners loot them)
		}
	}
	return false
}

func getItemDesc(item everquest.Item) string {
	var desc string
	// Name -- Optional
	desc += item.Name + "\n"
	// Magic / LORE / No-DROP / Temporary
	if item.Magic > 0 {
		desc += "MAGIC "
	}
	if item.Loregroup < 0 {
		desc += "LORE "
	}
	if item.Nodrop == 0 {
		desc += "NO TRADE "
	}
	if item.Augtype > 0 {
		desc += "AUGMENTATION "
	}
	if item.Placeablebitfield > 0 {
		desc += "PLACEABLE "
	}
	// SLOT
	desc += "\nSlot: "
	desc += fmt.Sprintf("%s ", getSlots(uint(item.Slots)))
	// Skill: bla Atk Delay: bla
	if item.Itemtype > 0 && item.Delay > 0 {
		desc += fmt.Sprintf("\nSkill: %s ", itemTypeToString(item.Itemtype))
	}
	if item.Delay > 0 {
		desc += fmt.Sprintf("Atk Delay: %d ", item.Delay)
	}
	if item.Damage > 0 {
		desc += fmt.Sprintf("\nDMG: %d ", item.Damage)
	}
	if item.Backstabdmg > 0 {
		desc += fmt.Sprintf("\nBackstab DMG: %d ", item.Backstabdmg)
	}
	// AC
	if item.Ac > 0 {
		desc += fmt.Sprintf("\nAC: %d", item.Ac)
	}
	// DMG Skill Mod
	if item.Extradmgskill > 0 {
		desc += fmt.Sprintf("\nSkill Add DMG: %s +%d", skillToString(item.Extradmgskill), item.Extradmgamt)
	}
	// Skill Mod
	if item.Skillmodtype > 0 {
		desc += fmt.Sprintf("\nSkill Mod: %s +%d%%", skillToString(item.Skillmodtype), item.Skillmodvalue)
		if item.Skillmodextra > 0 {
			desc += fmt.Sprintf("+%d", item.Skillmodextra)
		}
		desc += fmt.Sprintf(" (%d Max)", item.Skillmodmax)
	}
	// Base Stats
	desc += "\n" // We need to check if there are any base stats
	if item.Astr > 0 || item.HeroicStr > 0 {
		desc += fmt.Sprintf("STR: +%d", item.Astr)
		if item.HeroicStr > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicStr)
		}
		desc += " "
	}
	if item.Adex > 0 || item.HeroicDex > 0 {
		desc += fmt.Sprintf("DEX: +%d", item.Adex)
		if item.HeroicDex > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicDex)
		}
		desc += " "
	}
	if item.Asta > 0 || item.HeroicSta > 0 {
		desc += fmt.Sprintf("STA: +%d", item.Asta)
		if item.HeroicSta > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicSta)
		}
		desc += " "
	}
	if item.Acha > 0 || item.HeroicCha > 0 {
		desc += fmt.Sprintf("CHA: +%d", item.Acha)
		if item.HeroicCha > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicCha)
		}
		desc += " "
	}
	if item.Awis > 0 || item.HeroicWis > 0 {
		desc += fmt.Sprintf("WIS: +%d", item.Awis)
		if item.HeroicWis > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicWis)
		}
		desc += " "
	}
	if item.Aint > 0 || item.HeroicInt > 0 {
		desc += fmt.Sprintf("INT: +%d", item.Aint)
		if item.HeroicInt > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicInt)
		}
		desc += " "
	}
	if item.Aagi > 0 || item.HeroicAgi > 0 {
		desc += fmt.Sprintf("AGI: +%d", item.Aagi)
		if item.HeroicAgi > 0 {
			desc += fmt.Sprintf("+%d", item.HeroicAgi)
		}
		desc += " "
	}
	if item.Hp > 0 {
		desc += fmt.Sprintf("HP: +%d ", item.Hp)
	}
	if item.Mana > 0 {
		desc += fmt.Sprintf("MANA: +%d ", item.Mana)
	}
	if item.Endurance > 0 {
		desc += fmt.Sprintf("ENDUR: +%d ", item.Endurance)
	}
	// Saves
	if item.Fr > 0 || item.Cr > 0 || item.Dr > 0 || item.Mr > 0 || item.Pr > 0 {
		desc += "\n"
	}
	if item.Fr > 0 {
		desc += fmt.Sprintf("SV FIRE: +%d ", item.Fr)
	}
	if item.Dr > 0 {
		desc += fmt.Sprintf("SV DISEASE: +%d ", item.Dr)
	}
	if item.Cr > 0 {
		desc += fmt.Sprintf("SV COLD: +%d ", item.Cr)
	}
	if item.Mr > 0 {
		desc += fmt.Sprintf("SV MAGIC: +%d ", item.Mr)
	}
	if item.Pr > 0 {
		desc += fmt.Sprintf("SV POISON: +%d ", item.Pr)
	}
	// Advanced Stats
	if item.Attack > 0 || item.Regen > 0 || item.Manaregen > 0 || item.Enduranceregen > 0 || item.Haste > 0 {
		desc += "\n"
	}
	if item.Haste > 0 {
		desc += fmt.Sprintf("Haste: +%d%% ", item.Haste)
	}
	if item.Attack > 0 {
		desc += fmt.Sprintf("Attack: +%d ", item.Attack)
	}
	if item.Regen > 0 {
		desc += fmt.Sprintf("HP Regen: +%d ", item.Regen)
	}
	if item.Manaregen > 0 {
		desc += fmt.Sprintf("Mana Regen: +%d ", item.Manaregen)
	}
	if item.Enduranceregen > 0 {
		desc += fmt.Sprintf("Endurance Regen: +%d ", item.Enduranceregen)
	}
	// Levels
	if item.Reclevel > 0 {
		desc += fmt.Sprintf("\nRecommended level of %d ", item.Reclevel)
	}
	if item.Reqlevel > 0 {
		desc += fmt.Sprintf("\nRequired level of %d ", item.Reqlevel)
	}
	// Proc
	if item.Proceffect > 0 {
		effect, _ := spellDB.GetSpellByID(item.Proceffect)
		desc += fmt.Sprintf("\nEffect: %s (Combat, Casting Time: Instant)", effect.Name)
	}
	// Effects
	if item.Clickeffect > 0 {
		effect, _ := spellDB.GetSpellByID(item.Clickeffect)
		desc += fmt.Sprintf("\nEffect: %s ", effect.Name)
	}
	if item.Focuseffect > 0 {
		effect, _ := spellDB.GetSpellByID(item.Focuseffect)
		desc += fmt.Sprintf("\nFocus: %s ", effect.Name)
	}
	desc += "\n"
	desc += fmt.Sprintf("WT: %.1f Size: %s", float32(item.Weight)/10, itemSizeToString(item.Size))
	desc += "\n"
	desc += fmt.Sprintf("Class: %s ", getClasses(uint(item.Classes)))
	desc += "\n"
	desc += fmt.Sprintf("Race: %s ", getRaces(uint(item.Races)))
	// Augs
	if item.Augslot1type > 0 {
		desc += fmt.Sprintf("\nSlot 1, Type %d (%s)", item.Augslot1type, augTypeToString(item.Augslot1type))
	}
	if item.Augslot2type > 0 {
		desc += fmt.Sprintf("\nSlot 2, Type %d (%s)", item.Augslot2type, augTypeToString(item.Augslot2type))
	}
	if item.Augslot3type > 0 {
		desc += fmt.Sprintf("\nSlot 3, Type %d (%s)", item.Augslot3type, augTypeToString(item.Augslot3type))
	}
	if item.Augslot4type > 0 {
		desc += fmt.Sprintf("\nSlot 4, Type %d (%s)", item.Augslot4type, augTypeToString(item.Augslot4type))
	}
	if item.Augslot5type > 0 {
		desc += fmt.Sprintf("\nSlot 5, Type %d (%s)", item.Augslot5type, augTypeToString(item.Augslot5type))
	}
	if item.Augslot6type > 0 {
		desc += fmt.Sprintf("\nSlot 6, Type %d (%s)", item.Augslot6type, augTypeToString(item.Augslot6type))
	}
	return desc
}

func getSlots(bits uint) string {
	const MAXSLOTS = 23
	var result string
	if bits > 4194304 {
		return "ALL"
	}
	if bits <= 0 {
		return "NONE"
	}
	var i uint8
	for i = 0; i < MAXSLOTS; i++ {
		test := bits & (1 << i)
		if test > 0 {
			slot := bitToSlot(test)
			result += slot + " "
		}
	}

	return result
}

func bitToSlot(bit uint) string {
	switch bit {
	case 419304:
		return "POWER"
	case 2097152:
		return "AMMO"
	case 1048576:
		return "WAIST"
	case 524288:
		return "FEET"
	case 262144:
		return "LEGS"
	case 131072:
		return "CHEST"
	case 65536:
		return "" // _finger2_
	case 32768:
		return "FINGER"
	case 16384:
		return "SECONDARY"
	case 8192:
		return "PRIMARY"
	case 4096:
		return "HANDS"
	case 2048:
		return "RANGE"
	case 1024:
		return "" // _wrist2_
	case 512:
		return "WRIST"
	case 256:
		return "BACK"
	case 128:
		return "ARMS"
	case 64:
		return "SHOULDERS"
	case 32:
		return "NECK"
	case 16:
		return "" // _ear2_
	case 8:
		return "FACE"
	case 4:
		return "HEAD"
	case 2:
		return "EAR"
	case 1:
		return "CHARM"
	}
	return ""
}

func itemSizeToString(size int) string {
	switch size {
	case 0:
		return "TINY"
	case 1:
		return "SMALL"
	case 2:
		return "MEDIUM"
	case 3:
		return "LARGE"
	case 4:
		return "GIANT"
	case 5:
		return "GIGANTIC"
	}
	return ""
}

func getClasses(bits uint) string {
	const MAXCLASSES = 16
	var result string
	if bits >= 65535 {
		return "ALL"
	}
	if bits <= 0 {
		return "NONE"
	}
	var i uint8
	for i = 0; i <= MAXCLASSES; i++ {
		test := bits & (1 << i)
		if test > 0 {
			slot := bitToClass(test)
			result += slot + " "
		}
	}

	return result
}

func bitToClass(bit uint) string {
	switch bit {
	case 32768:
		return "BER"
	case 16384:
		return "BST"
	case 8192:
		return "ENC"
	case 4096:
		return "MAG"
	case 2048:
		return "WIZ"
	case 1024:
		return "NEC"
	case 512:
		return "SHM"
	case 256:
		return "ROG"
	case 128:
		return "BRD"
	case 64:
		return "MNK"
	case 32:
		return "DRU"
	case 16:
		return "SHD"
	case 8:
		return "RNG"
	case 4:
		return "PAL"
	case 2:
		return "CLR"
	case 1:
		return "WAR"
	}
	return ""
}

func getRaces(bits uint) string {
	const MAXRACES = 16
	var result string
	if bits >= 65535 {
		return "ALL"
	}
	if bits <= 0 {
		return "NONE"
	}
	var i uint8
	for i = 0; i <= MAXRACES; i++ {
		test := bits & (1 << i)
		if test > 0 {
			slot := bitToRace(test)
			result += slot + " "
		}
	}

	return result
}

func bitToRace(bit uint) string {
	switch bit {
	case 32768:
		return "DRK"
	case 16384:
		return "FRG"
	case 8192:
		return "VAH"
	case 4096:
		return "IKS"
	case 2048:
		return "GNM"
	case 1024:
		return "HFL"
	case 512:
		return "OGR"
	case 256:
		return "TRL"
	case 128:
		return "DWF"
	case 64:
		return "HEF"
	case 32:
		return "DEF"
	case 16:
		return "HIE"
	case 8:
		return "ELF"
	case 4:
		return "ERU"
	case 2:
		return "BAR"
	case 1:
		return "HUM"
	}
	return ""
}

func itemTypeToString(iType int) string {
	switch iType {
	case 0:
		return "1H Slashing"
	case 1:
		return "2H Slashing"
	case 2:
		return "Piercing"
	case 3:
		return "1H Blunt"
	case 4:
		return "2H Blunt"
	case 5:
		return "Archery"
	case 6:
		return "Unused"
	case 7:
		return "Throwing"
	case 8:
		return "Shield"
	case 9:
		return "Unused"
	case 10:
		return "Defence (Armor)"
	case 11:
		return "Involves Tradeskills (Not sure how)"
	case 12:
		return "Lock Picking"
	case 13:
		return "Unused"
	case 14:
		return "Food (Right Click to use)"
	case 15:
		return "Drink (Right Click to use)"
	case 16:
		return "Light Source"
	case 17:
		return "Common Inventory Item"
	case 18:
		return "Bind Wound"
	case 19:
		return "Thrown Casting Items (Explosive potions etc)"
	case 20:
		return "Spells / Song Sheets"
	case 21:
		return "Potions"
	case 22:
		return "Fletched Arrows?..."
	case 23:
		return "Wind Instruments"
	case 24:
		return "Stringed Instruments"
	case 25:
		return "Brass Instruments"
	case 26:
		return "Drum Instruments"
	case 27:
		return "Ammo"
	case 28:
		return "Unused"
	case 29:
		return "Jewlery Items (As far as I can tell)"
	case 30:
		return "Unused"
	case 31:
		return "This note is rolled up"
	case 32:
		return "This book is closed"
	case 33:
		return "Keys"
	case 34:
		return "Odd Items (Not sure what they are for)"
	case 35:
		return "2H Piercing"
	case 36:
		return "Fishing Poles"
	case 37:
		return "Fishing Bait"
	case 38:
		return "Alcoholic Beverages"
	case 39:
		return "More Keys"
	case 40:
		return "Compasses"
	case 41:
		return "Unused"
	case 42:
		return "Poisons"
	case 43:
		return "Unused"
	case 44:
		return "Unused"
	case 45:
		return "Hand to Hand"
	case 46:
		return "Unused"
	case 47:
		return "Unused"
	case 48:
		return "Unused"
	case 49:
		return "Unused"
	case 50:
		return "Unused"
	case 51:
		return "Unused"
	case 52:
		return "Charms"
	case 53:
		return "Dyes"
	case 54:
		return "Augments"
	case 55:
		return "Augment Solvents"
	case 56:
		return "Augment Distillers"
	case 57:
		return "Unknown"
	case 58:
		return "Fellowship Banner Materials"
	case 59:
		return "Cultural Armor Manuals"
	case 60:
		return "New Currencies"
	}
	return ""
}

func skillToString(skill int) string {
	// Take Skill ID and return SKill Name
	/*
			Skill ID
		Skill Name
		0
		1H Blunt
		1
		1H Slashing
		2
		2H Blunt
		3
		2H Slashing
		4
		Abjuration
		5
		Alteration
		6
		Apply Poison
		7
		Archery
		8
		Backstab
		9
		Bind Wound
		10
		Bash
		11
		Block
		12
		Brass Instruments
		13
		Channeling
		14
		Conjuration
		15
		Defense
		16
		Disarm
		17
		Disarm Traps
		18
		Divination
		19
		Dodge
		20
		Double Attack
		21
		Dragon Punch
		22
		Duel Wield
		23
		Eagle Strike
		24
		Evocation
		25
		Feign Death
		26
		Flying Kick
		27
		Forage
		28
		Hand To Hand
		29
		Hide
		30
		Kick
		31
		Meditate
		32
		Mend
		33
		Offense
		34
		Parry
		35
		Pick Lock
		36
		Piercing
		37
		Riposte
		38
		Round Kick
		39
		Safe Fall
		40
		Sense Heading
		41
		Sing
		42
		Sneak
		43
		Specialize Abjure
		44
		Specialize Alteration
		45
		Specialize Conjuration
		46
		Specialize Divinatation
		47
		Specialize Evocation
		48
		Pick Pockets
		49
		Stringed Instruments
		50
		Swimming
		51
		Throwing
		52
		Tiger Claw
		53
		Tracking
		54
		Wind Instruments
		55
		Fishing
		56
		Make Poison
		57
		Tinkering
		58
		Research
		59
		Alchemy
		60
		Baking
		61
		Tailoring
		62
		Sense Traps
		63
		Blacksmithing
		64
		Fletching
		65
		Brewing
		66
		Alcohol Tolerance
		67
		Begging
		68
		Jewelry Making
		69
		Pottery
		70
		Percussion Instruments
		71
		Intimidation
		72
		Berserking
		73
		Taunt
		74
		Frenzy
		75
		Remove Traps
		76
		Triple Attack
		77
		2H Piercing
	*/
	switch skill {
	case 0:
		return "1H Blunt"
	case 1:
		return "1H Slashing"
	case 2:
		return "2H Blunt"
	case 3:
		return "2H Slashing"
	case 4:
		return "Abjuration"
	case 5:
		return "Alteration"
	case 6:
		return "Apply Poison"
	case 7:
		return "Archery"
	case 8:
		return "Backstab"
	case 9:
		return "Bind Wound"
	case 10:
		return "Bash"
	case 11:
		return "Block"
	case 12:
		return "Brass Instruments"
	case 13:
		return "Channeling"
	case 14:
		return "Conjuration"
	case 15:
		return "Defense"
	case 16:
		return "Disarm"
	case 17:
		return "Disarm Traps"
	case 18:
		return "Divination"
	case 19:
		return "Dodge"
	case 20:
		return "Double Attack"
	case 21:
		return "Dragon Punch"
	case 22:
		return "Duel Wield"
	case 23:
		return "Eagle Strike"
	case 24:
		return "Evocation"
	case 25:
		return "Feign Death"
	case 26:
		return "Flying Kick"
	case 27:
		return "Forage"
	case 28:
		return "Hand To Hand"
	case 29:
		return "Hide"
	case 30:
		return "Kick"
	case 31:
		return "Meditate"
	case 32:
		return "Mend"
	case 33:
		return "Offense"
	case 34:
		return "Parry"
	case 35:
		return "Pick Lock"
	case 36:
		return "Piercing"
	case 37:
		return "Riposte"
	case 38:
		return "Round Kick"
	case 39:
		return "Safe Fall"
	case 40:
		return "Sense Heading"
	case 41:
		return "Sing"
	case 42:
		return "Sneak"
	case 43:
		return "Specialize Abjure"
	case 44:
		return "Specialize Alteration"
	case 45:
		return "Specialize Conjuration"
	case 46:
		return "Specialize Divination"
	case 47:
		return "Specialize Evocation"
	case 48:
		return "Pick Pockets"
	case 49:
		return "Stringed Instruments"
	case 50:
		return "Swimming"
	case 51:
		return "Throwing"
	case 52:
		return "Tiger Claw"
	case 53:
		return "Tracking"
	case 54:
		return "Wind Instruments"
	case 55:
		return "Fishing"
	case 56:
		return "Make Poison"
	case 57:
		return "Tinkering"
	case 58:
		return "Research"
	case 59:
		return "Alchemy"
	case 60:
		return "Baking"
	case 61:
		return "Tailoring"
	case 62:
		return "Sense Traps"
	case 63:
		return "Blacksmithing"
	case 64:
		return "Fletching"
	case 65:
		return "Brewing"
	case 66:
		return "Alcohol Tolerance"
	case 67:
		return "Begging"
	case 68:
		return "Jewelry Making"
	case 69:
		return "Pottery"
	case 70:
		return "Percussion Instruments"
	case 71:
		return "Intimidation"
	case 72:
		return "Berserking"
	case 73:
		return "Taunt"
	case 74:
		return "Frenzy"
	case 75:
		return "Remove Traps"
	case 76:
		return "Triple Attack"
	case 77:
		return "2H Piercing"
	}
	return ""
}

func augTypeToString(augType int) string {
	switch augType {
	case 1:
		return "General: Single Stat"
	case 2:
		return "General: Multiple Stats"
	case 3:
		return "General: Spell Effect"
	case 4:
		return "Weapon: General"
	case 5:
		return "General: Multiple Stats"
	case 6:
		return "Weapon: Base Damage"
	case 7:
		return "General: Group"
	case 8:
		return "General: Raid"
	case 9:
		return "General: Dragons Points"
	case 10:
		return "Crafted: Common"
	case 11:
		return "Crafted: Group"
	case 12:
		return "Crafted: Raid"
	case 13:
		return "Energeiac: Group"
	case 14:
		return "Energeiac: Raid"
	case 15:
		return "Emblem"
	case 16:
		return "Cultural: Group"
	case 17:
		return "Cultural: Raid"
	case 18:
		return "Special: Group"
	case 19:
		return "Special: Raid"
	case 20:
		return "Ornamentation"
	case 21:
		return "Special Ornamentation"
	case 22:
		return "Luck"
	}
	return ""
}
