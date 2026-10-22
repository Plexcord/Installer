//go:build !cli

/*
 * SPDX-License-Identifier: GPL-3.0
 * Plexcord Installer, a cross platform gui/cli app for installing Plexcord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 * Copyright (c) 2025 MutanPlex
 */

package main

import (
	"bytes"
	_ "embed"
	"errors"
	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"image"
	"image/color"
	"plexcordinstaller/buildinfo"
	// png decoder for icon
	_ "image/png"
	"os"
	path "path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	discords        []any
	radioIdx        int
	customChoiceIdx int

	customDir              string
	autoCompleteDir        string
	autoCompleteFile       string
	autoCompleteCandidates []string
	autoCompleteIdx        int
	lastAutoComplete       string
	didAutoComplete        bool

	modalId      = 0
	modalTitle   = "Oh No :("
	modalMessage = "You should never see this"

	acceptedOpenAsar   bool
	showedUpdatePrompt bool

	win  *g.MasterWindow
	logo image.Image
)

//go:embed winres/icon.png
var iconBytes []byte

func init() {
	LogLevel = LevelDebug
}

func main() {
	InitGithubDownloader()
	discords = FindDiscords()

	customChoiceIdx = len(discords)

	go func() {
		<-GithubDoneChan
		g.Update()
	}()

	go func() {
		<-SelfUpdateCheckDoneChan
		g.Update()
	}()

	win = g.NewMasterWindow("Plexcord Installer", 1200, 800, 0)
	icon, _, err := image.Decode(bytes.NewReader(iconBytes))
	if err != nil {
		Log.Warn("Failed to load application icon", err)
		Log.Debug(iconBytes, len(iconBytes))
	} else {
		win.SetIcon([]image.Image{icon})
		logo = icon
	}
	
	win.SetBgColor(Bg)
	
	win.Run(loop)
}

type CondWidget struct {
	predicate  bool
	ifWidget   func() g.Widget
	elseWidget func() g.Widget
}

func (w *CondWidget) Build() {
	if w.predicate {
		w.ifWidget().Build()
	} else if w.elseWidget != nil {
		w.elseWidget().Build()
	}
}

func getChosenInstall() *DiscordInstall {
	var choice *DiscordInstall
	if radioIdx == customChoiceIdx {
		choice = ParseDiscord(customDir, "")
		if choice == nil {
			g.OpenPopup("#invalid-custom-location")
		}
	} else {
		choice = discords[radioIdx].(*DiscordInstall)
	}
	return choice
}

func InstallLatestBuilds() (err error) {
	if IsDevInstall {
		return
	}

	err = installLatestBuilds()
	if err != nil {
		ShowModal("Uh Oh!", "Failed to install the latest Plexcord builds from GitHub:\n"+err.Error())
	}
	return
}

func handlePatch() {
	choice := getChosenInstall()
	if choice != nil {
		choice.Patch()
	}
}

func handleUnpatch() {
	choice := getChosenInstall()
	if choice != nil {
		choice.Unpatch()
	}
}

func handleOpenAsar() {
	if acceptedOpenAsar || getChosenInstall().IsOpenAsar() {
		handleOpenAsarConfirmed()
		return
	}

	g.OpenPopup("#openasar-confirm")
}

func handleOpenAsarConfirmed() {
	choice := getChosenInstall()
	if choice != nil {
		if choice.IsOpenAsar() {
			if err := choice.UninstallOpenAsar(); err != nil {
				handleErr(choice, err, "uninstall OpenAsar from")
			} else {
				g.OpenPopup("#openasar-unpatched")
				g.Update()
			}
		} else {
			if err := choice.InstallOpenAsar(); err != nil {
				handleErr(choice, err, "install OpenAsar on")
			} else {
				g.OpenPopup("#openasar-patched")
				g.Update()
			}
		}
	}
}

