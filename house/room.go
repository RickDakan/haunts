package house

import (
  "glop/gui"
  "path/filepath"
  "fmt"
  "os"
  "io"
  "haunts/base"
)

var (
  datadir string
  tags    Tags
)

// Sets the directory to/from which all data will be saved/loaded
// Also immediately loads/reloads all Tags
func SetDatadir(_datadir string) error {
  datadir = _datadir
  return loadTags()
}

func loadTags() error {
  return base.LoadJson(filepath.Join(datadir, "tags.json"), &tags)
}

type RoomSize struct {
  Name  string
  Dx,Dy int
}
func (r RoomSize) String() string {
  return fmt.Sprintf(r.format(), r.Name, r.Dx, r.Dy)
}
func (r *RoomSize) Scan(str string) {
  fmt.Sscanf(str, r.format(), &r.Name, &r.Dx, &r.Dy)
}
func (r *RoomSize) format() string {
  return "%s (%d, %d)"
}


type Tags struct {
  Themes     []string
  RoomSizes  []RoomSize
  HouseSizes []string
  Decor      []string
}


type WallData struct {
  Door_allowed bool
}

type Room struct {
  Name string
  Size RoomSize

  Furniture []*Furniture

  WallTextures []*WallTexture
  WallData []WallData // TODO: GET RID OF THIS

  // Paths to the floor and wall textures, relative to some basic datadir
  Floor_path string
  Wall_path  string

  Cell_data [][]CellData

  // What house themes this room is appropriate for
  Themes map[string]bool

  // What house sizes this room is appropriate for
  Sizes map[string]bool

  // What kinds of decorations are appropriate in this room
  Decor map[string]bool
}

func MakeRoom() *Room {
  var r Room
  r.Themes = make(map[string]bool)
  r.Sizes = make(map[string]bool)
  r.Decor = make(map[string]bool)
  return &r
}

func (r *Room) Resize(size RoomSize) {
  if len(r.Cell_data) > size.Dx {
    r.Cell_data = r.Cell_data[0 : size.Dx]
  }
  for len(r.Cell_data) < size.Dx {
    r.Cell_data = append(r.Cell_data, []CellData{})
  }
  for i := range r.Cell_data {
    if len(r.Cell_data[i]) > size.Dy {
      r.Cell_data[i] = r.Cell_data[i][0 : size.Dy]
    }
    for len(r.Cell_data[i]) < size.Dy {
      r.Cell_data[i] = append(r.Cell_data[i], CellData{})
    }
  }
  r.Size = size
}

func imagePathFilter(path string, isdir bool) bool {
  if isdir {
    return path[0] != '.'
  }
  ext := filepath.Ext(path)
  return ext == ".jpg" || ext == ".png"
}

type roomError struct {
  ErrorString string
}
func (re *roomError) Error() string {
  return re.ErrorString
}

// If prefix is a prefix of image_path, returns image_path relative to prefix
// Otherwise copies the file at image_path to inside prefix, possibly renaming
// it in the process by appending the value t to the name, then returns the
// path of the new file relative to prefix.
func (room *Room) ensureRelative(prefix, image_path string, t int64) (string, error) {
  if filepath.HasPrefix(image_path, prefix) {
    image_path = image_path[len(prefix) : ]
    if filepath.IsAbs(image_path) {
      image_path = image_path[1 : ]
    }
    return image_path, nil
  }
  target_path := filepath.Join(prefix, filepath.Base(image_path))
  info,err := os.Stat(target_path)
  if err == nil {
    if info.IsDir() {
      return "", &roomError{ fmt.Sprintf("'%s' is a directory, not a file", image_path) }
    }
    base_image := filepath.Base(image_path)
    ext := filepath.Ext(base_image)
    if len(ext) == 0 {
      return "", &roomError{ fmt.Sprintf("Unexpected filename '%s'", base_image) }
    }
    sans_ext := base_image[0 : len(base_image) - len(ext)]
    target_path = filepath.Join(prefix, fmt.Sprintf("%s.%d%s", sans_ext, t, ext))
  }

  source,err := os.Open(image_path)
  if err != nil {
    return "", err
  }
  defer source.Close()

  target,err := os.Create(target_path)
  if err != nil {
    return "", err
  }
  defer target.Close()

  _,err = io.Copy(target, source)
  if err != nil {
    os.Remove(target_path)
    return "", err
  }

  target_path = target_path[len(prefix) + 1: ]
  return target_path, nil
}

