package house

import (
  "glop/gin"
  "glop/gui"
  "path/filepath"
  "fmt"
  "glop/util/algorithm"
  "os"
  "io"
  "time"
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


type Room struct {
  Name string
  Size RoomSize

  Furniture []*Furniture

  // Paths to the floor and wall textures, relative to some basic datadir
  Floor_path string
  Wall_path  string

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

type RoomEditorPanel struct {
  *gui.HorizontalTable
  name       *gui.TextEditLine
  room_size  *gui.ComboBox
  floor_path *gui.FileWidget
  wall_path  *gui.FileWidget
  themes     *gui.CheckBoxes
  sizes      *gui.CheckBoxes
  decor      *gui.CheckBoxes

  Room       *Room
  RoomViewer *RoomViewer

  // If we're moving around temporary objects in the room or something they'll
  // be stored temporarily here
  object *Furniture

  // Distance from the mouse to the center of the object, in board coordinates
  drag_anchor struct{ x,y float32 }

  // True iff the selected object should be placed when the mouse button is
  // released.  If false this object will be placed when the mouse button is
  // clicked.
  drop_on_release bool
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
  return &room
}

func MakeRoomEditorPanel(room *Room, datadir string) *RoomEditorPanel {
  var rep RoomEditorPanel
  rep.Room = room
  if room.Name == "" {
    room.Name = "name"
  }
  rep.name = gui.MakeTextEditLine("standard", room.Name, 300, 1, 1, 1, 1)  

  if room.Floor_path == "" {
    room.Floor_path = datadir
  }
  fmt.Printf("floor path: %s\n", room.Floor_path)
  rep.floor_path = gui.MakeFileWidget(room.Floor_path, imagePathFilter)

  if room.Wall_path == "" {
    room.Wall_path = datadir
  }
  rep.wall_path = gui.MakeFileWidget(room.Wall_path, imagePathFilter)

  rep.room_size = gui.MakeComboTextBox(algorithm.Map(tags.RoomSizes, []string{}, func(a interface{}) interface{} { return a.(RoomSize).String() }).([]string), 300)
  for i := range tags.RoomSizes {
    if tags.RoomSizes[i].String() == room.Size.String() {
      rep.room_size.SetSelectedIndex(i)
      break
    }
  }
  rep.themes = gui.MakeCheckTextBox(tags.Themes, 300, room.Themes)
  rep.sizes = gui.MakeCheckTextBox(tags.HouseSizes, 300, room.Sizes)
  rep.decor = gui.MakeCheckTextBox(tags.Decor, 300, room.Decor)

  pane := gui.MakeVerticalTable()
  pane.Params().Spacing = 3  
  pane.Params().Background.R = 0.3
  pane.Params().Background.B = 1
  pane.AddChild(rep.name)
  pane.AddChild(rep.floor_path)
  pane.AddChild(rep.wall_path)
  pane.AddChild(rep.room_size)
  pane.AddChild(rep.themes)
  pane.AddChild(rep.sizes)
  pane.AddChild(rep.decor)
  fnames := GetAllFurnitureNames()
  for i := range fnames {
    name := fnames[i]
    pane.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(t int64) {
      f := MakeFurniture(name)
      if f == nil { return }
      fmt.Printf("obj: %v\n", f)
      rep.object = f
      rep.drop_on_release = false
      dx,dy := rep.object.Dims()
      rep.drag_anchor.x = float32(dx - 1) / 2
      rep.drag_anchor.y = float32(dy - 1) / 2
      rep.RoomViewer.SetTempObject(rep.object)
    }))
  }
  pane.AddChild(gui.MakeButton("standard", "Save!", 300, 1, 1, 1, 1, func(t int64) {
    target_path := room.Save(datadir, time.Now().UnixNano())
    if target_path != "" {
      base.SetStoreVal("last room path", target_path)
      // The paths can change when we save them so we should update the widgets
      if !filepath.IsAbs(room.Floor_path) {
        room.Floor_path = filepath.Join(target_path, room.Floor_path)
        rep.floor_path.SetPath(room.Floor_path)
      }
      if !filepath.IsAbs(room.Wall_path) {
        room.Wall_path = filepath.Join(target_path, room.Wall_path)
        rep.wall_path.SetPath(room.Wall_path)
      }
    }
  }))

  rep.HorizontalTable = gui.MakeHorizontalTable()
  rep.RoomViewer = MakeRoomViewer(room.Size.Dx, room.Size.Dy, 65)
  for _,f := range room.Furniture {
    rep.RoomViewer.AddFurniture(f)
  }
  rep.AddChild(rep.RoomViewer)
  rep.AddChild(pane)
  return &rep
}

