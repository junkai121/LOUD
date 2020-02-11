package loud

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ahmetb/go-cursor"
	"github.com/gliderlabs/ssh"
	"github.com/mgutz/ansi"
	"github.com/nsf/termbox-go"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

const allowMouseInputAndHideCursor string = "\x1b[?1003h\x1b[?25l"
const resetScreen string = "\x1bc"
const ellipsis = "…"
const hpon = "◆"
const hpoff = "◇"
const bgcolor = 232

// Screen represents a UI screen.
type Screen interface {
	SetDaemonFetchingFlag(bool)
	SaveGame()
	UpdateBlockHeight(int64)
	SetScreenSize(int, int)
	HandleInputKey(termbox.Event)
	GetScreenStatus() ScreenStatus
	SetScreenStatus(ScreenStatus)
	GetTxFailReason() string
	Render()
	Reset()
}

type ScreenStatus int

type GameScreen struct {
	world                  World
	user                   User
	screenSize             ssh.Window
	activeItem             Item
	lastInput              termbox.Event
	activeLine             int
	activeOrder            Order
	activeItemOrder        ItemOrder
	pylonEnterValue        string
	loudEnterValue         string
	inputText              string
	refreshingDaemonStatus bool
	blockHeight            int64
	txFailReason           string
	txResult               []byte
	refreshed              bool
	scrStatus              ScreenStatus
	colorCodeCache         map[string](func(string) string)
}

const (
	SHOW_LOCATION ScreenStatus = iota
	// in shop
	SELECT_SELL_ITEM
	WAIT_SELL_PROCESS
	RESULT_SELL_FINISH

	SELECT_BUY_ITEM
	WAIT_BUY_PROCESS
	RESULT_BUY_FINISH

	SELECT_UPGRADE_ITEM
	WAIT_UPGRADE_PROCESS
	RESULT_UPGRADE_FINISH
	// in forest
	SELECT_HUNT_ITEM
	WAIT_HUNT_PROCESS
	RESULT_HUNT_FINISH
	WAIT_GET_PYLONS
	RESULT_GET_PYLONS

	// in develop
	WAIT_CREATE_COOKBOOK
	RESULT_CREATE_COOKBOOK
	WAIT_SWITCH_USER
	RESULT_SWITCH_USER

	// in market
	SELECT_MARKET // buy loud or sell loud

	SHOW_LOUD_BUY_ORDERS                   // navigation using arrow and list should be sorted by price
	CREATE_BUY_LOUD_ORDER_ENTER_LOUD_VALUE // enter value after switching enter mode
	CREATE_BUY_LOUD_ORDER_ENTER_PYLON_VALUE
	WAIT_BUY_LOUD_ORDER_CREATION
	RESULT_BUY_LOUD_ORDER_CREATION
	WAIT_FULFILL_BUY_LOUD_ORDER // after done go to show loud buy orders
	RESULT_FULFILL_BUY_LOUD_ORDER

	SHOW_LOUD_SELL_ORDERS
	CREATE_SELL_LOUD_ORDER_ENTER_LOUD_VALUE
	CREATE_SELL_LOUD_ORDER_ENTER_PYLON_VALUE
	WAIT_SELL_LOUD_ORDER_CREATION
	RESULT_SELL_LOUD_ORDER_CREATION
	WAIT_FULFILL_SELL_LOUD_ORDER
	RESULT_FULFILL_SELL_LOUD_ORDER

	SHOW_SWORD_PYLON_ORDERS
	CREATE_SWORD_PYLON_ORDER_SELECT_SWORD
	CREATE_SWORD_PYLON_ORDER_ENTER_PYLON_VALUE
	WAIT_SWORD_PYLON_ORDER_CREATION
	RESULT_SWORD_PYLON_ORDER_CREATION
	WAIT_FULFILL_SWORD_PYLON_ORDER
	RESULT_FULFILL_SWORD_PYLON_ORDER

	SHOW_PYLON_SWORD_ORDERS
	CREATE_PYLON_SWORD_ORDER_SELECT_SWORD
	CREATE_PYLON_SWORD_ORDER_ENTER_PYLON_VALUE
	WAIT_PYLON_SWORD_ORDER_CREATION
	RESULT_PYLON_SWORD_ORDER_CREATION
	WAIT_FULFILL_PYLON_SWORD_ORDER
	RESULT_FULFILL_PYLON_SWORD_ORDER
)