// 1. Gob the file to datadir/rooms/<name>.room
// 2. If the wall path is not prefixed by datadir/rooms/walls, copy to datadir/rooms/walls/<name.wall - then fix the wall path
// 3. Same for floor path
func (room *Room) Save(datadir string, t int64) string {
  failed := func(err error) bool {
    // TODO: Log an error
    // TODO: Also save things *somewhere* so data isn't completely lost
    // TODO: Better to just pop up something that says the save failed
    if err != nil {
      fmt.Printf("Failed to save: %v\n", err)
      return true
    }
    return false
  }

  rooms_dir := filepath.Join(datadir, "rooms")
  floors_dir := filepath.Join(rooms_dir, "floors")
  err := os.MkdirAll(floors_dir, 0755)
  if failed(err) { return "" }
  walls_dir := filepath.Join(rooms_dir, "walls")
  err = os.MkdirAll(walls_dir, 0755)
  if failed(err) { return "" }

  target_path := filepath.Join(rooms_dir, room.Name + ".room")
  _,err = os.Stat(target_path)
  if err == nil {
    err = os.Rename(target_path, filepath.Join(rooms_dir, fmt.Sprintf(".%s.%d.room", room.Name, t)))
    if failed(err) { return "" }
  }

  putInDir := func(target_dir, source string, final *string) error {
    if !filepath.IsAbs(source) {
      room.Wall_path = filepath.Clean(filepath.Join(target_dir, source))
    }
    path, err := room.ensureRelative(target_dir, source, t)
    if err != nil {
      *final = ".."
      return err
    }
    rel := filepath.Clean(filepath.Join(target_dir, path))
    *final = base.RelativePath(target_path, rel)
    return nil
  }

  putInDir(floors_dir, room.Floor_path, &room.Floor_path)
  putInDir(walls_dir, room.Wall_path, &room.Wall_path)

  err = base.SaveJson(target_path, room)
  if failed(err) { return "" }
  // TODO: We should definitely gob instead of json encode the rooms, just doing
  // json for now for ease of debugging
  // // Create the target and gob the room to it
  // target, err := os.Create(target_path)
  // if failed(err) { return }
  // defer target.Close()
  // encoder := gob.NewEncoder(target)
  // err = encoder.Encode(room)
  // if failed(err) { return }
  return target_path
}

func LoadRoom(path string) *Room {
  var room Room
  err := base.LoadJson(path, &room)
  if err != nil {
    return nil
  }
  room.Floor_path = filepath.Clean(filepath.Join(path, room.Floor_path))
  room.Wall_path = filepath.Clean(filepath.Join(path, room.Wall_path))
  for i := range room.Furniture {
    room.Furniture[i].Load()
  }

  for i := range room.WallTextures {
    room.WallTextures[i].Load()
  }
  return &room
}

type tabWidget interface {
  Respond(*gui.Gui, gui.EventGroup) bool
  Collapse()
  Expand()
}

type RoomEditorPanel struct {
  *gui.HorizontalTable
  tab *gui.TabFrame
  widgets []tabWidget

  Room       *Room
  RoomViewer *RoomViewer
}


// Manually pass all events to the tabs, regardless of location, since the tabs
// need to know where the user clicks.
func (w *RoomEditorPanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  return w.widgets[w.tab.SelectedTab()].Respond(ui, group)
}

func (w *RoomEditorPanel) SelectTab(n int) {
  if n < 0 || n >= len(w.widgets) { return }
  if n != w.tab.SelectedTab() {
    w.widgets[w.tab.SelectedTab()].Collapse()
    w.tab.SelectTab(n)
    w.RoomViewer.SetEditMode(editNothing)
    w.widgets[n].Expand()
  }
}

func MakeRoomEditorPanel(room *Room, datadir string) *RoomEditorPanel {
  var rep RoomEditorPanel

  rep.Room = room
  rep.HorizontalTable = gui.MakeHorizontalTable()
  rep.RoomViewer = MakeRoomViewer(room, 65)
  for _,wt := range room.WallTextures {
    rep.Room.WallTextures = append(rep.Room.WallTextures, wt)
  }
  rep.RoomViewer.ReloadFloor(room.Floor_path)
  rep.RoomViewer.ReloadWall(room.Wall_path)

  rep.AddChild(rep.RoomViewer)


  var tabs []gui.Widget

  furniture := makeFurniturePanel(room, rep.RoomViewer, datadir)
  tabs = append(tabs, furniture)
  rep.widgets = append(rep.widgets, furniture)

  wall := MakeWallPanel(rep.Room, rep.RoomViewer)
  tabs = append(tabs, wall)
  rep.widgets = append(rep.widgets, wall)

  cells := MakeCellPanel(rep.Room, rep.RoomViewer)
  tabs = append(tabs, cells)
  rep.widgets = append(rep.widgets, cells)

  rep.tab = gui.MakeTabFrame(tabs)
  rep.AddChild(rep.tab)
  rep.RoomViewer.SetEditMode(editFurniture)

  return &rep
}

type selectMode int
const (
  modeNoSelect selectMode = iota
  modeSelect
  modeDeselect
)
