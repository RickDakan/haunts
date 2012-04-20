package house

import (
  "fmt"
  "github.com/runningwild/glop/gui"
  "io"
  "os"
  "path/filepath"
  "unsafe"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/mathgl"
  gl "github.com/chsc/gogl/gl21"
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

  // opengl stuff
  // vertex buffer for all of the vertices in the walls and floor
  vbuffer uint32

  // index buffers
  left_buffer  uint32
  right_buffer uint32
  floor_buffer uint32
}

type roomVertex struct {
  x,y,z float32
  u,v float32
}

type plane struct {
  index_buffer uint32
  texture      texture.Object
  mat          *mathgl.Mat4
}

func (room *roomDef) renderFurniture(floor mathgl.Mat4, base_alpha byte) {
  board_to_window := func(mx,my float32) (x,y float32) {
    v := mathgl.Vec4{X: mx, Y: my, W: 1}
    v.Transform(&floor)
    x, y = v.X, v.Y
    return
  }
  for _, furn := range room.Furniture {
    ix,iy := furn.Pos()
    near_x, near_y := float32(ix), float32(iy)
    idx, idy := furn.Dims()
    dx, dy := float32(idx), float32(idy)
    leftx,_ := board_to_window(near_x, near_y + dy)
    rightx,_ := board_to_window(near_x + dx, near_y)
    _,boty := board_to_window(near_x, near_y)
    furn.Render(mathgl.Vec2{leftx, boty}, rightx - leftx, base_alpha)
  }
}

// Need floor, right wall, and left wall matrices to draw the details
func (room *Room) render(floor,left,right mathgl.Mat4, base_alpha byte) {
  gl.Enable(gl.STENCIL_TEST)

  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  gl.Enable(gl.ALPHA_TEST)
  gl.BindBuffer(gl.ARRAY_BUFFER, room.vbuffer)

  gl.EnableClientState(gl.VERTEX_ARRAY)
  gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
  defer gl.DisableClientState(gl.VERTEX_ARRAY)
  defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)

  var vert roomVertex
  gl.VertexPointer(3, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.x)))
  gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.u)))

  planes := []plane{
    {room.left_buffer, room.Wall, &left},
    {room.right_buffer, room.Wall, &right},
    {room.floor_buffer, room.Floor, &floor},
  }

  gl.PushMatrix()
  defer gl.PopMatrix()

  var mul, run mathgl.Mat4
  for _, plane := range planes {
    gl.ClearStencil(0)
    gl.Clear(gl.STENCIL_BUFFER_BIT)
    gl.StencilFunc(gl.ALWAYS, 3, 3)
    gl.StencilOp(gl.KEEP, gl.REPLACE, gl.REPLACE)
    gl.LoadMatrixf(&floor[0])
    run.Assign(&floor)
    // Render the doors and cut out the stencil buffer so we leave them empty
    // if they're open
    switch plane.mat {
      case &left:
      for _, door := range room.Doors {
        door.TextureData().Bind()
        if door.Facing != FarLeft { continue }

        mul.Translation(float32(door.Pos), float32(room.Size.Dy), 0)
        run.Multiply(&mul)
        mul.RotationX(-3.1415926535 / 2)
        run.Multiply(&mul)
        gl.LoadMatrixf(&run[0])

        dx := float64(door.Width)
        dy := dx * float64(door.TextureData().Dy()) / float64(door.TextureData().Dx())
        gl.Color4ub(255, 255, 255, room.far_left.door_alpha)
        door.TextureData().Render(0, 0, dx, dy)
      }

      case &right:
      for _, door := range room.Doors {
        door.TextureData().Bind()
        if door.Facing != FarRight { continue }

        mul.Translation(float32(room.Size.Dx), float32(door.Pos), 0)
        run.Multiply(&mul)
        mul.RotationX(-3.1415926535 / 2)
        run.Multiply(&mul)
        mul.RotationY(-3.1415926535 / 2)
        run.Multiply(&mul)
        gl.LoadMatrixf(&run[0])

        dx := float64(door.Width)
        dy := dx * float64(door.TextureData().Dy()) / float64(door.TextureData().Dx())
        gl.Color4ub(255, 255, 255, room.far_right.door_alpha)
        door.TextureData().Render(0, 0, dx, dy)
      }
    }

    // Now draw the walls
    current_alpha := base_alpha
    switch plane.mat {
      case &left:
      current_alpha = byte((int(room.far_left.wall_alpha) * int(base_alpha)) >> 8)

      case &right:
      current_alpha = byte((int(room.far_right.wall_alpha) * int(base_alpha)) >> 8)
    }
    gl.Color4ub(255, 255, 255, current_alpha)
    gl.LoadMatrixf(&floor[0])
    gl.StencilFunc(gl.NOTEQUAL, 1, 1)
    gl.StencilOp(gl.KEEP, gl.REPLACE, gl.REPLACE)
    plane.texture.Data().Bind()
    gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, plane.index_buffer)
    gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_SHORT, nil)
    gl.StencilFunc(gl.EQUAL, 1, 3)
    gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
    gl.LoadMatrixf(&(*plane.mat)[0])

    // All wall textures need to be drawn three times, once for each wall and
    // once for the floor.
    for i := range room.WallTextures {
      wt := *room.WallTextures[i]
      dx, dy := float32(room.Size.Dx), float32(room.Size.Dy)
      switch plane.mat {
        case &left:
        if wt.X > dx {
          wt.X, wt.Y = dx + dy - wt.Y, dy + wt.X - dx
        }
        wt.Y -= dy

        case &right:
        if wt.Y > dy {
          wt.X, wt.Y = dx + wt.Y - dy, dy + dx - wt.X
        }
        if wt.X > dx {
          wt.Rot -= 3.1415926535 / 2
        }
        wt.X -= dx
      }
      R, G, B, A := wt.GetColor()
      gl.Color4ub(R, G, B, byte((int(current_alpha) * int(A)) >> 8))
      wt.Render()
    }
  }
  gl.Disable(gl.STENCIL_TEST)
  gl.LoadIdentity()
  room.renderFurniture(floor, base_alpha)
}

