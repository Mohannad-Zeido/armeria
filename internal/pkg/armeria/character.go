package armeria

import (
	"armeria/internal/pkg/misc"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Character struct {
	sync.RWMutex
	UUID                 string            `json:"uuid"`
	UnsafeName           string            `json:"name"`
	UnsafePassword       string            `json:"password"`
	UnsafeLocation       *Location         `json:"location"`
	UnsafeAttributes     map[string]string `json:"attributes"`
	UnsafeTempAttributes map[string]string `json:"-"`
	player               *Player           `json:"-"`
}

const (
	ColorRoomTitle int = 0
	ColorSay       int = 1
	ColorMovement  int = 2
	ColorError     int = 3
	ColorRoomDirs  int = 4
	ColorWhisper   int = 5
	ColorSuccess   int = 6
	ColorCmdHelp   int = 7
)

// UUID returns the uuid of the character.
func (c *Character) Id() string {
	return c.UUID
}

// Type returns the object type, since Character implements the Object interface.
func (c *Character) Type() int {
	return ObjectTypeCharacter
}

// UnsafeName returns the raw character name.
func (c *Character) Name() string {
	c.RLock()
	defer c.RUnlock()
	return c.UnsafeName
}

// FormattedName returns the formatted character name.
func (c *Character) FormattedName() string {
	c.RLock()
	defer c.RUnlock()
	return fmt.Sprintf("[b]%s[/b]", c.UnsafeName)
}

// FormattedNameWithTitle returns the formatted character name including the character's title (if set).
func (c *Character) FormattedNameWithTitle() string {
	c.RLock()
	defer c.RUnlock()

	title := c.UnsafeAttributes["title"]
	if title != "" {
		return fmt.Sprintf("[b]%s[/b] (%s)", c.UnsafeName, title)
	}
	return fmt.Sprintf("[b]%s[/b]", c.UnsafeName)
}

// CheckPassword returns a bool indicating whether the password is correct or not.
func (c *Character) CheckPassword(pw string) bool {
	c.RLock()
	defer c.RUnlock()

	byteHash := []byte(c.UnsafePassword)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(pw))
	if err != nil {
		return false
	}

	return true
}

// SetPassword hashes and sets a new password for the Character.
func (c *Character) SetPassword(pw string) {
	c.Lock()
	defer c.Unlock()

	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		Armeria.log.Fatal("error generating password hash",
			zap.Error(err),
		)
	}

	c.UnsafePassword = string(hash)
}

// PasswordHash returns the character's already-encrypted password as an md5 hash.
func (c *Character) PasswordHash() string {
	c.RLock()
	defer c.RUnlock()

	b := []byte(c.UnsafePassword)
	return fmt.Sprintf("%x", md5.Sum(b))
}

// Player returns the parent that is playing the character.
func (c *Character) Player() *Player {
	c.RLock()
	defer c.RUnlock()

	return c.player
}

// SetPlayer sets the parent that is playing the character.
func (c *Character) SetPlayer(p *Player) {
	c.Lock()
	defer c.Unlock()

	c.player = p
}

// UnsafeLocation returns the character's location.
func (c *Character) Location() *Location {
	c.RLock()
	defer c.RUnlock()

	return c.UnsafeLocation
}

// SetLocation is a helper function to set a Character's Location based on a Location object.
func (c *Character) SetLocation(l *Location) {
	c.Location().SetAreaUUID(l.AreaUUID())
	c.Location().Coords.Set(l.Coords.X(), l.Coords.Y(), l.Coords.Z(), l.Coords.I())
}

// Colorize will color text according to the character's color settings.
func (c *Character) Colorize(text string, color int) string {
	c.RLock()
	defer c.RUnlock()

	switch color {
	case ColorRoomTitle:
		return fmt.Sprintf("<span style='color:#6e94ff;font-weight:600'>%s</span>", text)
	case ColorSay:
		return fmt.Sprintf("<span style='color:#ffeb3b'>%s</span>", text)
	case ColorMovement:
		return fmt.Sprintf("<span style='color:#00bcd4'>%s</span>", text)
	case ColorError:
		return fmt.Sprintf("<span style='color:#e91e63'>%s</span>", text)
	case ColorRoomDirs:
		return fmt.Sprintf("<span style='color:#4c9af3'>%s</span>", text)
	case ColorWhisper:
		return fmt.Sprintf("<span style='color:#b730f7'>%s</span>", text)
	case ColorSuccess:
		return fmt.Sprintf("<span style='color:#8ee22b'>%s</span>", text)
	case ColorCmdHelp:
		return fmt.Sprintf("<span style='color:#e9761e'>%s</span>", text)
	default:
		return text
	}
}