// NewScreen manages the window rendering for game
func NewScreen(world World, user User) Screen {
	width, _ := terminal.Width()
	height, _ := terminal.Height()

	window := ssh.Window{
		Width:  int(width),
		Height: int(height),
	}

	screen := GameScreen{
		world:          world,
		user:           user,
		screenSize:     window,
		colorCodeCache: make(map[string](func(string) string))}

	return &screen
}

func (screen *GameScreen) SwitchUser(newUser User) {
	screen.user = newUser
}

func (screen *GameScreen) GetTxFailReason() string {
	return screen.txFailReason
}

func (screen *GameScreen) GetScreenStatus() ScreenStatus {
	return screen.scrStatus
}

func (screen *GameScreen) SetScreenStatus(newStatus ScreenStatus) {
	screen.scrStatus = newStatus
}

func (screen *GameScreen) Reset() {
	io.WriteString(os.Stdout, fmt.Sprintf("%s👋\n", resetScreen))
}

func (screen *GameScreen) SaveGame() {
	screen.user.Save()
}

func (screen *GameScreen) SetDaemonFetchingFlag(flag bool) {
	screen.refreshingDaemonStatus = flag
}

func (screen *GameScreen) UpdateBlockHeight(blockHeight int64) {
	screen.blockHeight = blockHeight
	screen.refreshed = false
	screen.Render()
}

func (screen *GameScreen) SetInputTextAndRender(text string) {
	screen.inputText = text
	screen.Render()
}

func (screen *GameScreen) pylonIcon() string {
	return screen.drawProgressMeter(1, 1, 117, bgcolor, 1)
}

func (screen *GameScreen) loudIcon() string {
	return screen.drawProgressMeter(1, 1, 208, bgcolor, 1)
}

func (screen *GameScreen) buyLoudDesc(loudValue interface{}, pylonValue interface{}) string {
	var desc = strings.Join([]string{
		"\n",
		screen.pylonIcon(),
		fmt.Sprintf("%v", pylonValue),
		"\n  ↓\n",
		screen.loudIcon(),
		fmt.Sprintf("%v", loudValue),
	}, "")
	return desc
}

func (screen *GameScreen) sellLoudDesc(loudValue interface{}, pylonValue interface{}) string {
	var desc = strings.Join([]string{
		"\n",
		screen.loudIcon(),
		fmt.Sprintf("%v", loudValue),
		"\n  ↓\n",
		screen.pylonIcon(),
		fmt.Sprintf("%v", pylonValue),
	}, "")
	return desc
}