func handleErr(di *DiscordInstall, err error, action string) {
	if errors.Is(err, os.ErrPermission) {
		switch runtime.GOOS {
		case "windows":
			err = errors.New("Permission denied. Make sure your Discord is fully closed (from the tray)!")
		case "darwin":
			// FIXME: This text is not selectable which is a bit mehhh
			command := "sudo chown -R \"${USER}:wheel\" " + di.path
			err = errors.New("Permission denied. Please grant the installer Full Disk Access in the system settings (privacy & security page).\n\nIf that also doesn't work, try running the following command in your terminal:\n" + command)
		default:
			err = errors.New("Permission denied. Maybe try running me as Administrator/Root?")
		}
	}

	ShowModal("Failed to "+action+" this Install", err.Error())
}

func HandleScuffedInstall() {
	g.OpenPopup("#scuffed-install")
}

func (di *DiscordInstall) Patch() {
	if CheckScuffedInstall() {
		return
	}
	if err := di.patch(); err != nil {
		handleErr(di, err, "patch")
	} else {
		g.OpenPopup("#patched")
	}
}

func (di *DiscordInstall) Unpatch() {
	if err := di.unpatch(); err != nil {
		handleErr(di, err, "unpatch")
	} else {
		g.OpenPopup("#unpatched")
	}
}

func onCustomInputChanged() {
	p := customDir
	if len(p) != 0 {
		// Select the custom option for people
		radioIdx = customChoiceIdx
	}

	dir := path.Dir(p)

	isNewDir := strings.HasSuffix(p, "/")
	wentUpADir := !isNewDir && dir != autoCompleteDir

	if isNewDir || wentUpADir {
		autoCompleteDir = dir
		// reset all the funnies
		autoCompleteIdx = 0
		lastAutoComplete = ""
		autoCompleteFile = ""
		autoCompleteCandidates = nil

		// Generate autocomplete items
		files, err := os.ReadDir(dir)
		if err == nil {
			for _, file := range files {
				autoCompleteCandidates = append(autoCompleteCandidates, file.Name())
			}
		}
	} else if !didAutoComplete {
		// reset auto complete and update our file
		autoCompleteFile = path.Base(p)
		lastAutoComplete = ""
	}

	if wentUpADir {
		autoCompleteFile = path.Base(p)
	}

	didAutoComplete = false
}

// go can you give me []any?
// to pass to giu RangeBuilder?
// yeeeeees
// actually returns []string like a boss
func makeAutoComplete() []any {
	input := strings.ToLower(autoCompleteFile)

	var candidates []any
	for _, e := range autoCompleteCandidates {
		file := strings.ToLower(e)
		if autoCompleteFile == "" || strings.HasPrefix(file, input) {
			candidates = append(candidates, e)
		}
	}
	return candidates
}

func makeRadioOnChange(i int) func() {
	return func() {
		radioIdx = i
	}
}

func Tooltip(label string) g.Widget {
	return g.Style().
		SetStyle(g.StyleVarWindowPadding, 10, 8).
		SetStyleFloat(g.StyleVarWindowRounding, 8).
		To(
			g.Tooltip(label),
		)
}

func InfoModal(id, title, description string) g.Widget {
	return RawInfoModal(id, title, description, false)
}

func RawInfoModal(id, title, description string, isOpenAsar bool) g.Widget {
	isDynamic := strings.HasPrefix(id, "#modal") && !strings.Contains(description, "\n")
	return g.Style().
		SetStyle(g.StyleVarWindowPadding, 30, 30).
		SetStyleFloat(g.StyleVarWindowRounding, 12).
		SetColor(g.StyleColorPopupBg, Surface).
		To(
			g.PopupModal(id).
				Flags(g.WindowFlagsNoTitleBar | Ternary(isDynamic, g.WindowFlagsAlwaysAutoResize, 0)).
				Layout(
					g.Align(g.AlignCenter).To(
						g.Style().SetFontSize(24).To(
							g.Label(title),
						),
						g.Style().SetFontSize(18).To(
							g.Label(description).Wrapped(isDynamic),
						),
						&CondWidget{id == "#scuffed-install", func() g.Widget {
							return g.Column(
								g.Dummy(0, 10),
								g.Button("Take me there!").OnClick(func() {
									// this issue only exists on windows so using Windows specific path is oki
									username := os.Getenv("USERNAME")
									programData := os.Getenv("PROGRAMDATA")
									g.OpenURL("file://" + path.Join(programData, username))
								}).Size(200, 30),
							)
						}, nil},
						g.Dummy(0, 20),
						&CondWidget{isOpenAsar,
							func() g.Widget {
								return g.Row(
									g.Button("Accept").
										OnClick(func() {
											acceptedOpenAsar = true
											g.CloseCurrentPopup()
										}).
										Size(100, 30),
									g.Button("Cancel").
										OnClick(func() {
											g.CloseCurrentPopup()
										}).
										Size(100, 30),
								)
							},
							func() g.Widget {
								return g.Button("Ok").
									OnClick(func() {
										g.CloseCurrentPopup()
									}).
									Size(100, 30)
							},
						},
					),
				),
		)
}

