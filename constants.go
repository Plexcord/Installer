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
	Bg           = color.RGBA{R: 0x31, G: 0x33, B: 0x38, A: 0xFF}
	Surface      = color.RGBA{R: 0x2B, G: 0x2D, B: 0x31, A: 0xFF}
	Field        = color.RGBA{R: 0x1E, G: 0x1F, B: 0x22, A: 0xFF}
	Button       = color.RGBA{R: 0x4E, G: 0x50, B: 0x58, A: 0xFF}
	ButtonHover  = color.RGBA{R: 0x6D, G: 0x6F, B: 0x78, A: 0xFF}
	ButtonActive = color.RGBA{R: 0x5C, G: 0x5E, B: 0x66, A: 0xFF}
	Ink          = color.RGBA{R: 0xF2, G: 0xF3, B: 0xF5, A: 0xFF}
	Muted        = color.RGBA{R: 0xB5, G: 0xBA, B: 0xC1, A: 0xFF}
	Accent       = color.RGBA{R: 0x6A, G: 0x1D, B: 0xE6, A: 0xFF}
	AccentHover  = color.RGBA{R: 0x7B, G: 0x35, B: 0xF0, A: 0xFF}
	AccentActive = color.RGBA{R: 0x58, G: 0x14, B: 0xC4, A: 0xFF}
	AccentText   = color.RGBA{R: 0xB3, G: 0x94, B: 0xF7, A: 0xFF}
	DiscordRed   = color.RGBA{R: 0xDA, G: 0x37, B: 0x3C, A: 0xFF}
	RedActive    = color.RGBA{R: 0xA1, G: 0x28, B: 0x28, A: 0xFF}
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
