package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

var needsLooted []string

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
	seedInferredItems()

	plug.LootMatch, _ = regexp.Compile(configuration.Everquest.RegexLoot)
}

// Handle for ParsePlugin sends a message if a parse was pasted to the parse channel
func (p *LootPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "system" {
		match := p.LootMatch.FindStringSubmatch(msg.Msg)
		if len(match) > 0 {
			player := match[1]
			if player == "You" {
				player = getPlayerName(configuration.Everquest.LogPath)
			}
			loot := match[2]
			corpse := match[3]
			// fmt.Printf("%#+v\n", loot)
			class := "Unknown"
			if _, ok := Roster[player]; ok {
				class = Roster[player].Class
			}
			if loot != "" && strings.Contains(loot, "Spell: ") || strings.Contains(loot, "Ancient: ") || isSpellProvider(loot) || isAwardedLoot(loot) {
				// Lookup spell name, and what players need it
				loot = inferLoot(class, loot) // Check if item results in a class specific item, and replace it here.
				id, _ := itemDB.FindIDByName(loot)
				item, _ := itemDB.GetItemByID(id)
				fmt.Fprintf(out, "> %s (%s) looted %s from %s\n```%s```\n", player, class, item.Name, corpse, getItemDesc(item)) // TODO: Make this a sexy print with item stats
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

func removeLootFromLooted(item string) {
	var itemPos int
	for pos, name := range needsLooted {
		if name == item {
			itemPos = pos
		}
	}
	needsLooted = append(needsLooted[:itemPos], needsLooted[itemPos+1:]...)
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

func inferLoot(class string, loot string) string {
	for item, infer := range inferredItems {
		if strings.EqualFold(item, loot) {
			for _, inferClass := range infer {
				if strings.EqualFold(class, inferClass.Class) {
					return inferClass.Item
				}
			}
			return loot // Couldn't find usable class, don't waste cycles
		}
	}
	return loot
}

type InferredItem struct {
	Class string
	Item  string
}

var inferredItems map[string][]InferredItem

func seedInferredItems() {
	inferredItems = make(map[string][]InferredItem)
	inferredItems["Timeless Breastplate Mold"] = []InferredItem{
		{"Bard", "Rizlona's Fiery Chestplate"},
		{"Cleric", "Ultor's Chestguard of Faith"},
		{"Paladin", "Trydan's Chestplate of Nobility"},
		{"Shadow Knight", "Grimror's Guard of the Plagues"},
		{"Warrior", "Raex's Chestplate of Destruction"},
	}
	inferredItems["Timeless Leather Tunic Pattern"] = []InferredItem{
		{"Beastlord", "Dumul's Chestwraps of the Brute"},
		{"Druid", "Kerasha's Sylvan Tunic"},
		{"Monk", "Ton Po's Chestwraps of Composure"},
	}
	inferredItems["Timeless Chain Tunic Pattern"] = []InferredItem{
		{"Berserker", "Galladan's Stormwrath Tunic"},
		{"Ranger", "Askr's Thunderous Chainmail"},
		{"Rogue", "Bidilis' Hauberk of the Elusive"},
		{"Shaman", "Rosrak's Hauberk of the Primal"},
	}
	inferredItems["Timeless Silk Robe Pattern"] = []InferredItem{
		{"Enchanter", "Romar's Robe of Visions"},
		{"Magician", "Magi`Kot's Robe of Convergence"},
		{"Necromancer", "Miragul's Shroud of Risen Souls"},
		{"Wizard", "Maelin's Robe of Lore"},
	}
	inferredItems["Taelosian Geomancy Stone Jelki"] = []InferredItem{
		{"Bard", "Song: Echo of the Trusik"},
		{"Beastlord", "Spell: Trushar's Mending"},
		{"Berserker", "Tome of Battle Cry of the Mastruq"},
		{"Cleric", "Spell: Holy Elixir"},
		{"Druid", "Spell: Sylvan Fire"},
		{"Enchanter", "Spell: Bliss of the Nihil"},
		{"Magician", "Spell: Elemental Siphon"},
		{"Monk", "Tome of Phantom Shadow"},
		{"Necromancer", "Spell: Night Stalker"},
		{"Paladin", "Spell: Wave of Trushar"},
		{"Ranger", "Spell: Sylvan Burn"},
		{"Rogue", "Tome of Kyv Strike"},
		{"Shadow Knight", "Spell: Black Shroud"},
		{"Shaman", "Spell: Breath of Trushar"},
		{"Warrior", "Tome of Bellow of the Mastruq"},
		{"Wizard", "Spell: White Fire"},
	}
	inferredItems["Taelosian Geomancy Stone Eril"] = []InferredItem{
		{"Bard", "Song: Dark Echo"},
		{"Beastlord", "Spell: Trushar's Frost"},
		{"Cleric", "Spell: Order"},
		{"Druid", "Spell: Sylvan Embers"},
		{"Enchanter", "Spell: Madness of Ikkibi"},
		{"Magician", "Spell: Monster Summoning IV"},
		{"Necromancer", "Spell: Night's Beckon"},
		{"Paladin", "Spell: Holy Order"},
		{"Ranger", "Spell: Sylvan Call"},
		{"Shadow Knight", "Spell: Miasmic Spear"},
		{"Shaman", "Spell: Daluda's Mending"},
		{"Wizard", "Spell: Telaka"},
	}
	inferredItems["Taelosian Geomancy Stone Yiktu"] = []InferredItem{
		{"Bard", "Song: War March of the Mastruq"},
		{"Beastlord", "Spell: Turepta Blood"},
		{"Cleric", "Spell: Holy Light"},
		{"Druid", "Spell: Sylvan Infusion"},
		{"Enchanter", "Spell: Dreary Deeds"},
		{"Magician", "Spell: Rock of Taelosia"},
		{"Necromancer", "Spell: Night Fire"},
		{"Paladin", "Spell: Light of Order"},
		{"Ranger", "Spell: Sylvan Light"},
		{"Shadow Knight", "Spell: Mental Horror"},
		{"Shaman", "Spell: Balance of the Nihil"},
		{"Wizard", "Spell: Black Ice"},
	}
	inferredItems["Chaos Runes"] = []InferredItem{
		{"Bard", "Spell: Ancient: Chaos Chant"},
		{"Beastlord", "Spell: Ancient: Frozen Chaos"},
		{"Berserker", "Tome of Ancient: Cry of Chaos"},
		{"Cleric", "Spell: Ancient: Chaos Censure"},
		{"Druid", "Spell: Ancient: Chaos Frost"},
		{"Enchanter", "Spell: Ancient: Chaos Madness"},
		{"Magician", "Spell: Ancient: Chaos Vortex"},
		{"Monk", "Tome of Ancient: Phantom Chaos"},
		{"Necromancer", "Spell: Ancient: Seduction of Chaos"},
		{"Paladin", "Spell: Ancient: Force of Chaos"},
		{"Ranger", "Spell: Ancient: Burning Chaos"},
		{"Rogue", "Tome of Ancient: Chaos Strike"},
		{"Shadow Knight", "Spell: Ancient: Bite of Chaos"},
		{"Shaman", "Spell: Ancient: Chaotic Pain"},
		{"Warrior", "Tome of Ancient: Chaos Cry"},
		{"Wizard", "Spell: Ancient: Strike of Chaos"},
	}
	// Muramite
	inferredItems["Muramite Helm Armor"] = []InferredItem{
		{"Bard", "Luvwen's Helm of Melody"},
		{"Beastlord", "Kizash's Savage Heart Cap"},
		{"Berserker", "Harlad's Helm of Fury"},
		{"Cleric", "Dakkamor's Helm of the Divine"},
		{"Druid", "Gaelin's Woodland Cap"},
		{"Enchanter", "Lelyen's Circlet of Entrancement"},
		{"Magician", "Jennu's Circlet of Creation"},
		{"Monk", "Pressl's Cap of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Circlet"},
		{"Paladin", "Trimdet's Helm of Chivalry"},
		{"Ranger", "Nadien's Helm of the Archer"},
		{"Rogue", "Nodnol's Helm of the Scoundrel"},
		{"Shadow Knight", "Rayin's Helm of Abhorrence"},
		{"Shaman", "Kanleku's Helm of Spirits"},
		{"Warrior", "Vadd's Helm of Elite Combat"},
		{"Wizard", "Nunkin's Circlet of Pure Elements"},
	}
	inferredItems["Muramite Sleeve Armor"] = []InferredItem{
		{"Bard", "Luvwen's Vambraces of Melody"},
		{"Beastlord", "Kizash's Savage Heart Sleeves"},
		{"Berserker", "Harlad's Vambraces of Fury"},
		{"Cleric", "Dakkamor's Vambraces of the Divine"},
		{"Druid", "Gaelin's Woodland Sleeves"},
		{"Enchanter", "Lelyen's Sleeves of Entrancement"},
		{"Magician", "Jennu's Sleeves of Creation"},
		{"Monk", "Pressl's Sleeves of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Sleeves"},
		{"Paladin", "Trimdet's Vambraces of Chivalry"},
		{"Ranger", "Nadien's Vambraces of the Archer"},
		{"Rogue", "Nodnol's Vambraces of the Scoundrel"},
		{"Shadow Knight", "Rayin's Vambraces of Abhorrence"},
		{"Shaman", "Kanleku's Vambraces of Spirits"},
		{"Warrior", "Vadd's Vambraces of Elite Combat"},
		{"Wizard", "Nunkin's Sleeves of Pure Elements"},
	}
	inferredItems["Muramite Bracer Armor"] = []InferredItem{
		{"Bard", "Luvwen's Bracer of Melody"},
		{"Beastlord", "Kizash's Savage Heart Bracer"},
		{"Berserker", "Harlad's Bracer of Fury"},
		{"Cleric", "Dakkamor's Bracer of the Divine"},
		{"Druid", "Gaelin's Woodland Bracer"},
		{"Enchanter", "Lelyen's Bracer of Entrancement"},
		{"Magician", "Jennu's Bracer of Creation"},
		{"Monk", "Pressl's Bracer of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Bracer"},
		{"Paladin", "Trimdet's Bracer of Chivalry"},
		{"Ranger", "Nadien's Bracer of the Archer"},
		{"Rogue", "Nodnol's Bracer of the Scoundrel"},
		{"Shadow Knight", "Rayin's Bracer of Abhorrence"},
		{"Shaman", "Kanleku's Bracer of Spirits"},
		{"Warrior", "Vadd's Bracer of Elite Combat"},
		{"Wizard", "Nunkin's Bracer of Pure Elements"},
	}
	inferredItems["Muramite Glove Armor"] = []InferredItem{
		{"Bard", "Luvwen's Gauntlets of Melody"},
		{"Beastlord", "Kizash's Savage Heart Gloves"},
		{"Berserker", "Harlad's Gauntlets of Fury"},
		{"Cleric", "Dakkamor's Gauntlets of the Divine"},
		{"Druid", "Gaelin's Woodland Gauntlets"},
		{"Enchanter", "Lelyen's Gloves of Entrancement"},
		{"Magician", "Jennu's Gloves of Creation"},
		{"Monk", "Pressl's Gloves of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Gloves"},
		{"Paladin", "Trimdet's Gauntlets of Chivalry"},
		{"Ranger", "Nadien's Gauntlets of the Archer"},
		{"Rogue", "Nodnol's Gauntlets of the Scoundrel"},
		{"Shadow Knight", "Rayin's Gauntlets of Abhorrence"},
		{"Shaman", "Kanleku's Gauntlets of Spirits"},
		{"Warrior", "Vadd's Gauntlets of Elite Combat"},
		{"Wizard", "Nunkin's Gloves of Pure Elements"},
	}
	inferredItems["Muramite Boot Armor"] = []InferredItem{
		{"Bard", "Luvwen's Boots of Melody"},
		{"Beastlord", "Kizash's Savage Heart Sandals"},
		{"Berserker", "Harlad's Boots of Fury"},
		{"Cleric", "Dakkamor's Boots of the Divine"},
		{"Druid", "Gaelin's Woodland Sandals"},
		{"Enchanter", "Lelyen's Sandals of Entrancement"},
		{"Magician", "Jennu's Sandals of Creation"},
		{"Monk", "Pressl's Sandals of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Sandals"},
		{"Paladin", "Trimdet's Boots of Chivalry"},
		{"Ranger", "Nadien's Boots of the Archer"},
		{"Rogue", "Nodnol's Boots of the Scoundrel"},
		{"Shadow Knight", "Rayin's Boots of Abhorrence"},
		{"Shaman", "Kanleku's Boots of Spirits"},
		{"Warrior", "Vadd's Boots of Elite Combat"},
		{"Wizard", "Nunkin's Sandals of Pure Elements"},
	}
	inferredItems["Muramite Greaves Armor"] = []InferredItem{
		{"Bard", "Luvwen's Legplates of Melody"},
		{"Beastlord", "Kizash's Leggings of Savage Heart"},
		{"Berserker", "Harlad's Greaves of Fury"},
		{"Cleric", "Dakkamor's Legplates of the Divine"},
		{"Druid", "Gaelin's Leggings of Woodlands"},
		{"Enchanter", "Lelyen's Pantaloons of Entrancement"},
		{"Magician", "Jennu's Pantaloons of Creation"},
		{"Monk", "Pressl's Leggings of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Pantaloons"},
		{"Paladin", "Trimdet's Legplates of Chivalry"},
		{"Ranger", "Nadien's Greaves of the Archer"},
		{"Rogue", "Nodnol's Greaves of the Scoundrel"},
		{"Shadow Knight", "Rayin's Legplates of Abhorrence"},
		{"Shaman", "Kanleku's Greaves of Spirits"},
		{"Warrior", "Vadd's Legplates of Elite Combat"},
		{"Wizard", "Nunkin's Pantaloons of Pure Elements"},
	}
	inferredItems["Muramite Chest Armor"] = []InferredItem{
		{"Bard", "Luvwen's Chestplate of Melody"},
		{"Beastlord", "Kizash's Tunic Savage Heart"},
		{"Berserker", "Harlad's Chainmail of Fury"},
		{"Cleric", "Dakkamor's Chestplate of the Divine"},
		{"Druid", "Gaelin's Tunic of Woodlands"},
		{"Enchanter", "Lelyen's Robe of Entrancement"},
		{"Magician", "Jennu's Robe of Creation"},
		{"Monk", "Pressl's Robe of Balance"},
		{"Necromancer", "Nolaen's Lifereaper Robe"},
		{"Paladin", "Trimdet's Chestplate of Chivalry"},
		{"Ranger", "Nadien's Chainmail of the Archer"},
		{"Rogue", "Nodnol's Chainmail of the Scoundrel"},
		{"Shadow Knight", "Rayin's Chestplate of Abhorrence"},
		{"Shaman", "Kanleku's Chainmail of Spirits"},
		{"Warrior", "Vadd's Chestplate of Elite Combat"},
		{"Wizard", "Nunkin's Robe of Pure Elements"},
	}
	// OOW Tier 2
	inferredItems["Jayruk's Vest"] = []InferredItem{ // Chest
		{"Bard", "Farseeker's Plate Chestguard of Harmony"},
		{"Beastlord", "Savagesoul Jerkin of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Chestguard of the Vindicator"},
		{"Cleric", "Faithbringer's Breastplate of Conviction"},
		{"Druid", "Everspring Jerkin of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Vest of Coercion"},
		{"Magician", "Glyphwielder's Tunic of the Summoner"},
		{"Monk", "Fiercehand Shroud of the Focused"},
		{"Necromancer", "Blightbringer's Tunic of the Grave"},
		{"Paladin", "Dawnseeker's Chestpiece of the Defender"},
		{"Ranger", "Bladewhisper Chain Vest of Journeys"},
		{"Rogue", "Whispering Tunic of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Chestguard of the Hateful"},
		{"Shaman", "Ritualchanter's Tunic of the Ancestors"},
		{"Warrior", "Gladiator's Plate Chestguard of War"},
		{"Wizard", "Academic's Robe of the Arcanists"},
	}
	inferredItems["Lenarsk's Embossed Leather Pouch"] = []InferredItem{ // Arms
		{"Bard", "Farseeker's Plate Armbands of Harmony"},
		{"Beastlord", "Savagesoul Sleeves of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Sleeves of the Vindicator"},
		{"Cleric", "Faithbringer's Armguards of Conviction"},
		{"Druid", "Everspring Sleeves of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Armguards of Coercion"},
		{"Magician", "Glyphwielder's Sleeves of the Summoner"},
		{"Monk", "Fiercehand Sleeves of the Focused"},
		{"Necromancer", "Blightbringer's Armband of the Grave"},
		{"Paladin", "Dawnseeker's Sleeves of the Defender"},
		{"Ranger", "Bladewhisper Chain Sleeves of Journeys"},
		{"Rogue", "Whispering Armguard of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Armguards of the Hateful"},
		{"Shaman", "Ritualchanter's Armguards of the Ancestors"},
		{"Warrior", "Gladiator's Plate Sleeves of War"},
		{"Wizard", "Academic's Sleeves of the Arcanists"},
	}
	inferredItems["Makyah's Axe"] = []InferredItem{ // Hands
		{"Bard", "Farseeker's Plate Gloves of Harmony"},
		{"Beastlord", "Savagesoul Gloves of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Gloves of the Vindicator"},
		{"Cleric", "Faithbringer's Gloves of Conviction"},
		{"Druid", "Everspring Mitts of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Handguards of Coercion"},
		{"Magician", "Glyphwielder's Gloves of the Summoner"},
		{"Monk", "Fiercehand Gloves of the Focused"},
		{"Necromancer", "Blightbringer's Handguards of the Grave"},
		{"Paladin", "Dawnseeker's Mitts of the Defender"},
		{"Ranger", "Bladewhisper Chain Gloves of Journeys"},
		{"Rogue", "Whispering Gloves of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Gloves of the Hateful"},
		{"Shaman", "Ritualchanter's Mitts of the Ancestors"},
		{"Warrior", "Gladiator's Plate Gloves of War"},
		{"Wizard", "Academic's Gloves of the Arcanists"},
	}
	inferredItems["Muramite Cruelty Medal"] = []InferredItem{ // Feet
		{"Bard", "Farseeker's Plate Boots of Harmony"},
		{"Beastlord", "Savagesoul Sandals of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Boots of the Vindicator"},
		{"Cleric", "Faithbringer's Boots of Conviction"},
		{"Druid", "Everspring Slippers of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Shoes of Coercion"},
		{"Magician", "Glyphwielder's Slippers of the Summoner"},
		{"Monk", "Fiercehand Tabis of the Focused"},
		{"Necromancer", "Blightbringer's Sandals of the Grave"},
		{"Paladin", "Dawnseeker's Boots of the Defender"},
		{"Ranger", "Bladewhisper Chain Boots of Journeys"},
		{"Rogue", "Whispering Boots of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Boots of the Hateful"},
		{"Shaman", "Ritualchanter's Boots of the Ancestors"},
		{"Warrior", "Gladiator's Plate Boots of War"},
		{"Wizard", "Academic's Slippers of the Arcanists"},
	}
	inferredItems["Patorav's Amulet"] = []InferredItem{ // Legs
		{"Bard", "Farseeker's Plate Legguards of Harmony"},
		{"Beastlord", "Savagesoul Legguards of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Leggings of the Vindicator"},
		{"Cleric", "Faithbringer's Leggings of Conviction"},
		{"Druid", "Everspring Pants of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Leggings of Coercion"},
		{"Magician", "Glyphwielder's Leggings of the Summoner"},
		{"Monk", "Fiercehand Leggings of the Focused"},
		{"Necromancer", "Blightbringer's Pants of the Grave"},
		{"Paladin", "Dawnseeker's Leggings of the Defender"},
		{"Ranger", "Bladewhisper Chain Legguards of Journeys"},
		{"Rogue", "Whispering Pants of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Legguards of the Hateful"},
		{"Shaman", "Ritualchanter's Leggings of the Ancestors"},
		{"Warrior", "Gladiator's Plate Legguards of War"},
		{"Wizard", "Academic's Pants of the Arcanists"},
	}
	inferredItems["Patorav's Walking Stick"] = []InferredItem{ // Head
		{"Bard", "Farseeker's Plate Helm of Harmony"},
		{"Beastlord", "Savagesoul Cap of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Helm of the Vindicator "},
		{"Cleric", "Faithbringer's Cap of Conviction"},
		{"Druid", "Everspring Cap of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Skullcap of Coercion"},
		{"Magician", "Glyphwielder's Hat of the Summoner"},
		{"Monk", "Fiercehand Cap of the Focused"},
		{"Necromancer", "Blightbringer's Cap of the Grave"},
		{"Paladin", "Dawnseeker's Coif of the Defender"},
		{"Ranger", "Bladewhisper Chain Cap of Journeys"},
		{"Rogue", "Whispering Hat of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Helm of the Hateful"},
		{"Shaman", "Ritualchanter's Cap of the Ancestors"},
		{"Warrior", "Gladiator's Plate Helm of War"},
		{"Wizard", "Academic's Cap of the Arcanists"},
	}
	inferredItems["Riftseeker Heart"] = []InferredItem{ // Wrists
		{"Bard", "Farseeker's Plate Wristguard of Harmony"},
		{"Beastlord", "Savagesoul Wristband of the Wilds"},
		{"Berserker", "Wrathbringer's Chain Wristguard of the Vindicator"},
		{"Cleric", "Faithbringer's Wristband of Conviction"},
		{"Druid", "Everspring Wristband of the Tangled Briars"},
		{"Enchanter", "Mindreaver's Bracer of Coercion"},
		{"Magician", "Glyphwielder's Wristband of the Summoner"},
		{"Monk", "Fiercehand Wristband of the Focused"},
		{"Necromancer", "Blightbringer's Bracer of the Grave"},
		{"Paladin", "Dawnseeker's Wristguard of the Defender"},
		{"Ranger", "Bladewhisper Chain Wristband of Journeys"},
		{"Rogue", "Whispering Bracer of Shadows"},
		{"Shadow Knight", "Duskbringer's Plate Wristguard of the Hateful"},
		{"Shaman", "Ritualchanter's Wristband of the Ancestors"},
		{"Warrior", "Gladiator's Plate Bracer of War"},
		{"Wizard", "Academic's Wristband of the Arcanists"},
	}
}