func UpdateModal() g.Widget {
	return g.Style().
		SetStyle(g.StyleVarWindowPadding, 30, 30).
		SetStyleFloat(g.StyleVarWindowRounding, 12).
		SetColor(g.StyleColorPopupBg, Surface).
		To(
			g.PopupModal("#update-prompt").
				Flags(g.WindowFlagsNoTitleBar | g.WindowFlagsAlwaysAutoResize).
				Layout(
					g.Align(g.AlignCenter).To(
						g.Style().SetFontSize(24).To(
							g.Label("Your Installer is outdated!"),
						),
						g.Style().SetFontSize(18).To(
							g.Label(
								"Would you like to update now?\n\n"+
									"Once you press Update Now, the new installer will automatically be downloaded.\n"+
									"The installer will temporarily seem unresponsive. Just wait!\n"+
									"Once the update is done, the Installer will automatically reopen.\n\n"+
									"On MacOs, Auto updates are not supported, so it will instead open in browser.",
							),
						),
						g.Row(
							g.Button("Update Now").
								OnClick(func() {
									if runtime.GOOS == "darwin" {
										g.CloseCurrentPopup()
										g.OpenURL(GetInstallerDownloadLink())
										return
									}

									err := UpdateSelf()
									g.CloseCurrentPopup()

									if err != nil {
										ShowModal("Failed to update self!", err.Error())
									} else {
										if err = RelaunchSelf(); err != nil {
											ShowModal("Failed to restart self! Please do it manually.", err.Error())
										}
									}
								}).
								Size(100, 30),
							g.Button("Later").
								OnClick(func() {
									g.CloseCurrentPopup()
								}).
								Size(100, 30),
						),
					),
				),
		)
}

func ShowModal(title, desc string) {
	modalTitle = title
	modalMessage = desc
	modalId++
	g.OpenPopup("#modal" + strconv.Itoa(modalId))
}