func (screen *GameScreen) tradeTableColorDesc() []string {
	var infoLines = []string{}
	infoLines = append(infoLines, "white     ➝ other's order")
	infoLines = append(infoLines, screen.blueBoldFont()("bluebold")+"  ➝ selected order")
	infoLines = append(infoLines, screen.brownBoldFont()("brownbold")+" ➝ my order + selected")
	infoLines = append(infoLines, screen.brownFont()("brown")+"     ➝ my order")
	infoLines = append(infoLines, "\n")
	return infoLines
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

func (screen *GameScreen) blueBoldFont() func(string) string {
	return screen.colorFunc(fmt.Sprintf("%v+bh:%v", 117, 232))
}

func (screen *GameScreen) brownBoldFont() func(string) string {
	return screen.colorFunc(fmt.Sprintf("%v+bh:%v", 181, 232))
}

func (screen *GameScreen) brownFont() func(string) string {
	return screen.colorFunc(fmt.Sprintf("%v:%v", 181, 232))
}

func (screen *GameScreen) renderOrderTableLine(text1 string, text2 string, text3 string, isActiveLine bool, isDisabledLine bool) string {
	calcText := "│" + centerText(text1, " ", 20) + "│" + centerText(text2, " ", 15) + "│" + centerText(text3, " ", 15) + "│"
	if isActiveLine && isDisabledLine {
		onColor := screen.brownBoldFont()
		return onColor(calcText)
	} else if isActiveLine {
		onColor := screen.blueBoldFont()
		return onColor(calcText)
	} else if isDisabledLine {
		onColor := screen.brownFont()
		return onColor(calcText)
	}
	return calcText
}

func (screen *GameScreen) renderOrderTable(orders []Order) []string {
	infoLines := []string{}
	infoLines = append(infoLines, "╭────────────────────┬───────────────┬───────────────╮")
	// infoLines = append(infoLines, "│ LOUD price (pylon) │ Amount (loud) │ Total (pylon) │")
	infoLines = append(infoLines, screen.renderOrderTableLine("LOUD price (pylon)", "Amount (loud)", "Total (pylon)", false, false))
	infoLines = append(infoLines, "├────────────────────┼───────────────┼───────────────┤")
	numLines := screen.screenSize.Height/2 - 7
	if screen.activeLine >= len(orders) {
		screen.activeLine = len(orders) - 1
	}
	activeLine := screen.activeLine
	startLine := activeLine - numLines + 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + numLines
	if endLine > len(orders) {
		endLine = len(orders)
	}
	for li, order := range orders[startLine:endLine] {
		infoLines = append(
			infoLines,
			screen.renderOrderTableLine(
				fmt.Sprintf("%.4f", order.Price),
				fmt.Sprintf("%d", order.Amount),
				fmt.Sprintf("%d", order.Total),
				startLine+li == activeLine,
				order.IsMyOrder,
			),
		)
	}
	infoLines = append(infoLines, "╰────────────────────┴───────────────┴───────────────╯")
	return infoLines
}

func (screen *GameScreen) renderItemOrderTableLine(text1 string, text2 string, isActiveLine bool, isDisabledLine bool) string {
	calcText := "│" + centerText(text1, " ", 36) + "│" + centerText(text2, " ", 15) + "│"
	if isActiveLine && isDisabledLine {
		onColor := screen.brownBoldFont()
		return onColor(calcText)
	} else if isActiveLine {
		onColor := screen.blueBoldFont()
		return onColor(calcText)
	} else if isDisabledLine {
		onColor := screen.brownFont()
		return onColor(calcText)
	}
	return calcText
}

func (screen *GameScreen) renderItemOrderTable(orders []ItemOrder) []string {
	infoLines := []string{}
	infoLines = append(infoLines, "╭────────────────────────────────────┬───────────────╮")
	// infoLines = append(infoLines, "│ Item                │ Price (pylon) │")
	infoLines = append(infoLines, screen.renderItemOrderTableLine("Item", "Price (pylon)", false, false))
	infoLines = append(infoLines, "├────────────────────────────────────┼───────────────┤")
	numLines := screen.screenSize.Height/2 - 7
	if screen.activeLine >= len(orders) {
		screen.activeLine = len(orders) - 1
	}
	activeLine := screen.activeLine
	startLine := activeLine - numLines + 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + numLines
	if endLine > len(orders) {
		endLine = len(orders)
	}
	for li, order := range orders[startLine:endLine] {
		infoLines = append(
			infoLines,
			screen.renderItemOrderTableLine(
				fmt.Sprintf("%s Lv%d  ", localize(order.TItem.Name), order.TItem.Level),
				fmt.Sprintf("%d", order.Price),
				startLine+li == activeLine,
				order.IsMyOrder,
			),
		)
	}
	infoLines = append(infoLines, "╰────────────────────────────────────┴───────────────╯")
	return infoLines
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

func (screen *GameScreen) drawFill(x, y, width, height int) {
	color := ansi.ColorCode(fmt.Sprintf("0:%v", bgcolor))

	midString := fmt.Sprintf("%%s%%s%%%vs", (width))
	for i := 0; i <= height; i++ {
		io.WriteString(os.Stdout, fmt.Sprintf(midString, cursor.MoveTo(y+i, x), color, " "))
	}
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

func (screen *GameScreen) InputActive() bool {
	switch screen.scrStatus {
	case CREATE_BUY_LOUD_ORDER_ENTER_LOUD_VALUE:
		return true
	case CREATE_BUY_LOUD_ORDER_ENTER_PYLON_VALUE:
		return true
	case CREATE_SELL_LOUD_ORDER_ENTER_LOUD_VALUE:
		return true
	case CREATE_SELL_LOUD_ORDER_ENTER_PYLON_VALUE:
		return true
	}
	return false
}

func (screen *GameScreen) renderInputValue() {
	inputBoxWidth := uint32(screen.screenSize.Width/2) - 2
	inputWidth := inputBoxWidth - 9
	move := cursor.MoveTo(screen.screenSize.Height-1, 2)

	chatFunc := screen.colorFunc(fmt.Sprintf("231:%v", bgcolor))
	chat := chatFunc("INPUT▶ ")
	fmtString := fmt.Sprintf("%%-%vs", inputWidth)

	if screen.InputActive() {
		chatFunc = screen.colorFunc(fmt.Sprintf("0+b:%v", bgcolor-1))
	}

	fixedChat := truncateLeft(screen.inputText, int(inputWidth))

	inputText := fmt.Sprintf("%s%s%s", move, chat, chatFunc(fmt.Sprintf(fmtString, fixedChat)))

	io.WriteString(os.Stdout, inputText)
}

func (screen *GameScreen) renderCharacterSheet() {
	var HP uint64 = 10
	var MaxHP uint64 = 10
	bgcolor := uint64(bgcolor)
	warning := ""
	if float32(HP) < float32(MaxHP)*.25 {
		bgcolor = 124
		warning = localize("health low warning")
	} else if float32(HP) < float32(MaxHP)*.1 {
		bgcolor = 160
		warning = localize("health critical warning")
	}

	x := screen.screenSize.Width/2 - 1
	width := (screen.screenSize.Width - x)
	fmtFunc := screen.colorFunc(fmt.Sprintf("255:%v", bgcolor))

	infoLines := []string{
		centerText(fmt.Sprintf("%v", screen.user.GetUserName()), " ", width),
		centerText(warning, "─", width),
		screen.pylonIcon() + fmtFunc(truncateRight(fmt.Sprintf(" %s: %v", "Pylon", screen.user.GetPylonAmount()), width-1)),
		screen.loudIcon() + fmtFunc(truncateRight(fmt.Sprintf(" %s: %v", localize("gold"), screen.user.GetGold()), width-1)),
		screen.drawProgressMeter(HP, MaxHP, 196, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" HP: %v/%v", HP, MaxHP), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 225, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" XP: %v/%v", HP, 10), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 208, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" AP: %v/%v", HP, MaxHP), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 117, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" RP: %v/%v", HP, MaxHP), width-10)),
		// screen.drawProgressMeter(HP, MaxHP, 76, bgcolor, 10) + fmtFunc(truncateRight(fmt.Sprintf(" MP: %v/%v", HP, MaxHP), width-10)),
	}

	infoLines = append(infoLines, centerText(localize("inventory items"), "─", width))
	items := screen.user.InventoryItems()
	for _, item := range items {
		infoLines = append(infoLines, truncateRight(fmt.Sprintf("%s Lv%d", localize(item.Name), item.Level), width))
	}
	infoLines = append(infoLines, centerText(" ❦ ", "─", width))

	for index, line := range infoLines {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s", cursor.MoveTo(2+index, x), fmtFunc(line)))
		if index+2 > int(screen.screenSize.Height) {
			break
		}
	}

	nodeLines := []string{
		centerText(localize("pylons network status"), " ", width),
		centerText(screen.user.GetLastTransaction(), " ", width),
	}

	blockHeightText := centerText(localize("block height")+": "+strconv.FormatInt(screen.blockHeight, 10), " ", width)
	if screen.refreshingDaemonStatus {
		nodeLines = append(nodeLines, screen.blueBoldFont()(blockHeightText))
	} else {
		nodeLines = append(nodeLines, blockHeightText)
	}
	nodeLines = append(nodeLines, centerText(" ❦ ", "─", width))

	for index, line := range nodeLines {
		io.WriteString(os.Stdout, fmt.Sprintf("%s%s", cursor.MoveTo(2+len(infoLines)+index, x), fmtFunc(line)))
		if index+2 > int(screen.screenSize.Height) {
			break
		}
	}

	lastLine := len(infoLines) + len(nodeLines) + 1
	screen.drawFill(x, lastLine+1, width, screen.screenSize.Height-(lastLine+2))
}

