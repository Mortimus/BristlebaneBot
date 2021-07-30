package dkp

import everquest "github.com/Mortimus/goEverquest"

type DKPHolder struct {
	everquest.GuildMember
	Main     string
	Name     string
	DKPRank  int
	DKP      int
	Thirty   float64
	Sixty    float64
	Ninety   float64
	Lifetime float64
}
