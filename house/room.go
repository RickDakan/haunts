package house

import (
  "glop/gui"
  "haunts/base"
  "path/filepath"
  "fmt"
  "glop/util/algorithm"
)

var (
  datadir string
  tags    Tags
)

func init() {
  fmt.Printf("")
}

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
  Themes []string
  Sizes  []RoomSize
  Decor  []string
}

type Room struct {
  Name  string
  RoomSize

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
  decor      *gui.CheckBoxes

  Room *Room
}

func MakeRoomEditorPanel(room *Room) (*RoomViewer, *RoomEditorPanel) {
  var rep RoomEditorPanel
  rep.Room = room
  rep.name = gui.MakeTextEditLine("standard", "name", 300, 1, 1, 1, 1)  
  rep.floor_path = gui.MakeFileWidget(datadir)
  rep.wall_path = gui.MakeFileWidget(datadir)
  rep.room_size = gui.MakeComboTextBox(algorithm.Map(tags.Sizes, []string{}, func(a interface{}) interface{} { return a.(RoomSize).String() }).([]string), 300)
  rep.themes = gui.MakeCheckTextBox(tags.Themes, 300)
  rep.decor = gui.MakeCheckTextBox(tags.Decor, 300)

  pane := gui.MakeVerticalTable()
  pane.Params().Spacing = 3  
  pane.Params().Background.R = 0.3
  pane.Params().Background.B = 1
  pane.AddChild(rep.name)
  pane.AddChild(rep.floor_path)
  pane.AddChild(rep.wall_path)
  pane.AddChild(rep.room_size)
  pane.AddChild(rep.themes)
  pane.AddChild(rep.decor)

  rep.HorizontalTable = gui.MakeHorizontalTable()
  viewer := MakeRoomViewer(30, 30, 65)
  viewer.ReloadFloor("/Users/runningwild/Downloads/floor_01.png")
  rep.AddChild(viewer)
  rep.AddChild(pane)
  return viewer, &rep
}

func (w *RoomEditorPanel) Think(ui *gui.Gui, t int64) {
  w.HorizontalTable.Think(ui, t)
  w.Room.Name = w.name.GetText()
  w.Room.RoomSize = tags.Sizes[w.room_size.GetComboedIndex()]
  w.Room.Floor_path = w.floor_path.GetPath()
  w.Room.Wall_path = w.wall_path.GetPath()

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