func renderInstaller() g.Widget {
	candidates := makeAutoComplete()
	wi, _ := win.GetSize()
	w := float32(wi) - 96

	var currentDiscord *DiscordInstall
	if radioIdx != customChoiceIdx {
		currentDiscord = discords[radioIdx].(*DiscordInstall)
	}
	var isOpenAsar = currentDiscord != nil && currentDiscord.IsOpenAsar()

	if CanUpdateSelf() && !showedUpdatePrompt {
		showedUpdatePrompt = true
		g.OpenPopup("#update-prompt")
	}

	narrow := w < 760
	bw := Ternary(narrow, (w-10)/2, (w-30)/4)

	btnInstall := g.Style().
		SetColor(g.StyleColorButton, Accent).
		SetColor(g.StyleColorButtonHovered, AccentHover).
		SetColor(g.StyleColorButtonActive, AccentActive).
		SetDisabled(GithubError != nil).
		To(
			g.Button("Install").
				OnClick(handlePatch).
				Size(bw, 46),
			Tooltip("Patch the selected Discord Install"),
		)
	btnRepair := g.Style().
		SetDisabled(GithubError != nil).
		To(
			g.Button("Reinstall / Repair").
				OnClick(func() {
					if IsDevInstall {
						handlePatch()
					} else {
						err := InstallLatestBuilds()
						if err == nil {
							handlePatch()
						}
					}
				}).
				Size(bw, 46),
			Tooltip("Reinstall & Update Plexcord"),
		)
	btnUninstall := g.Style().
		SetColor(g.StyleColorButtonHovered, DiscordRed).
		SetColor(g.StyleColorButtonActive, RedActive).
		To(
			g.Button("Uninstall").
				OnClick(handleUnpatch).
				Size(bw, 46),
			Tooltip("Unpatch the selected Discord Install"),
		)
	btnOpenAsar := g.Style().To(
		g.Button(Ternary(isOpenAsar, "Uninstall OpenAsar", Ternary(currentDiscord != nil, "Install OpenAsar", "(Un-)Install OpenAsar"))).
			OnClick(handleOpenAsar).
			Size(bw, 46),
		Tooltip("Manage OpenAsar"),
	)

	layout := g.Layout{
		g.Dummy(0, 8),

		g.Style().SetFontSize(17).To(
			renderCard(Surface,
				"Only download Plexcord from GitHub or plexcord.club",
				"Any other site offering Plexcord is not us and may try to steal your Discord account. "+
					"If you got this installer anywhere else, close it, uninstall what you installed, run a malware scan and change your Discord password.",
				g.Dummy(0, 2),
				g.Style().SetStyle(g.StyleVarFramePadding, 12, 6).To(
					g.Row(
						g.Button("Source code on GitHub").OnClick(func() {
							g.OpenURL("https://github.com/Plexcord/Installer")
						}),
						g.Button("plexcord.club").OnClick(func() {
							g.OpenURL("https://plexcord.club")
						}),
					),
				),
			),
		),

		g.Dummy(0, 12),

		g.Style().SetFontSize(24).To(
			g.Label("Select a Discord install"),
		),

		&CondWidget{len(discords) == 0, func() g.Widget {
			s := "No Discord installs found. You first need to install Discord."
			if runtime.GOOS == "linux" {
				s += " snap is not supported."
			}
			return g.Label(s)
		}, nil},

		g.Style().SetFontSize(18).To(
			g.RangeBuilder("Discords", discords, func(i int, v any) g.Widget {
				d := v.(*DiscordInstall)
				//goland:noinspection GoDeprecation
				radio := g.RadioButton(strings.Title(d.branch), radioIdx == i).
					OnChange(makeRadioOnChange(i))
				if !d.isPatched {
					return radio
				}
				return g.Row(
					radio,
					g.Style().SetColor(g.StyleColorText, AccentText).To(
						g.Label("Plexcord installed"),
					),
				)
			}),

			g.RadioButton("Custom Install Location", radioIdx == customChoiceIdx).
				OnChange(makeRadioOnChange(customChoiceIdx)),
		),

		g.Dummy(0, 5),
		g.Style().
			SetStyle(g.StyleVarFramePadding, 16, 16).
			SetFontSize(18).
			To(
				g.InputText(&customDir).Hint("The custom location").
					Size(w).
					Flags(g.InputTextFlagsCallbackCompletion).
					OnChange(onCustomInputChanged).
					// this library has its own autocomplete but it's broken
					Callback(
						func(data imgui.InputTextCallbackData) int32 {
							if len(candidates) == 0 {
								return 0
							}
							// just wrap around
							if autoCompleteIdx >= len(candidates) {
								autoCompleteIdx = 0
							}

							// used by change handler
							didAutoComplete = true

							start := len(customDir)
							// Delete previous auto complete
							if lastAutoComplete != "" {
								start -= len(lastAutoComplete)
								data.DeleteBytes(start, len(lastAutoComplete))
							} else if autoCompleteFile != "" { // delete partial input
								start -= len(autoCompleteFile)
								data.DeleteBytes(start, len(autoCompleteFile))
							}

							// Insert auto complete
							lastAutoComplete = candidates[autoCompleteIdx].(string)
							data.InsertBytes(start, []byte(lastAutoComplete))
							autoCompleteIdx++

							return 0
						},
					),
			),
		g.RangeBuilder("AutoComplete", candidates, func(i int, v any) g.Widget {
			dir := v.(string)
			return g.Label(dir)
		}),

		g.Dummy(0, 20),

		g.Style().SetFontSize(19).To(
			Ternary(narrow, g.Layout{g.Row(btnInstall, btnRepair), g.Row(btnUninstall, btnOpenAsar)},
				g.Layout{g.Row(btnInstall, btnRepair, btnUninstall, btnOpenAsar)}),
		),

		InfoModal("#patched", "Successfully Patched", "If Discord is still open, fully close it first.\n"+
			"Then, start it and verify Plexcord installed successfully by looking for its category in Discord Settings"),
		InfoModal("#unpatched", "Successfully Unpatched", "If Discord is still open, fully close it first. Then start it again, it should be back to stock!"),
		InfoModal("#scuffed-install", "Hold On!", "You have a broken Discord Install.\n"+
			"Sometimes Discord decides to install to the wrong location for some reason!\n"+
			"You need to fix this before patching, otherwise Plexcord will likely not work.\n\n"+
			"Use the below button to jump there and delete any folder called Discord or Squirrel.\n"+
			"If the folder is now empty, feel free to go back a step and delete that folder too.\n"+
			"Then see if Discord still starts. If not, reinstall it"),
		RawInfoModal("#openasar-confirm", "OpenAsar", "OpenAsar is an open-source alternative of Discord desktop's app.asar.\n"+
			"Plexcord is in no way affiliated with OpenAsar.\n"+
			"You're installing OpenAsar at your own risk. If you run into issues with OpenAsar,\n"+
			"no support will be provided, join the OpenAsar Server instead!\n\n"+
			"To install OpenAsar, press Accept and click 'Install OpenAsar' again.", true),
		InfoModal("#openasar-patched", "Successfully Installed OpenAsar", "If Discord is still open, fully close it first. Then start it again and verify OpenAsar installed successfully!"),
		InfoModal("#openasar-unpatched", "Successfully Uninstalled OpenAsar", "If Discord is still open, fully close it first. Then start it again and it should be back to stock!"),
		InfoModal("#invalid-custom-location", "Invalid Location", "The specified location is not a valid Discord install.\nMake sure you select the base folder.\n\nHint: Discord snap is not supported. use flatpak or .deb"),
		InfoModal("#modal"+strconv.Itoa(modalId), modalTitle, modalMessage),

		UpdateModal(),
	}

	return layout
}

