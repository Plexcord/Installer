/*
 * SPDX-License-Identifier: GPL-3.0
 * Plexcord Installer, a cross platform gui/cli app for installing Plexcord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 * Copyright (c) 2025 MutanPlex
 */

package main

import (
	"image/color"
	"plexcordinstaller/buildinfo"
)

const ReleaseUrl = "https://api.github.com/repos/MutanPlex/Plexcord/releases/latest"
const ReleaseUrlFallback = "https://plexcord.club/releases/plexcord"
const InstallerReleaseUrl = "https://api.github.com/repos/Plexcord/Installer/releases/latest"
const InstallerReleaseUrlFallback = "https://plexcord.club/releases/installer"

var UserAgent = "PlexcordInstaller/" + buildinfo.InstallerGitHash + " (https://github.com/Plexcord/Installer)"

var (
	DiscordGreen  = color.RGBA{R: 0x2D, G: 0x7C, B: 0x46, A: 0xFF}
	DiscordRed    = color.RGBA{R: 0xEC, G: 0x41, B: 0x44, A: 0xFF}
	DiscordBlue   = color.RGBA{R: 0x58, G: 0x65, B: 0xF2, A: 0xFF}
	DiscordYellow = color.RGBA{R: 0xfe, G: 0xe7, B: 0x5c, A: 0xff}
)

var LinuxDiscordNames = []string{
	"Discord",
	"DiscordPTB",
	"DiscordCanary",
	"DiscordDevelopment",
	"discord",
	"discordptb",
	"discordcanary",
	"discorddevelopment",
	"discord-ptb",
	"discord-canary",
	"discord-development",
	// Flatpak
	"com.discordapp.Discord",
	"com.discordapp.DiscordPTB",
	"com.discordapp.DiscordCanary",
	"com.discordapp.DiscordDevelopment",
}
