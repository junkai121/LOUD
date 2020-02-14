package loud

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Pylons-tech/pylons/x/pylons/handlers"
	"github.com/ahmetb/go-cursor"
)

func (screen *GameScreen) renderUserSituation() {
	infoLines := []string{}
	desc := ""
	waitProcessEnd := localize("wait process to end")
	switch screen.scrStatus {
	case SHOW_LOCATION:
		locationDescMap := map[UserLocation]string{
			HOME:     localize("home desc"),
			FOREST:   localize("forest desc"),
			SHOP:     localize("shop desc"),
			MARKET:   localize("market desc"),
			SETTINGS: localize("settings desc"),
			DEVELOP:  localize("develop desc"),
		}
		desc = locationDescMap[screen.user.GetLocation()]
	case CREATE_BUY_LOUD_REQUEST_ENTER_PYLON_VALUE:
		desc = "Please enter pylon amount to use (should be integer value)" // TODO should add localize
	case CREATE_SELL_LOUD_REQUEST_ENTER_PYLON_VALUE:
		desc = "Please enter pylon amount to get (should be integer value)" // TODO should add localize
	case CREATE_BUY_LOUD_REQUEST_ENTER_LOUD_VALUE:
		desc = "Please enter loud amount to buy (should be integer value)" // TODO should add localize
	case CREATE_SELL_LOUD_REQUEST_ENTER_LOUD_VALUE:
		desc = "Please enter loud amount to sell (should be integer value)" // TODO should add localize

	case SHOW_LOUD_BUY_REQUESTS:
		infoLines = screen.renderTradeRequestTable(buyTradeRequests)
	case SHOW_LOUD_SELL_REQUESTS:
		infoLines = screen.renderTradeRequestTable(sellTradeRequests)
	case SELECT_BUY_ITEM:
		desc = localize("select buy item desc")
	case SELECT_SELL_ITEM:
		desc = localize("select sell item desc")
	case SELECT_HUNT_ITEM:
		desc = localize("select hunt item desc")
	case SELECT_UPGRADE_ITEM:
		desc = localize("select upgrade item desc")
	case WAIT_BUY_LOUD_REQUEST_CREATION:
		desc = localize("you are now waiting for loud buy request creation")
		desc += screen.buyLoudDesc(screen.loudEnterValue, screen.pylonEnterValue)
	case WAIT_SELL_LOUD_REQUEST_CREATION:
		desc = localize("you are now waiting for loud sell request creation")
		desc += screen.sellLoudDesc(screen.loudEnterValue, screen.pylonEnterValue)
	case WAIT_BUY_PROCESS:
		desc = fmt.Sprintf("%s %s Lv%d.\n%s", localize("wait buy process desc"), localize(screen.activeItem.Name), screen.activeItem.Level, waitProcessEnd)
	case WAIT_HUNT_PROCESS:
		if len(screen.activeItem.Name) > 0 {
			desc = fmt.Sprintf("%s %s Lv%d.\n%s", localize("wait hunt process desc"), localize(screen.activeItem.Name), screen.activeItem.Level, waitProcessEnd)
		} else {
			switch string(screen.lastInput.Ch) {
			case "I":
				fallthrough
			case "i":
				desc = fmt.Sprintf("%s\n%s", localize("Getting initial gold from pylon"), waitProcessEnd)
			default:
				desc = fmt.Sprintf("%s\n%s", localize("hunting without weapon"), waitProcessEnd)
			}
		}
	case WAIT_GET_PYLONS:
		desc = localize("You are waiting for getting pylons process")
	case WAIT_SWITCH_USER:
		desc = localize("You are waiting for switching to new user")
	case WAIT_CREATE_COOKBOOK:
		desc = localize("You are waiting for creating cookbook")
	case WAIT_SELL_PROCESS:
		desc = fmt.Sprintf("%s %s Lv%d.\n%s", localize("wait sell process desc"), localize(screen.activeItem.Name), screen.activeItem.Level, waitProcessEnd)
	case WAIT_UPGRADE_PROCESS:
		desc = fmt.Sprintf("%s %s.\n%s", localize("wait upgrade process desc"), localize(screen.activeItem.Name), waitProcessEnd)
	case RESULT_BUY_LOUD_REQUEST_CREATION:
		if screen.txFailReason != "" {
			desc = localize("loud buy request creation fail reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = localize("loud buy request was successfully created")
			desc += screen.buyLoudDesc(screen.loudEnterValue, screen.pylonEnterValue)
		}
	case RESULT_SELL_LOUD_REQUEST_CREATION:
		if screen.txFailReason != "" {
			desc = localize("loud sell request creation fail reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = localize("loud sell request was successfully created")
			desc += screen.sellLoudDesc(screen.loudEnterValue, screen.pylonEnterValue)
		}
	case RESULT_BUY_FINISH:
		if screen.txFailReason != "" {
			desc = localize("buy failed reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = fmt.Sprintf("%s %s Lv%d.\n%s", localize("result buy finish desc"), localize(screen.activeItem.Name), screen.activeItem.Level, localize("use for hunting"))
		}
	case RESULT_HUNT_FINISH:
		if screen.txFailReason != "" {
			desc = localize("hunt failed reason") + ": " + localize(screen.txFailReason)
		} else {
			respOutput := handlers.ExecuteRecipeSerialize{}
			json.Unmarshal(screen.txResult, &respOutput)
			switch string(screen.lastInput.Ch) {
			case "I":
				fallthrough
			case "i":
				desc = fmt.Sprintf("%s %d.", localize("Got initial gold from pylons. Amount is"), respOutput.Amount)
			default:
				desc = fmt.Sprintf("%s %d.", localize("result hunt finish desc"), respOutput.Amount)
			}
		}
	case RESULT_GET_PYLONS:
		if screen.txFailReason != "" {
			desc = localize("get pylon failed reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = fmt.Sprintf("You got extra pylons for loud game")
		}
	case RESULT_SWITCH_USER:
		if screen.txFailReason != "" {
			desc = localize("switch user fail reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = fmt.Sprintf("You switched user to %s", screen.user.GetUserName())
		}
	case RESULT_CREATE_COOKBOOK:
		if screen.txFailReason != "" {
			desc = localize("create cookbook failed reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = fmt.Sprintf("You created a new cookbook for a new game build")
		}
	case RESULT_SELL_FINISH:
		if screen.txFailReason != "" {
			desc = localize("sell failed reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = fmt.Sprintf("%s %s Lv%d.", localize("result sell finish desc"), localize(screen.activeItem.Name), screen.activeItem.Level)
		}
	case RESULT_UPGRADE_FINISH:
		if screen.txFailReason != "" {
			desc = localize("upgrade failed reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = fmt.Sprintf("%s: %s.", localize("result upgrade finish desc"), localize(screen.activeItem.Name))
		}
	case SHOW_BUY_SWORD_REQUESTS:
		infoLines = screen.renderItemTradeRequestTable(swordBuyTradeRequests)
	case CREATE_SELL_SWORD_REQUEST_SELECT_SWORD:
		infoLines = screen.renderItemTable(screen.user.InventoryItems())
	case CREATE_SELL_SWORD_REQUEST_ENTER_PYLON_VALUE:
		desc = "Please enter pylon amount to use (should be integer value)" // TODO should add localize
	case WAIT_SELL_SWORD_REQUEST_CREATION:
		desc = localize("you are now waiting for sword sell request creation")
		desc += screen.sellSwordDesc(screen.activeItem, screen.pylonEnterValue)
	case RESULT_SELL_SWORD_REQUEST_CREATION:
		if screen.txFailReason != "" {
			desc = localize("sword sell request creation fail reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = localize("sword sell request was successfully created")
			desc += screen.sellSwordDesc(screen.activeItem, screen.pylonEnterValue)
		}
	case SHOW_SELL_SWORD_REQUESTS:
		infoLines = screen.renderItemTradeRequestTable(swordSellTradeRequests)
	case CREATE_BUY_SWORD_REQUEST_SELECT_SWORD:
		infoLines = screen.renderItemTable(worldItems)
	case CREATE_BUY_SWORD_REQUEST_ENTER_PYLON_VALUE:
		desc = "Please enter pylon amount to use (should be integer value)" // TODO should add localize
	case WAIT_BUY_SWORD_REQUEST_CREATION:
		desc = localize("you are now waiting for sword buy request creation")
		desc += screen.buySwordDesc(screen.activeItem, screen.pylonEnterValue)
	case RESULT_BUY_SWORD_REQUEST_CREATION:
		if screen.txFailReason != "" {
			desc = localize("sword buy request creation fail reason") + ": " + localize(screen.txFailReason)
		} else {
			desc = localize("sword buy request was successfully created")
			desc += screen.buySwordDesc(screen.activeItem, screen.pylonEnterValue)
		}
	// For FULFILL trades, msg should be reversed, since user is opposite
	case WAIT_FULFILL_BUY_LOUD_REQUEST:
		request := screen.activeTradeRequest
		desc = localize("you are now selling loud for pylon") + fmt.Sprintf(" at %.4f.\n", request.Price)
		desc += screen.sellLoudDesc(request.Amount, request.Total)
	case WAIT_FULFILL_SELL_LOUD_REQUEST:
		request := screen.activeTradeRequest
		desc = localize("you are now buying loud from pylon") + fmt.Sprintf(" at %.4f.\n", request.Price)
		desc += screen.buyLoudDesc(request.Amount, request.Total)
	case RESULT_FULFILL_BUY_LOUD_REQUEST:
		if screen.txFailReason != "" {
			desc = localize("sell loud failed reason") + ": " + localize(screen.txFailReason)
		} else {
			request := screen.activeTradeRequest
			desc = localize("you have sold loud coin successfully from loud/pylon market") + fmt.Sprintf(" at %.4f.\n", request.Price)
			desc += screen.sellLoudDesc(request.Amount, request.Total)
		}
	case RESULT_FULFILL_SELL_LOUD_REQUEST:
		if screen.txFailReason != "" {
			desc = localize("buy loud failed reason") + ": " + localize(screen.txFailReason)
		} else {
			request := screen.activeTradeRequest
			desc = localize("you have bought loud coin successfully from loud/pylon market") + fmt.Sprintf(" at %.4f.\n", request.Price)
			desc += screen.buyLoudDesc(request.Amount, request.Total)
		}
	case WAIT_FULFILL_SELL_SWORD_REQUEST:
		request := screen.activeItemTradeRequest
		desc = localize("you are now buying sword ") + fmt.Sprintf(" at %d.\n", request.Price)
		desc += screen.buySwordDesc(request.TItem, fmt.Sprintf("%d", request.Price))
	case WAIT_FULFILL_BUY_SWORD_REQUEST:
		request := screen.activeItemTradeRequest
		desc = localize("you are now selling sword ") + fmt.Sprintf(" at %d.\n", request.Price)
		desc += screen.sellSwordDesc(request.TItem, fmt.Sprintf("%d", request.Price))
	case RESULT_FULFILL_SELL_SWORD_REQUEST:
		if screen.txFailReason != "" {
			desc = localize("buy sword failed reason") + ": " + localize(screen.txFailReason)
		} else {
			request := screen.activeItemTradeRequest
			desc = localize("you have bought sword successfully from sword/pylon market") + fmt.Sprintf(" at %d.\n", request.Price)
			desc += screen.buySwordDesc(request.TItem, fmt.Sprintf("%d", request.Price))
		}
	case RESULT_FULFILL_BUY_SWORD_REQUEST:
		if screen.txFailReason != "" {
			desc = localize("sell sword failed reason") + ": " + localize(screen.txFailReason)
		} else {
			request := screen.activeItemTradeRequest
			desc = localize("you have sold sword successfully from sword/pylon market") + fmt.Sprintf(" at %d.\n", request.Price)
			desc += screen.sellSwordDesc(request.TItem, fmt.Sprintf("%d", request.Price))
		}
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