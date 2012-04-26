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
}

type roomVertex struct {
  x,y,z float32
  u,v float32

  // Texture coordinates for the los texture
  los_u, los_v float32
}

type plane struct {
  index_buffer uint32
  texture      texture.Object
  mat          *mathgl.Mat4
}

func visibilityOfObject(xoff, yoff int, ro RectObject, los_tex *LosTexture) byte {
  if los_tex == nil {
    return 255
  }
  x, y := ro.Pos()
  x += xoff
  y += yoff
  dx, dy := ro.Dims()
  count := 0
  pix := los_tex.Pix()
  for i := x; i < x + dx; i++ {
    if y - 1 >= 0 && pix[i][y-1] > LosVisibilityThreshold {
      count++
    }
    if y + dy + 1 < LosTextureSize && pix[i][y+dy+1] > LosVisibilityThreshold {
      count++
    }
  }
  for j := y; j < y + dy; j++ {
    if x - 1 > 0 && pix[x-1][j] > LosVisibilityThreshold {
      count++
    }
    if x + dx + 1 < LosTextureSize && pix[x+dx+1][j] > LosVisibilityThreshold {
      count++
    }
  }
  if count >= dx + dy {
    return 255
  }
  v := 256 * float64(count) / float64(dx + dy)
  if v < 0 {
    v = 0
  }
  if v > 255 {
    v = 255
  }
  return byte(v)
}

func (room *Room) renderFurniture(floor mathgl.Mat4, base_alpha byte, drawables []Drawable, los_tex *LosTexture) {
  board_to_window := func(mx,my float32) (x,y float32) {
    v := mathgl.Vec4{X: mx, Y: my, W: 1}
    v.Transform(&floor)
    x, y = v.X, v.Y
    return
  }

  var all []RectObject
  for _, d := range drawables {
    x, y := d.Pos()
    if x < room.X { continue }
    if y < room.Y { continue }
    if x >= room.X + room.Size.Dx { continue }
    if y >= room.Y + room.Size.Dy { continue }
    all = append(all, offsetDrawable{d, -room.X, -room.Y})
  }
  for _, f := range room.Furniture {
    all = append(all, f)
  }
  all = OrderRectObjects(all)

  for i := len(all) - 1; i >= 0; i-- {
    d := all[i].(Drawable)
    fx,fy := d.FPos()
    near_x, near_y := float32(fx), float32(fy)
    idx, idy := d.Dims()
    dx, dy := float32(idx), float32(idy)
    leftx,_ := board_to_window(near_x, near_y + dy)
    rightx,_ := board_to_window(near_x + dx, near_y)
    _,boty := board_to_window(near_x, near_y)
    vis := visibilityOfObject(room.X, room.Y, d, los_tex)
    r, g, b, a := d.Color()
    r = alphaMult(r, vis)
    g = alphaMult(g, vis)
    b = alphaMult(b, vis)
    a = alphaMult(a, vis)
    a = alphaMult(a, base_alpha)
    gl.Color4ub(r, g, b, a)
    d.Render(mathgl.Vec2{leftx, boty}, rightx - leftx)
  }
}

func (room *Room) getNearWallAlpha(los_tex *LosTexture) (left, right byte) {
  if los_tex == nil {
    return 255, 255
  }
  pix := los_tex.Pix()
  v1, v2 := 0, 0
  for y := room.Y; y < room.Y + room.Size.Dy; y++ {
    if pix[room.X][y] > LosVisibilityThreshold {
      v1++
    }
    if pix[room.X - 1][y] > LosVisibilityThreshold {
      v2++
    }
  }
  if v1 < v2 {
    v1 = v2
  }
  right = byte((v1 * 255) / room.Size.Dy)
  v1, v2 = 0, 0
  for x := room.X; x < room.X + room.Size.Dx; x++ {
    if pix[x][room.Y] > LosVisibilityThreshold {
      v1++
    }
    if pix[x][room.Y - 1] > LosVisibilityThreshold {
      v2++
    }
  }
  if v1 < v2 {
    v1 = v2
  }
  left = byte((v1 * 255) / room.Size.Dy)
  return
}

