package main

import (
	"io"

	everquest "github.com/Mortimus/goEverquest"
)

var Handlers []LogHandler

type LogHandler interface {
	Handle(msg *everquest.EqLog, out io.Writer)
	Info(out io.Writer)
	OutputChannel() int
}

type Plugin struct {
	Name    string
	Version string
	Author  string
	Output  int
}

const (
	TESTOUT = iota
	STDOUT
	BIDOUT
	INVESTIGATEOUT
	RAIDOUT
	SPELLOUT
	FLAGOUT
	PARSEOUT
)

type DiscordWriter struct {
	Channel string
}

func (dw *DiscordWriter) Write(p []byte) (n int, err error) {
	_, err = discord.ChannelMessageSend(dw.Channel, string(p[:]))
	n = len(p)
	return n, err
}

var BidWriter DiscordWriter
var InvestigateWriter DiscordWriter
var RaidWriter DiscordWriter
var SpellWriter DiscordWriter
var FlagWriter DiscordWriter
var ParseWriter DiscordWriter

func init() {
	BidWriter.Channel = configuration.Discord.LootChannelID
	InvestigateWriter.Channel = configuration.Discord.InvestigationChannelID
	RaidWriter.Channel = configuration.Discord.RaidDumpChannelID
	SpellWriter.Channel = configuration.Discord.SpellDumpChannelID
	FlagWriter.Channel = configuration.Discord.FlagChannelID
	ParseWriter.Channel = configuration.Discord.ParseChannelID
}

func printPlugins(out io.Writer) {
	for _, handler := range Handlers {
		handler.Info(out)
	}
}
