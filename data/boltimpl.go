package loud

import (
	"log"

	bolt "github.com/coreos/bbolt"
)

type dbWorld struct {
	filename string
	database *bolt.DB
}

func (w *dbWorld) GetUser(username string) User {
	return getUserFromDB(w, username)
}

func (w *dbWorld) newUser(username string) UserData {
	userData := UserData{
		Username: username,
		Location: HOME,
		Gold:     0,
		Items:    []Item{},
	}

	return userData
}

func (w *dbWorld) Close() {
	if w.database != nil {
		w.database.Close()
	}
}

func (w *dbWorld) load() {
	log.Printf("Loading world database %s", w.filename)
	db, err := bolt.Open(w.filename, 0600, nil)

	if err != nil {
		panic(err)
	}

	// Make default tables
	db.Update(func(tx *bolt.Tx) error {
		buckets := []string{"users"}

		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))

			if err != nil {
				return err
			}
		}

		return nil
	})

	w.database = db
}

// LoadWorldFromDB will set up an on-disk based world
func LoadWorldFromDB(filename string) World {
	newWorld := dbWorld{filename: filename}
	newWorld.load()
	return &newWorld
}

// UserData is a JSON-serializable set of information about a User.
type UserData struct {
	Gold        int
	PylonAmount int
	Username    string `json:""`
	Location    UserLocation
	Items       []Item
	PrivKey     string
}

type dbUser struct {
	UserData
	world           *dbWorld
	lastTransaction string
}

func (user *dbUser) GetPrivKey() string {
	return user.UserData.PrivKey
}

func (user *dbUser) GetLocation() UserLocation {
	return user.UserData.Location
}

func (user *dbUser) SetLocation(loc UserLocation) {
	user.UserData.Location = loc
}

func (user *dbUser) Reload() {
	var record []byte
	user.world.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))

		record = bucket.Get([]byte(user.UserData.Username))

		return nil
	})

	if record == nil {
		log.Printf("User %s does not exist, creating anew...", user.UserData.Username)
		user.UserData = user.world.newUser(user.UserData.Username)
		user.Save()
	} else {
		MSGUnpack(record, &(user.UserData))
		log.Printf("Loaded user %v", user.UserData)
	}
	log.Println("start InitPylonAccount")
	user.UserData.PrivKey = InitPylonAccount(user.UserData.Username)
	log.Println("finished InitPylonAccount PrivKey=", user.UserData.PrivKey)
	// Initial Synchronization
	log.Println("start initial synchronization from node")
	SyncFromNode(user)
	log.Println("finished initial synchronization from node")
}

func (user *dbUser) Save() {
	bytes, err := MSGPack(user.UserData)
	if err != nil {
		log.Printf("Can't marshal user: %v", err)
		return
	}

	user.world.database.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))

		err = bucket.Put([]byte(user.UserData.Username), bytes)

		return err
	})
}

func (user *dbUser) GetUserName() string {
	return user.UserData.Username
}

func (user *dbUser) SetGold(amount int) {
	user.UserData.Gold = amount
}
func (user *dbUser) GetGold() int {
	return user.UserData.Gold
}

func (user *dbUser) GetPylonAmount() int {
	return user.UserData.PylonAmount
}

func (user *dbUser) SetPylonAmount(amount int) {
	user.UserData.PylonAmount = amount
}

func (user *dbUser) SetItems(items []Item) {
	user.UserData.Items = items
}

func (user *dbUser) InventoryItems() []Item {
	return user.UserData.Items
}

func (user *dbUser) UpgradableItems() []Item {
	iis := user.InventoryItems()
	uis := []Item{}
	for _, ii := range iis {
		if ii.Level == 1 {
			uis = append(uis, ii)
		}
	}
	return uis
}

func (user *dbUser) GetLastTransaction() string {
	return user.lastTransaction
}

func (user *dbUser) SetLastTransaction(trans string) {
	user.lastTransaction = trans
}

func getUserFromDB(world *dbWorld, username string) User {
	user := dbUser{
		UserData: UserData{
			Username: username,
		},
		world: world,
	}

	user.Reload()

	return &user
}