func (room *Room) getMaxLosAlpha(los_tex *LosTexture) byte {
  if los_tex == nil {
    return 255
  }
  var max_room_alpha byte = 0
  pix := los_tex.Pix()
  for x := room.X; x < room.X + room.Size.Dx; x++ {
    for y := room.Y; y < room.Y + room.Size.Dy; y++ {
      if pix[x][y] > max_room_alpha {
        max_room_alpha = pix[x][y]
      }
    }
  }
  max_room_alpha = byte(255 * (float64(max_room_alpha - LosMinVisibility) / float64(255 - LosMinVisibility)))
  return max_room_alpha
}

func alphaMult(a, b byte) byte {
  return byte((int(a) * int(b)) >> 8)
}

// Need floor, right wall, and left wall matrices to draw the details
func (room *Room) render(floor,left,right mathgl.Mat4, base_alpha byte, drawables []Drawable, los_tex *LosTexture) {
  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  gl.Enable(gl.STENCIL_TEST)
  gl.ClearStencil(0)
  gl.Clear(gl.STENCIL_BUFFER_BIT)


  gl.EnableClientState(gl.VERTEX_ARRAY)
  gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
  defer gl.DisableClientState(gl.VERTEX_ARRAY)
  defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)

  var vert roomVertex

  planes := []plane{
    {room.left_buffer, room.Wall, &left},
    {room.right_buffer, room.Wall, &right},
    {room.floor_buffer, room.Floor, &floor},
  }

  gl.PushMatrix()
  defer gl.PopMatrix()

  if los_tex != nil {
    gl.LoadMatrixf(&floor[0])
    gl.ClientActiveTexture(gl.TEXTURE1)
    gl.ActiveTexture(gl.TEXTURE1)
    gl.Enable(gl.TEXTURE_2D)
    gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
    los_tex.Bind()
//      texture.LoadFromPath("/Users/runningwild/code/src/github.com/runningwild/haunts/haunts.app/Contents/textures/orient.png").Bind()
    gl.BindBuffer(gl.ARRAY_BUFFER, room.vbuffer)
    gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.los_u)))
    gl.ClientActiveTexture(gl.TEXTURE0)
    gl.ActiveTexture(gl.TEXTURE0)
    base.EnableShader(true)
  }

  current_alpha := base_alpha
  left_alpha := byte((int(room.far_left.wall_alpha) * int(base_alpha)) >> 8)
  right_alpha := byte((int(room.far_right.wall_alpha) * int(base_alpha)) >> 8)

  var mul, run mathgl.Mat4
  for _, plane := range planes {
    gl.BindBuffer(gl.ARRAY_BUFFER, room.vbuffer)
    gl.VertexPointer(3, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.x)))
    gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.u)))
    gl.LoadMatrixf(&floor[0])
    run.Assign(&floor)

    // Render the doors and cut out the stencil buffer so we leave them empty
    // if they're open
    gl.StencilFunc(gl.ALWAYS, 1, 1)
    gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
    switch plane.mat {
      case &left:
      gl.Color4ub(255, 255, 255, left_alpha)
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
        door.TextureData().Render(0, 0, dx, dy)
      }

      case &right:
      gl.Color4ub(255, 255, 255, right_alpha)
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
        door.TextureData().Render(0, 0, dx, dy)
      }

      case &floor:
      gl.Color4ub(255, 255, 255, current_alpha)
    }

    gl.StencilFunc(gl.NOTEQUAL, 1, 1)
    gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)

    // Now draw the walls
    gl.LoadMatrixf(&floor[0])
    plane.texture.Data().Bind()
    gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, plane.index_buffer)
    gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_SHORT, nil)
  }

  for _, wt := range room.getWallTextures() {
    wt.setupGlStuff(room)
    gl.LoadMatrixf(&floor[0])
    if wt.gl.vbuffer != 0 {
      wt.Texture.Data().Bind()
      R, G, B, A := wt.Color()
      gl.BindBuffer(gl.ARRAY_BUFFER, wt.gl.vbuffer)
      gl.VertexPointer(3, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.x)))
      gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.u)))
      gl.ClientActiveTexture(gl.TEXTURE1)
      gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(unsafe.Sizeof(vert)), gl.Pointer(unsafe.Offsetof(vert.los_u)))
      gl.ClientActiveTexture(gl.TEXTURE0)
      if wt.gl.floor_buffer != 0 {
        gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wt.gl.floor_buffer)
        gl.Color4ub(R, G, B, alphaMult(A, current_alpha))
        gl.DrawElements(gl.TRIANGLES, wt.gl.floor_count, gl.UNSIGNED_SHORT, nil)
      }
      if wt.gl.left_buffer != 0 {
        gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wt.gl.left_buffer)
        gl.Color4ub(R, G, B, alphaMult(A, left_alpha))
        gl.DrawElements(gl.TRIANGLES, wt.gl.left_count, gl.UNSIGNED_SHORT, nil)
      }
      if wt.gl.right_buffer != 0 {
        gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wt.gl.right_buffer)
        gl.Color4ub(R, G, B, alphaMult(A, right_alpha))
        gl.DrawElements(gl.TRIANGLES, wt.gl.right_count, gl.UNSIGNED_SHORT, nil)
      }
    }
  }
  if los_tex != nil {
    base.EnableShader(false)
    gl.ActiveTexture(gl.TEXTURE1)
    gl.Disable(gl.TEXTURE_2D)
    gl.ActiveTexture(gl.TEXTURE0)
    gl.ClientActiveTexture(gl.TEXTURE1)
    gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
    gl.ClientActiveTexture(gl.TEXTURE0)
  }
  gl.Color4ub(255, 255, 255, 255)
  gl.LoadIdentity()
  gl.Disable(gl.STENCIL_TEST)
  room.renderFurniture(floor, base_alpha, drawables, los_tex)
}

