package loud

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/ahmetb/go-cursor"
	"github.com/gliderlabs/ssh"
	"github.com/mgutz/ansi"
	"github.com/nsf/termbox-go"
	"golang.org/x/crypto/ssh/terminal"
)

// Screen represents a UI screen.
type Screen interface {
	SetScreenSize(int, int)
	HandleInputKey(termbox.Event)
	Render()
	Reset()
}

type ScreenStatus int

const (
	SHOW_LOCATION ScreenStatus = iota
	SELECT_SELL_ITEM
	SELECT_BUY_ITEM
	SELECT_HUNT_ITEM
	SELECT_UPGRADE_ITEM
)

type GameScreen struct {
	world          World
	user           User
	screenSize     ssh.Window
	refreshed      bool
	scrStatus      ScreenStatus
	colorCodeCache map[string](func(string) string)
}

const allowMouseInputAndHideCursor string = "\x1b[?1003h\x1b[?25l"
const resetScreen string = "\x1bc"
const ellipsis = "…"
const hpon = "◆"
const hpoff = "◇"
const bgcolor = 232

func truncateRight(message string, width int) string {
	if utf8.RuneCountInString(message) < width {
		fmtString := fmt.Sprintf("%%-%vs", width)

		return fmt.Sprintf(fmtString, message)
	}
	return string([]rune(message)[0:width-1]) + ellipsis
}

func truncateLeft(message string, width int) string {
	if utf8.RuneCountInString(message) < width {
		fmtString := fmt.Sprintf("%%-%vs", width)

		return fmt.Sprintf(fmtString, message)
	}
	strLen := utf8.RuneCountInString(message)
	return ellipsis + string([]rune(message)[strLen-width:strLen-1])
}

func justifyRight(message string, width int) string {
	if utf8.RuneCountInString(message) < width {
		fmtString := fmt.Sprintf("%%%vs", width)

		return fmt.Sprintf(fmtString, message)
	}
	strLen := utf8.RuneCountInString(message)
	return ellipsis + string([]rune(message)[strLen-width:strLen-1])
}

func centerText(message, pad string, width int) string {
	if utf8.RuneCountInString(message) > width {
		return truncateRight(message, width)
	}
	leftover := width - utf8.RuneCountInString(message)
	left := leftover / 2
	right := leftover - left

	if pad == "" {
		pad = " "
	}

	leftString := ""
	for utf8.RuneCountInString(leftString) <= left && utf8.RuneCountInString(leftString) <= right {
		leftString += pad
	}

	return fmt.Sprintf("%s%s%s", string([]rune(leftString)[0:left]), message, string([]rune(leftString)[0:right]))
}

func (screen *GameScreen) SetScreenSize(Width, Height int) {
	screen.screenSize = ssh.Window{
		Width:  Width,
		Height: Height,
	}
	screen.refreshed = false
}

func (screen *GameScreen) colorFunc(color string) func(string) string {
	_, ok := screen.colorCodeCache[color]

	if !ok {
		screen.colorCodeCache[color] = ansi.ColorFunc(color)
	}

	return screen.colorCodeCache[color]
}

func (screen *GameScreen) drawBox(x, y, width, height int) {
	color := ansi.ColorCode(fmt.Sprintf("255:%v", bgcolor))

	for i := 1; i < width; i++ {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s─", cursor.MoveTo(y, x+i), color))
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s─", cursor.MoveTo(y+height, x+i), color))
	}

	for i := 1; i < height; i++ {
		midString := fmt.Sprintf("%%s%%s│%%%vs│", (width - 1))
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s│", cursor.MoveTo(y+i, x), color))
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s│", cursor.MoveTo(y+i, x+width), color))
		io.WriteString(os.Stdout, fmt.Sprintf(midString, cursor.MoveTo(y+i, x), color, " "))
	}

	io.WriteString(os.Stdout, fmt.Sprintf("%s%s╭", cursor.MoveTo(y, x), color))
	io.WriteString(os.Stdout, fmt.Sprintf("%s%s╰", cursor.MoveTo(y+height, x), color))
	io.WriteString(os.Stdout, fmt.Sprintf("%s%s╮", cursor.MoveTo(y, x+width), color))
	io.WriteString(os.Stdout, fmt.Sprintf("%s%s╯", cursor.MoveTo(y+height, x+width), color))
}

