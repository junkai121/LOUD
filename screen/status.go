package screen

type ScreenStatus string

const (
	SHOW_LOCATION = "SHOW_LOCATION"
	// in shop
	SELECT_SELL_ITEM   = "SELECT_SELL_ITEM"
	WAIT_SELL_PROCESS  = "WAIT_SELL_PROCESS"
	RESULT_SELL_FINISH = "RESULT_SELL_FINISH"

	SELECT_BUY_ITEM        = "SELECT_BUY_ITEM"
	WAIT_BUY_ITEM_PROCESS  = "WAIT_BUY_ITEM_PROCESS"
	RESULT_BUY_ITEM_FINISH = "RESULT_BUY_ITEM_FINISH"

	SELECT_BUY_CHARACTER        = "SELECT_BUY_CHARACTER"
	WAIT_BUY_CHARACTER_PROCESS  = "WAIT_BUY_CHARACTER_PROCESS"
	RESULT_BUY_CHARACTER_FINISH = "RESULT_BUY_CHARACTER_FINISH"

	SELECT_UPGRADE_ITEM   = "SELECT_UPGRADE_ITEM"
	WAIT_UPGRADE_PROCESS  = "WAIT_UPGRADE_PROCESS"
	RESULT_UPGRADE_FINISH = "RESULT_UPGRADE_FINISH"
	// in forest
	SELECT_HUNT_ITEM   = "SELECT_HUNT_ITEM"
	WAIT_HUNT_PROCESS  = "WAIT_HUNT_PROCESS"
	RESULT_HUNT_FINISH = "RESULT_HUNT_FINISH"
	WAIT_GET_PYLONS    = "WAIT_GET_PYLONS"
	RESULT_GET_PYLONS  = "RESULT_GET_PYLONS"

	// in develop
	WAIT_CREATE_COOKBOOK   = "WAIT_CREATE_COOKBOOK"
	RESULT_CREATE_COOKBOOK = "RESULT_CREATE_COOKBOOK"
	WAIT_SWITCH_USER       = "WAIT_SWITCH_USER"
	RESULT_SWITCH_USER     = "RESULT_SWITCH_USER"

	// in market
	SHOW_LOUD_BUY_REQUESTS                    = "SHOW_LOUD_BUY_REQUESTS"                   // navigation using arrow and list should be sorted by price
	CREATE_BUY_LOUD_REQUEST_ENTER_LOUD_VALUE  = "CREATE_BUY_LOUD_REQUEST_ENTER_LOUD_VALUE" // enter value after switching enter mode
	CREATE_BUY_LOUD_REQUEST_ENTER_PYLON_VALUE = "CREATE_BUY_LOUD_REQUEST_ENTER_PYLON_VALUE"
	WAIT_BUY_LOUD_REQUEST_CREATION            = "WAIT_BUY_LOUD_REQUEST_CREATION"
	RESULT_BUY_LOUD_REQUEST_CREATION          = "RESULT_BUY_LOUD_REQUEST_CREATION"
	WAIT_FULFILL_BUY_LOUD_REQUEST             = "WAIT_FULFILL_BUY_LOUD_REQUEST" // after done go to show loud buy requests
	RESULT_FULFILL_BUY_LOUD_REQUEST           = "RESULT_FULFILL_BUY_LOUD_REQUEST"

	SHOW_LOUD_SELL_REQUESTS                    = "SHOW_LOUD_SELL_REQUESTS"
	CREATE_SELL_LOUD_REQUEST_ENTER_LOUD_VALUE  = "CREATE_SELL_LOUD_REQUEST_ENTER_LOUD_VALUE"
	CREATE_SELL_LOUD_REQUEST_ENTER_PYLON_VALUE = "CREATE_SELL_LOUD_REQUEST_ENTER_PYLON_VALUE"
	WAIT_SELL_LOUD_REQUEST_CREATION            = "WAIT_SELL_LOUD_REQUEST_CREATION"
	RESULT_SELL_LOUD_REQUEST_CREATION          = "RESULT_SELL_LOUD_REQUEST_CREATION"
	WAIT_FULFILL_SELL_LOUD_REQUEST             = "WAIT_FULFILL_SELL_LOUD_REQUEST"
	RESULT_FULFILL_SELL_LOUD_REQUEST           = "RESULT_FULFILL_SELL_LOUD_REQUEST"

	SHOW_SELL_SWORD_REQUESTS                    = "SHOW_SELL_SWORD_REQUESTS"
	CREATE_SELL_SWORD_REQUEST_SELECT_SWORD      = "CREATE_SELL_SWORD_REQUEST_SELECT_SWORD"
	CREATE_SELL_SWORD_REQUEST_ENTER_PYLON_VALUE = "CREATE_SELL_SWORD_REQUEST_ENTER_PYLON_VALUE"
	WAIT_SELL_SWORD_REQUEST_CREATION            = "WAIT_SELL_SWORD_REQUEST_CREATION"
	RESULT_SELL_SWORD_REQUEST_CREATION          = "RESULT_SELL_SWORD_REQUEST_CREATION"
	WAIT_FULFILL_SELL_SWORD_REQUEST             = "WAIT_FULFILL_SELL_SWORD_REQUEST"
	RESULT_FULFILL_SELL_SWORD_REQUEST           = "RESULT_FULFILL_SELL_SWORD_REQUEST"

	SHOW_BUY_SWORD_REQUESTS                    = "SHOW_BUY_SWORD_REQUESTS"
	CREATE_BUY_SWORD_REQUEST_SELECT_SWORD      = "CREATE_BUY_SWORD_REQUEST_SELECT_SWORD"
	CREATE_BUY_SWORD_REQUEST_ENTER_PYLON_VALUE = "CREATE_BUY_SWORD_REQUEST_ENTER_PYLON_VALUE"
	WAIT_BUY_SWORD_REQUEST_CREATION            = "WAIT_BUY_SWORD_REQUEST_CREATION"
	RESULT_BUY_SWORD_REQUEST_CREATION          = "RESULT_BUY_SWORD_REQUEST_CREATION"
	WAIT_FULFILL_BUY_SWORD_REQUEST             = "WAIT_FULFILL_BUY_SWORD_REQUEST"
	RESULT_FULFILL_BUY_SWORD_REQUEST           = "RESULT_FULFILL_BUY_SWORD_REQUEST"
)