func (screen *GameScreen) RunSelectedLoudBuyTrade() {
	if len(buyOrders) <= screen.activeLine || screen.activeLine < 0 {
		// when activeLine is not refering to real order but when it is refering to nil order
		screen.txFailReason = localize("you haven't selected any buy order")
		screen.scrStatus = RESULT_FULFILL_BUY_LOUD_ORDER
		screen.refreshed = false
		screen.Render()
	} else {
		screen.scrStatus = WAIT_FULFILL_BUY_LOUD_ORDER
		screen.activeOrder = buyOrders[screen.activeLine]
		screen.refreshed = false
		screen.Render()
		go func() {
			txhash, err := FulfillTrade(screen.user, buyOrders[screen.activeLine].ID)

			log.Println("ended sending request for creating buy loud order")
			if err != nil {
				screen.txFailReason = err.Error()
				screen.scrStatus = RESULT_FULFILL_BUY_LOUD_ORDER
				screen.refreshed = false
				screen.Render()
			} else {
				time.AfterFunc(2*time.Second, func() {
					screen.txResult, screen.txFailReason = ProcessTxResult(screen.user, txhash)
					screen.scrStatus = RESULT_FULFILL_BUY_LOUD_ORDER
					screen.refreshed = false
					screen.Render()
				})
			}
		}()
	}
}