func (screen *GameScreen) drawFill(x, y, width, height int) {
	color := ansi.ColorCode(fmt.Sprintf("0:%v", bgcolor))

	midString := fmt.Sprintf("%%s%%s%%%vs", (width))
	for i := 0; i <= height; i++ {
		io.WriteString(os.Stdout, fmt.Sprintf(midString, cursor.MoveTo(y+i, x), color, " "))
	}
}

func (screen *GameScreen) drawProgressMeter(min, max, fgcolor, bgcolor, width uint64) string {
	var blink bool
	if min > max {
		min = max
		blink = true
	}
	proportion := float64(float64(min) / float64(max))
	if math.IsNaN(proportion) {
		proportion = 0.0
	} else if proportion < 0.05 {
		blink = true
	}
	onWidth := uint64(float64(width) * proportion)
	offWidth := uint64(float64(width) * (1.0 - proportion))

	onColor := screen.colorFunc(fmt.Sprintf("%v:%v", fgcolor, bgcolor))
	offColor := onColor

	if blink {
		onColor = screen.colorFunc(fmt.Sprintf("%v+B:%v", fgcolor, bgcolor))
	}

	if (onWidth + offWidth) > width {
		onWidth = width
		offWidth = 0
	} else if (onWidth + offWidth) < width {
		onWidth += width - (onWidth + offWidth)
	}

	on := ""
	off := ""

	for i := 0; i < int(onWidth); i++ {
		on += hpon
	}

	for i := 0; i < int(offWidth); i++ {
		off += hpoff
	}

	return onColor(on) + offColor(off)
}

func (screen *GameScreen) drawVerticalLine(x, y, height int) {
	color := ansi.ColorCode(fmt.Sprintf("255:%v", bgcolor))
	for i := 1; i < height; i++ {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s│", cursor.MoveTo(y+i, x), color))
	}

	io.WriteString(os.Stdout, fmt.Sprintf("%s%s┬", cursor.MoveTo(y, x), color))
	io.WriteString(os.Stdout, fmt.Sprintf("%s%s┴", cursor.MoveTo(y+height, x), color))
}

func (screen *GameScreen) drawHorizontalLine(x, y, width int) {
	color := ansi.ColorCode(fmt.Sprintf("255:%v", bgcolor))
	for i := 1; i < width; i++ {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s─", cursor.MoveTo(y, x+i), color))
	}

	io.WriteString(os.Stdout, fmt.Sprintf("%s%s├", cursor.MoveTo(y, x), color))
	io.WriteString(os.Stdout, fmt.Sprintf("%s%s┤", cursor.MoveTo(y, x+width), color))
}

func (screen *GameScreen) redrawBorders() {
	io.WriteString(os.Stdout, ansi.ColorCode(fmt.Sprintf("255:%v", bgcolor)))
	screen.drawBox(1, 1, screen.screenSize.Width-1, screen.screenSize.Height-1)
	screen.drawVerticalLine(screen.screenSize.Width/2-2, 1, screen.screenSize.Height)

	y := screen.screenSize.Height
	if y < 20 {
		y = 5
	} else {
		y = (y / 2) - 2
	}
	screen.drawHorizontalLine(1, y+2, screen.screenSize.Width/2-3)
	screen.drawHorizontalLine(1, screen.screenSize.Height-2, screen.screenSize.Width/2-3)
}

