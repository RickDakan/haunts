package house

import (
  "fmt"
  "github.com/runningwild/glop/gui"
  "io"
  "os"
  "path/filepath"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
)

func GetAllRoomNames() []string {
  return base.GetAllNamesInRegistry("rooms")
}

func LoadAllRoomsInDir(dir string) {
  base.RemoveRegistry("rooms")
  base.RegisterRegistry("rooms", make(map[string]*roomDef))
  base.RegisterAllObjectsInDir("rooms", dir, ".room", "json")
}

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


type roomDef struct {
  Name string
  Size RoomSize

  Furniture []*Furniture  `registry:"loadfrom-furniture"`

  WallTextures []*WallTexture  `registry:"loadfrom-wall_textures"`

  Floor texture.Object
  Wall  texture.Object

  Cell_data [][]CellData

  // What house themes this room is appropriate for
  Themes map[string]bool

  // What house sizes this room is appropriate for
  Sizes map[string]bool

  // What kinds of decorations are appropriate in this room
  Decor map[string]bool
}

func (room *roomDef) Dims() (dx,dy int) {
  return room.Size.Dx, room.Size.Dy
}

func (r *roomDef) Resize(size RoomSize) {
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
func (room *roomDef) ensureRelative(prefix, image_path string, t int64) (string, error) {
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

type tabWidget interface {
  Respond(*gui.Gui, gui.EventGroup) bool
  Reload()
  Collapse()
  Expand()
}

type RoomEditorPanel struct {
  *gui.HorizontalTable
  tab *gui.TabFrame
  widgets []tabWidget

  panels struct {
    furniture *FurniturePanel
    wall      *WallPanel
    cell      *CellPanel
  }

  room   roomDef
  viewer *RoomViewer
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
    w.viewer.SetEditMode(editNothing)
    w.widgets[n].Expand()
  }
}

func (w *RoomEditorPanel) GetViewer() Viewer {
  return w.viewer
}

type Viewer interface {
  gui.Widget
  Zoom(float64)
  Drag(float64,float64)
  WindowToBoard(int,int) (float32,float32)
  BoardToWindow(float32,float32) (int,int)
}

type Editor interface {
  gui.Widget

  Save() (string, error)
  Load(path string) error

  // Called when we tab into the editor from another editor.  It's possible that
  // a portion of what is being edited in the new editor was changed in another
  // editor, so we reload everything so we can see the up-to-date version.
  Reload()

  GetViewer() Viewer

  // TODO: Deprecate when tabs handle the switching themselves
  SelectTab(int)
}

func MakeRoomEditorPanel() Editor {
  var rep RoomEditorPanel

  rep.HorizontalTable = gui.MakeHorizontalTable()
  rep.viewer = MakeRoomViewer(&rep.room, 65)
  rep.AddChild(rep.viewer)


  var tabs []gui.Widget

  rep.panels.furniture = makeFurniturePanel(&rep.room, rep.viewer)
  tabs = append(tabs, rep.panels.furniture)
  rep.widgets = append(rep.widgets, rep.panels.furniture)

  rep.panels.wall = MakeWallPanel(&rep.room, rep.viewer)
  tabs = append(tabs, rep.panels.wall)
  rep.widgets = append(rep.widgets, rep.panels.wall)

  rep.panels.cell = MakeCellPanel(&rep.room, rep.viewer)
  tabs = append(tabs, rep.panels.cell)
  rep.widgets = append(rep.widgets, rep.panels.cell)

  rep.tab = gui.MakeTabFrame(tabs)
  rep.AddChild(rep.tab)
  rep.viewer.SetEditMode(editFurniture)

  return &rep
}

func (rep *RoomEditorPanel) Load(path string) error {
  var room roomDef
  err := base.LoadAndProcessObject(path, "json", &room)
  if err == nil {
    rep.room = room
    for _,tab := range rep.widgets {
      tab.Reload()
    }
  }
  return err
}

func (rep *RoomEditorPanel) Save() (string, error) {
  path := filepath.Join(datadir, "rooms", rep.room.Name + ".room")
  err := base.SaveJson(path, rep.room)
  return path, err
}

func (rep *RoomEditorPanel) Reload() {
}

type selectMode int
const (
  modeNoSelect selectMode = iota
  modeSelect
  modeDeselect
)