func (room *roomDef) setupGlStuff() {
  if room.vbuffer != 0 {
    gl.DeleteBuffers(1, &room.vbuffer)
    gl.DeleteBuffers(1, &room.left_buffer)
    gl.DeleteBuffers(1, &room.right_buffer)
    gl.DeleteBuffers(1, &room.floor_buffer)
  }
  dx := float32(room.Size.Dx)
  dy := float32(room.Size.Dy)
  var dz float32
  if room.Wall.Data().Dx() > 0 {
    dz = -float32(room.Wall.Data().Dy() * (room.Size.Dx + room.Size.Dy)) / float32(room.Wall.Data().Dx())
  }

  // c is the u-texcoord of the corner of the room
  c := float32(room.Size.Dx) / float32(room.Size.Dx + room.Size.Dy)

  vs := []roomVertex{
    // Walls
    { 0,  dy,  0, 0, 1},
    {dx,  dy,  0, c, 1},
    {dx,   0,  0, 1, 1},
    { 0,  dy, dz, 0, 0},
    {dx,  dy, dz, c, 0},
    {dx,   0, dz, 1, 0},

    // Floor
    { 0,  0, 0, 0, 1},
    { 0, dy, 0, 0, 0},
    {dx, dy, 0, 1, 0},
    {dx,  0, 0, 1, 1},
  }
  gl.GenBuffers(1, &room.vbuffer)
  gl.BindBuffer(gl.ARRAY_BUFFER, room.vbuffer)
  size := int(unsafe.Sizeof(roomVertex{}))
  gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(size*len(vs)), gl.Pointer(&vs[0].x), gl.STATIC_DRAW)

  // left wall indices
  is := []uint16{0, 3, 4, 0, 4, 1}
  gl.GenBuffers(1, &room.left_buffer)
  gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, room.left_buffer)
  gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(is[0]))*len(is)), gl.Pointer(&is[0]), gl.STATIC_DRAW)

  // right wall indices
  is = []uint16{1, 4, 5, 1, 5, 2}
  gl.GenBuffers(1, &room.right_buffer)
  gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, room.right_buffer)
  gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(is[0]))*len(is)), gl.Pointer(&is[0]), gl.STATIC_DRAW)

  // floor indices
  is = []uint16{6, 7, 8, 6, 8, 9}
  gl.GenBuffers(1, &room.floor_buffer)
  gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, room.floor_buffer)
  gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(is[0]))*len(is)), gl.Pointer(&is[0]), gl.STATIC_DRAW)
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