func (screen *GameScreen) renderUserCommands() {
	infoLines := []string{}
	if screen.scrStatus == SHOW_LOCATION {
		cmdMap := map[UserLocation]string{
			HOME:   "F)orest\nS)hop",
			FOREST: "Hu)nt\nH)ome\nS)hop",
			SHOP:   "B)uy Items\nSe)ll Items\nUp)grade Items\nH)ome\nF)orest",
		}

		cmdString := cmdMap[screen.user.GetLocation()]
		infoLines = strings.Split(cmdString, "\n")
	} else if screen.scrStatus == SELECT_BUY_ITEM {
		shopItems := []Item{
			Item{
				ID:    "001",
				Name:  "Wooden sword",
				Level: 1,
			},
			Item{
				ID:    "002",
				Name:  "Copper sword",
				Level: 1,
			},
		}
		for idx, item := range shopItems {
			infoLines = append(infoLines, fmt.Sprintf("%d) %s Lv%d", idx+1, item.Name, item.Level))
		}
		infoLines = append(infoLines, "C)ancel")
	} else if screen.scrStatus == SELECT_SELL_ITEM {
		userItems := screen.user.InventoryItems()
		for idx, item := range userItems {
			infoLines = append(infoLines, fmt.Sprintf("%d) %s Lv%d", idx+1, item.Name, item.Level))
		}
		infoLines = append(infoLines, "C)ancel")
	} else if screen.scrStatus == SELECT_HUNT_ITEM {
		userWeaponItems := screen.user.InventoryItems()
		infoLines = append(infoLines, "N)o Item")
		for idx, item := range userWeaponItems {
			infoLines = append(infoLines, fmt.Sprintf("%d) %s Lv%d", idx+1, item.Name, item.Level))
		}
		infoLines = append(infoLines, "C)ancel")
	} else if screen.scrStatus == SELECT_UPGRADE_ITEM {
		userUpgradeItems := screen.user.InventoryItems()
		for idx, item := range userUpgradeItems {
			infoLines = append(infoLines, fmt.Sprintf("%d) %s Lv%d", idx+1, item.Name, item.Level))
		}
		infoLines = append(infoLines, "C)ancel")
	}

	// box start point (x, y)
	x := 2
	y := screen.screenSize.Height/2 + 1

	bgcolor := uint64(bgcolor)
	fmtFunc := screen.colorFunc(fmt.Sprintf("255:%v", bgcolor))
	for index, line := range infoLines {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s",
			cursor.MoveTo(y+index, x), fmtFunc(line)))
		if index+2 > int(screen.screenSize.Height) {
			break
		}
	}
}

func (screen *GameScreen) renderUserSituation() {
	infoLines := []string{}
	desc := ""
	if screen.scrStatus == SHOW_LOCATION {
		locationDescMap := map[UserLocation]string{
			HOME:   "You are now at home.\nIf you like to hunt something go to forest.\nAnd if you like to buy/sell something go to shop",
			FOREST: "You are now at forest.\nYou can hunt here or go back to home.",
			SHOP:   "You are now at a shop.\nIf you want you can buy or sell items here.",
		}
		desc = locationDescMap[screen.user.GetLocation()]
	} else if screen.scrStatus == SELECT_BUY_ITEM {
		desc = "You are gonna buy an item.\nPlease select an item to buy."
	} else if screen.scrStatus == SELECT_SELL_ITEM {
		desc = "You are gonna sell an item.\nPlease select an item to sell."
	} else if screen.scrStatus == SELECT_HUNT_ITEM {
		desc = "You are preparing for hunt.\nPlease select an item to carry."
	} else if screen.scrStatus == SELECT_UPGRADE_ITEM {
		desc = "You are now upgrading an item.\nPlease select an item to upgrade."
	}
	basicLines := strings.Split(desc, "\n")

	for _, line := range basicLines {
		infoLines = append(infoLines, ChunkString(line, screen.screenSize.Width/2-4)...)
	}

	// box start point (x, y)
	x := 2
	y := 2

	bgcolor := uint64(bgcolor)
	fmtFunc := screen.colorFunc(fmt.Sprintf("255:%v", bgcolor))
	for index, line := range infoLines {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s", cursor.MoveTo(y+index, x), fmtFunc(line)))
		if index+2 > int(screen.screenSize.Height) {
			break
		}
	}
}

func (screen *GameScreen) renderCharacterSheet() {
	var HP uint64 = 10
	var MaxHP uint64 = 10
	bgcolor := uint64(bgcolor)
	warning := ""
	if float32(HP) < float32(MaxHP)*.25 {
		bgcolor = 124
		warning = " (Health low) "
	} else if float32(HP) < float32(MaxHP)*.1 {
		bgcolor = 160
		warning = " (Health CRITICAL) "
	}

	x := screen.screenSize.Width/2 - 1
	width := (screen.screenSize.Width - x)
	fmtFunc := screen.colorFunc(fmt.Sprintf("255:%v", bgcolor))

	infoLines := []string{
		centerText(fmt.Sprintf("%v", screen.user.GetUserName()), " ", width),
		centerText(warning, "─", width),
		screen.drawProgressMeter(1, 1, 208, bgcolor, 1) + fmtFunc(truncateRight(fmt.Sprintf(" Gold: %v", screen.user.GetGold()), width-1)),
		screen.drawProgressMeter(HP, MaxHP, 196, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" HP: %v/%v", HP, MaxHP), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 225, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" XP: %v/%v", HP, 10), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 208, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" AP: %v/%v", HP, MaxHP), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 117, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" RP: %v/%v", HP, MaxHP), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 76, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" MP: %v/%v", HP, MaxHP), width-10)),
	}

	infoLines = append(infoLines, centerText(" Inventory Items ", "─", width))
	items := screen.user.InventoryItems()
	for _, item := range items {
		infoLines = append(infoLines, truncateRight(fmt.Sprintf("%s Lv%d", item.Name, item.Level), width))
	}
	infoLines = append(infoLines, centerText(" ❦ ", "─", width))

	for index, line := range infoLines {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s", cursor.MoveTo(2+index, x), fmtFunc(line)))
		if index+2 > int(screen.screenSize.Height) {
			break
		}
	}

	lastLine := len(infoLines) + 1
	screen.drawFill(x, lastLine+1, width, screen.screenSize.Height-(lastLine+2))
}