func (screen *GameScreen) RunSelectedLoudSellTrade() {
	if len(sellOrders) <= screen.activeLine || screen.activeLine < 0 {
		screen.txFailReason = localize("you haven't selected any sell order")
		screen.scrStatus = RESULT_FULFILL_SELL_LOUD_ORDER
		screen.refreshed = false
		screen.Render()
	} else {
		screen.scrStatus = WAIT_FULFILL_SELL_LOUD_ORDER
		screen.activeOrder = sellOrders[screen.activeLine]
		screen.refreshed = false
		screen.Render()
		go func() {
			log.Println("started sending request for creating sell loud order")
			txhash, err := FulfillTrade(screen.user, sellOrders[screen.activeLine].ID)
			log.Println("ended sending request for creating sell loud order")
			if err != nil {
				screen.txFailReason = err.Error()
				screen.scrStatus = RESULT_FULFILL_SELL_LOUD_ORDER
				screen.refreshed = false
				screen.Render()
			} else {
				time.AfterFunc(2*time.Second, func() {
					screen.txResult, screen.txFailReason = ProcessTxResult(screen.user, txhash)
					screen.scrStatus = RESULT_FULFILL_SELL_LOUD_ORDER
					screen.refreshed = false
					screen.Render()
				})
			}
		}()
	}
}

func (screen *GameScreen) RunSelectedSwordBuyOrder() {
	if len(swordBuyOrders) <= screen.activeLine || screen.activeLine < 0 {
		screen.txFailReason = localize("you haven't selected any buy item order")
		screen.scrStatus = RESULT_FULFILL_PYLON_SWORD_ORDER
		screen.refreshed = false
		screen.Render()
	} else {
		screen.scrStatus = WAIT_FULFILL_PYLON_SWORD_ORDER
		screen.activeItemOrder = swordBuyOrders[screen.activeLine]
		screen.refreshed = false
		screen.Render()
		go func() {
			log.Println("started sending request for creating buying item order")
			txhash, err := FulfillTrade(screen.user, swordBuyOrders[screen.activeLine].ID)
			log.Println("ended sending request for creating buying item order")
			if err != nil {
				screen.txFailReason = err.Error()
				screen.scrStatus = RESULT_FULFILL_PYLON_SWORD_ORDER
				screen.refreshed = false
				screen.Render()
			} else {
				time.AfterFunc(2*time.Second, func() {
					screen.txResult, screen.txFailReason = ProcessTxResult(screen.user, txhash)
					screen.scrStatus = RESULT_FULFILL_PYLON_SWORD_ORDER
					screen.refreshed = false
					screen.Render()
				})
			}
		}()
	}
}

func (screen *GameScreen) RunSelectedSwordSellOrder() {
	if len(swordSellOrders) <= screen.activeLine || screen.activeLine < 0 {
		screen.txFailReason = localize("you haven't selected any sell item order")
		screen.scrStatus = RESULT_FULFILL_SWORD_PYLON_ORDER
		screen.refreshed = false
		screen.Render()
	} else {
		screen.scrStatus = WAIT_FULFILL_SWORD_PYLON_ORDER
		screen.activeItemOrder = swordSellOrders[screen.activeLine]
		screen.refreshed = false
		screen.Render()
		go func() {
			log.Println("started sending request for creating selling item order")
			txhash, err := FulfillTrade(screen.user, swordSellOrders[screen.activeLine].ID)
			log.Println("ended sending request for creating selling item order")
			if err != nil {
				screen.txFailReason = err.Error()
				screen.scrStatus = RESULT_FULFILL_SWORD_PYLON_ORDER
				screen.refreshed = false
				screen.Render()
			} else {
				time.AfterFunc(2*time.Second, func() {
					screen.txResult, screen.txFailReason = ProcessTxResult(screen.user, txhash)
					screen.scrStatus = RESULT_FULFILL_SWORD_PYLON_ORDER
					screen.refreshed = false
					screen.Render()
				})
			}
		}()
	}
}

func (screen *GameScreen) Render() {
	var HP uint64 = 10

	if screen.screenSize.Height < 20 || screen.screenSize.Width < 60 {
		clear := cursor.ClearEntireScreen()
		move := cursor.MoveTo(1, 1)
		io.WriteString(os.Stdout,
			fmt.Sprintf("%s%s%s", clear, move, localize("screen size warning")))
		return
	} else if HP == 0 {
		clear := cursor.ClearEntireScreen()
		dead := localize("dead")
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
	screen.renderInputValue()
}