func (w *RoomEditorPanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if w.HorizontalTable.Respond(ui, group) {
    return true
  }
  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    if w.object != nil {
      w.RoomViewer.SetTempObject(nil)
      w.object = nil
    }
    return true
  }
  if found,event := group.FindEvent(gin.MouseLButton); found {
    if w.object != nil && (event.Type == gin.Press || (event.Type == gin.Release && w.drop_on_release)) {
      w.RoomViewer.SetTempObject(nil)
      w.RoomViewer.AddFurniture(w.object)
      w.Room.Furniture = append(w.Room.Furniture, w.object)
      w.object = nil
    } else if w.object == nil && event.Type == gin.Press {
      bx,by := w.RoomViewer.WindowToBoard(event.Key.Cursor().Point())
      w.object = w.RoomViewer.SelectFurnitureAt(event.Key.Cursor().Point())
      w.Room.Furniture = algorithm.Choose(w.Room.Furniture, func(a interface{}) bool {
        return a.(*Furniture) != w.object
      }).([]*Furniture)
      w.drop_on_release = true
      if w.object != nil {
        px,py := w.object.Pos()
        w.drag_anchor.x = bx - float32(px) - 0.5
        w.drag_anchor.y = by - float32(py) - 0.5
      }
    }
    return true
  }
  return false
}

func (w *RoomEditorPanel) Think(ui *gui.Gui, t int64) {
  if w.object != nil {
    mx,my := gin.In().GetCursor("Mouse").Point()
    bx,by := w.RoomViewer.WindowToBoard(mx, my)
    w.object.X = int(bx - w.drag_anchor.x)
    w.object.Y = int(by - w.drag_anchor.y)
    w.RoomViewer.MoveFurniture()
  }
  w.HorizontalTable.Think(ui, t)
  w.Room.Name = w.name.GetText()

  w.Room.Size = tags.RoomSizes[w.room_size.GetComboedIndex()]
  w.RoomViewer.SetDims(w.Room.Size.Dx, w.Room.Size.Dy)

  w.Room.Floor_path = w.floor_path.GetPath()
  w.Room.Wall_path = w.wall_path.GetPath()

  w.RoomViewer.ReloadFloor(w.Room.Floor_path)
  w.RoomViewer.ReloadWall(w.Room.Wall_path)


  for i := range tags.Themes {
    selected := false
    for _,j := range w.themes.GetSelectedIndexes() {
      if j == i {
        selected = true
        break
      }
    }
    if selected {
      w.Room.Themes[tags.Themes[i]] = true
    } else if _,ok := w.Room.Themes[tags.Themes[i]]; ok {
      delete(w.Room.Themes, tags.Themes[i])
    }
  }

  for i := range tags.Decor {
    selected := false
    for _,j := range w.decor.GetSelectedIndexes() {
      if j == i {
        selected = true
        break
      }
    }
    if selected {
      w.Room.Decor[tags.Decor[i]] = true
    } else if _,ok := w.Room.Decor[tags.Decor[i]]; ok {
      delete(w.Room.Decor, tags.Decor[i])
    }
  }
}