func (screen *GameScreen) HandleInputKey(input termbox.Event) {
	Key := string(input.Ch)
	switch Key {
	case "H": // HOME
		fallthrough
	case "h":
		screen.user.SetLocation(HOME)
		screen.refreshed = false
	case "F": // FOREST
		fallthrough
	case "f":
		screen.user.SetLocation(FOREST)
		screen.refreshed = false
	case "S": // SHOP
		fallthrough
	case "s":
		screen.user.SetLocation(SHOP)
		screen.refreshed = false
	case "C": // CANCEL
		fallthrough
	case "c":
		screen.scrStatus = SHOW_LOCATION
		screen.refreshed = false
	case "U": // HUNT
		fallthrough
	case "u":
		screen.scrStatus = SELECT_HUNT_ITEM
		screen.refreshed = false
	case "B": // BUY
		fallthrough
	case "b": // BUY
		screen.scrStatus = SELECT_BUY_ITEM
		screen.refreshed = false
	case "E": // SELL
		fallthrough
	case "e":
		screen.scrStatus = SELECT_SELL_ITEM
		screen.refreshed = false
	case "P": // UPGRADE ITEM
		fallthrough
	case "N": // No Item
		fallthrough
	case "n":
		screen.scrStatus = SHOW_LOCATION
		screen.refreshed = false
	case "p":
		screen.scrStatus = SELECT_UPGRADE_ITEM
		screen.refreshed = false
	case "1": // SELECT 1st item
		screen.scrStatus = SHOW_LOCATION
		screen.refreshed = false
	case "2": // SELECT 2nd item
		screen.scrStatus = SHOW_LOCATION
		screen.refreshed = false
	case "3": // SELECT 3rd item
		screen.scrStatus = SHOW_LOCATION
		screen.refreshed = false
	case "4": // SELECT 4th item
		screen.scrStatus = SHOW_LOCATION
		screen.refreshed = false
	}
	screen.Render()
}

func (screen *GameScreen) Render() {
	var HP uint64 = 10

	// screen.user.Reload()

	if screen.screenSize.Height < 20 || screen.screenSize.Width < 60 {
		clear := cursor.ClearEntireScreen()
		move := cursor.MoveTo(1, 1)
		io.WriteString(os.Stdout,
			fmt.Sprintf("%s%sScreen is too small. Make your terminal larger. (60x20 minimum)", clear, move))
		return
	} else if HP == 0 {
		clear := cursor.ClearEntireScreen()
		dead := "You died. Respawning..."
		move := cursor.MoveTo(screen.screenSize.Height/2, screen.screenSize.Width/2-utf8.RuneCountInString(dead)/2)
		io.WriteString(os.Stdout, clear+move+dead)
		screen.refreshed = false
		return
	}

	if !screen.refreshed {
		clear := cursor.ClearEntireScreen() + allowMouseInputAndHideCursor
		io.WriteString(os.Stdout, clear)
		screen.redrawBorders()
		screen.refreshed = true
	}

	screen.renderUserCommands()
	screen.renderUserSituation()
	screen.renderCharacterSheet()
}

func (screen *GameScreen) Reset() {
	io.WriteString(os.Stdout, fmt.Sprintf("%s👋\n", resetScreen))
}

// NewScreen manages the window rendering for game
func NewScreen(world World, user User) Screen {
	width, height, _ := terminal.GetSize(0)

	window := ssh.Window{
		Width:  width,
		Height: height,
	}

	screen := GameScreen{
		world:          world,
		user:           user,
		screenSize:     window,
		colorCodeCache: make(map[string](func(string) string))}

	return &screen
}