var cardHeights = map[string]float32{}

func renderCard(bg color.Color, title, body string, footer ...g.Widget) g.Widget {
	const pad, itemSpacing = 18, 10
	h := cardHeights[title]
	if h == 0 {
		h = 120
	}
	measure := g.Custom(func() {
		if got := imgui.CursorPosY() - itemSpacing + pad; got != cardHeights[title] {
			cardHeights[title] = got
			g.Update()
		}
	})

	return g.Style().
		SetColor(g.StyleColorChildBg, bg).
		SetStyle(g.StyleVarWindowPadding, pad, pad).
		To(
			g.Child().
				Size(g.Auto, h).
				Flags(g.WindowFlagsAlwaysUseWindowPadding|g.WindowFlagsNoScrollbar|g.WindowFlagsNoScrollWithMouse).
				Layout(append(g.Layout{
					g.Style().SetFontSize(20).To(g.Label(title)),
					g.Style().SetColor(g.StyleColorText, Muted).To(g.Label(body).Wrapped(true)),
				}, append(footer, measure)...)...),
		)
}

func loop() {
	g.PushWindowPadding(48, 40)
	g.PushStyleColor(g.StyleColorWindowBg, Bg)

	g.SingleWindow().
		RegisterKeyboardShortcuts(
			g.WindowShortcut{Key: g.KeyUp, Callback: func() {
				if radioIdx > 0 {
					radioIdx--
				}
			}},
			g.WindowShortcut{Key: g.KeyDown, Callback: func() {
				if radioIdx < customChoiceIdx {
					radioIdx++
				}
			}},
		).
		Layout(
			g.Style().
				SetStyleFloat(g.StyleVarFrameRounding, 6).
				SetStyleFloat(g.StyleVarChildRounding, 8).
				SetStyleFloat(g.StyleVarPopupRounding, 10).
				SetStyleFloat(g.StyleVarGrabRounding, 6).
				SetStyleFloat(g.StyleVarScrollbarRounding, 6).
				SetStyleFloat(g.StyleVarFrameBorderSize, 0).
				SetStyleFloat(g.StyleVarChildBorderSize, 0).
				SetStyleFloat(g.StyleVarPopupBorderSize, 0).
				SetStyle(g.StyleVarItemSpacing, 10, 10).
				SetColor(g.StyleColorText, Ink).
				SetColor(g.StyleColorTextDisabled, Muted).
				SetColor(g.StyleColorChildBg, Surface).
				SetColor(g.StyleColorPopupBg, Surface).
				SetColor(g.StyleColorFrameBg, Field).
				SetColor(g.StyleColorFrameBgHovered, Surface).
				SetColor(g.StyleColorFrameBgActive, Field).
				SetColor(g.StyleColorButton, Button).
				SetColor(g.StyleColorButtonHovered, ButtonHover).
				SetColor(g.StyleColorButtonActive, ButtonActive).
				SetColor(g.StyleColorCheckMark, Accent).
				SetColor(g.StyleColorScrollbarBg, Bg).
				SetColor(g.StyleColorScrollbarGrab, Button).
				SetColor(g.StyleColorModalWindowDimBg, color.RGBA{A: 0xA0}).
				To(
					g.Row(
						&CondWidget{logo != nil, func() g.Widget {
							return g.ImageWithRgba(logo).Size(52, 52)
						}, nil},
						g.Column(
							g.Style().SetFontSize(30).To(
								g.Label("Plexcord Installer"),
							),
							g.Style().SetFontSize(17).SetColor(g.StyleColorText, Muted).To(
								g.Label("Official installer for Plexcord, a Discord client mod"),
							),
						),
					),

					&CondWidget{GithubError != nil, func() g.Widget {
						return g.Style().SetFontSize(17).To(
							renderCard(DiscordRed, "Couldn't reach GitHub",
								"Installing is turned off until the installer can check for the latest Plexcord version. "+
									"Check your internet connection and restart the installer.\n"+GithubError.Error(),
								),
						)
					}, nil},

					renderInstaller(),

					// technical details last, for people who want to check them
					g.Dummy(0, 16),
					g.Style().SetFontSize(16).SetColor(g.StyleColorText, Muted).To(
						g.Label(Ternary(IsDevInstall, "Dev install: ", "Plexcord is downloaded from its GitHub releases to ")+PlexcordDirectory).Wrapped(true),
						g.Row(
							g.Style().
								SetColor(g.StyleColorText, Ink).
								SetStyle(g.StyleVarFramePadding, 10, 4).
								To(
									g.Button("Open folder").OnClick(func() {
										g.OpenURL("file://" + path.Dir(PlexcordDirectory))
									}),
								),
							&CondWidget{!IsDevInstall, func() g.Widget {
								return g.Label("Set PLEXCORD_USER_DATA_DIR to change it.")
							}, nil},
						),
						g.Row(
							g.Label("Installer "+buildinfo.InstallerTag+" ("+buildinfo.InstallerGitHash+")"),
							&CondWidget{IsSelfOutdated, func() g.Widget {
								return g.Style().SetColor(g.StyleColorText, DiscordRed).To(
									g.Label("·  update available"),
								)
							}, nil},
							g.Label("·  Installed "+InstalledHash),
							&CondWidget{GithubError == nil, func() g.Widget {
								return g.Label(Ternary(IsDevInstall, "·  Dev mode, not updating", "·  Latest "+LatestHash))
							}, nil},
						),
					),
				),
		)

	g.PopStyleColor()
	g.PopStyle()
}