// LoggedIn handles everything that needs to happen when a character enters the game.
func (c *Character) LoggedIn() {
	room := c.Location().Room()
	area := c.Location().Area()

	// Add character to room
	if room == nil || area == nil {
		Armeria.log.Fatal("character logged into an invalid area/room",
			zap.String("character", c.Name()),
		)
		return
	}
	room.AddObjectToRoom(c)

	// Use command: /look
	Armeria.commandManager.ProcessCommand(c.Player(), "look", false)

	// Show message to others in the same room
	roomChars := room.Characters(c)
	for _, char := range roomChars {
		pc := char.Player()
		pc.client.ShowText(
			fmt.Sprintf("%s connected and appeared here with you.", c.Name()),
		)
	}

	area.CharacterEntered(c, true)
	room.CharacterEntered(c, true)

	Armeria.log.Info("character entered the game",
		zap.String("character", c.Name()),
	)
}

// LoggedOut handles everything that needs to happen when a character leaves the game.
func (c *Character) LoggedOut() {
	room := c.Location().Room()
	area := c.Location().Area()

	// Remove character from room
	if room == nil || area == nil {
		Armeria.log.Fatal("character logged out of an invalid area/room",
			zap.String("character", c.Name()),
		)
		return
	}
	room.RemoveObjectFromRoom(c)

	// Show message to others in the same room
	roomChars := room.Characters(nil)
	for _, char := range roomChars {
		pc := char.Player()
		pc.client.ShowText(
			fmt.Sprintf("%s disconnected and is no longer here with you.", c.Name()),
		)
	}

	area.CharacterLeft(c, true)
	room.CharacterLeft(c, true)

	// Clear temp attributes
	for key, _ := range c.UnsafeTempAttributes {
		delete(c.UnsafeTempAttributes, key)
	}

	Armeria.log.Info("character left the game",
		zap.String("character", c.Name()),
	)
}

// TempAttribute retrieves a previously-saved temp attribute.
func (c *Character) TempAttribute(name string) string {
	c.RLock()
	defer c.RUnlock()

	return c.UnsafeTempAttributes[name]
}

// SetTempAttribute sets a temporary attribute, which is cleared on log out. Additionally, these
// attributes are not validated.
func (c *Character) SetTempAttribute(name string, value string) {
	c.Lock()
	defer c.Unlock()

	if c.UnsafeTempAttributes == nil {
		c.UnsafeTempAttributes = make(map[string]string)
	}

	c.UnsafeTempAttributes[name] = value
}

// SetAttribute sets a permanent attribute and only valid attributes can be set.
func (c *Character) SetAttribute(name string, value string) error {
	c.Lock()
	defer c.Unlock()

	if !misc.Contains(ValidCharacterAttributes(), name) {
		return errors.New("attribute name is invalid")
	}

	c.UnsafeAttributes[name] = value
	return nil
}

// Attribute returns a permanent attribute.
func (c *Character) Attribute(name string) string {
	c.RLock()
	defer c.RUnlock()

	if len(c.UnsafeAttributes[name]) == 0 {
		return CharacterAttributeDefault(name)
	}

	return c.UnsafeAttributes[name]
}

// MoveAllowed will check if moving to a particular location is valid/allowed.
func (c *Character) MoveAllowed(to *Location) (bool, string) {
	r := to.Room()
	if r == nil {
		return false, "You cannot move that way."
	}

	if len(c.TempAttribute(TempAttributeGhost)) > 0 {
		return true, ""
	}

	if r.Attribute("type") == "track" {
		return false, "You cannot walk onto the train tracks!"
	}

	return true, ""
}

// Move will move the character to a new location (no move checks are performed).
func (c *Character) Move(to *Location, msgToChar string, msgToOld string, msgToNew string) {
	oldRoom := c.Location().Room()
	newRoom := to.Room()
	oldArea := c.Location().Area()
	newArea := to.Area()

	oldRoom.RemoveObjectFromRoom(c)
	newRoom.AddObjectToRoom(c)

	for _, char := range oldRoom.Characters(nil) {
		char.Player().client.ShowText(msgToOld)
	}

	for _, char := range newRoom.Characters(c) {
		char.Player().client.ShowText(msgToNew)
	}

	c.Player().client.ShowText(msgToChar)

	c.SetLocation(to)

	if oldArea.Id() != newArea.Id() {
		oldArea.CharacterLeft(c, false)
		newArea.CharacterEntered(c, false)
	}

	// Trigger character entered / left events on the new and old rooms, respectively
	newRoom.CharacterEntered(c, false)
	oldRoom.CharacterLeft(c, false)
}

// EditorData returns the JSON used for the object editor.
func (c *Character) EditorData() *ObjectEditorData {
	var props []*ObjectEditorDataProperty
	for _, attrName := range ValidCharacterAttributes() {
		propType := "editable"
		if attrName == AttributePicture {
			propType = "picture"
		}

		props = append(props, &ObjectEditorDataProperty{
			PropType: propType,
			Name:     attrName,
			Value:    c.Attribute(attrName),
		})
	}

	return &ObjectEditorData{
		Name:       c.Name(),
		ObjectType: "character",
		Properties: props,
	}
}

// HasPermission returns true if the Character has a particular permission.
func (c *Character) HasPermission(p string) bool {
	c.RLock()
	defer c.RUnlock()

	perms := strings.Split(c.UnsafeAttributes["permissions"], " ")
	return misc.Contains(perms, p)
}