func (room *Room) setupGlStuff() {
  if room.X == room.gl.x && room.Y == room.gl.y && room.Size.Dx == room.gl.dx && room.Size.Dy == room.gl.dy {
    return
  }
  base.Log().Printf("SETUP GL")
  room.gl.x = room.X
  room.gl.y = room.Y
  room.gl.dx = room.Size.Dx
  room.gl.dy = room.Size.Dy
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

  // Conveniently casted values
  frx := float32(room.X)
  fry := float32(room.Y)
  frdx := float32(room.Size.Dx)
  frdy := float32(room.Size.Dy)

  // c is the u-texcoord of the corner of the room
  c := frdx / (frdx + frdy)

  lt_llx := frx / LosTextureSize
  lt_lly := fry / LosTextureSize
  lt_urx := (frx + frdx) / LosTextureSize
  lt_ury := (fry + frdy) / LosTextureSize

  vs := []roomVertex{
    // Walls
    { 0,  dy,  0, 0, 1, lt_ury, lt_llx},
    {dx,  dy,  0, c, 1, lt_ury, lt_urx},
    {dx,   0,  0, 1, 1, lt_lly, lt_urx},
    { 0,  dy, dz, 0, 0, lt_ury, lt_llx},
    {dx,  dy, dz, c, 0, lt_ury, lt_urx},
    {dx,   0, dz, 1, 0, lt_lly, lt_urx},

    // Floor
    { 0,  0, 0, 0, 1, lt_lly, lt_llx},
    { 0, dy, 0, 0, 0, lt_ury, lt_llx},
    {dx, dy, 0, 1, 0, lt_ury, lt_urx},
    {dx,  0, 0, 1, 1, lt_lly, lt_urx},
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
