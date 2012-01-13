package house

import (
  "glop/gui"
  "glop/gin"
  "reflect"
  "gl"
)

type CellPos struct {
  X,Y int
}
func (cp CellPos) InRange(size RoomSize) bool {
  return cp.X >= 0 && cp.X < size.Dx && cp.Y >= 0 && cp.Y < size.Dy
}

type CellData struct {
  CanHaveDoor       bool  `Doors`
  CanSpawnExplorers bool  `Spawn Explorers`
  CanSpawnOthers    bool  `Spawn Others`
  CanBeGoal         bool  `Be Goal`
}

func (cd CellData) Render(x,y,dx,dy int) {
  if !cd.CanHaveDoor {
    gl.Color4d(1, 0, 0, 0.4)
    gl.Begin(gl.QUADS)
    if x == 0 {
      gl.Vertex3i(x, y, 0)
      gl.Vertex3i(x, y, -1)
      gl.Vertex3i(x, y + 1, -1)
      gl.Vertex3i(x, y + 1, 0)
    }
    if y == 0 {
      gl.Vertex3i(x, y, 0)
      gl.Vertex3i(x, y, -1)
      gl.Vertex3i(x + 1, y, -1)
      gl.Vertex3i(x + 1, y, 0)
    }
    if x == dx - 1 {
      gl.Vertex3i(x + 1, y, 0)
      gl.Vertex3i(x + 1, y, -1)
      gl.Vertex3i(x + 1, y + 1, -1)
      gl.Vertex3i(x + 1, y + 1, 0)
    }
    if y == dy - 1 {
      gl.Vertex3i(x, y + 1, 0)
      gl.Vertex3i(x, y + 1, -1)
      gl.Vertex3i(x + 1, y + 1, -1)
      gl.Vertex3i(x + 1, y + 1, 0)
    }
    gl.End()
  }

  size := 0.3
  border := 0.1
  if cd.CanSpawnExplorers {
    sx := float64(x) + border
    sy := float64(y) + border
    gl.Color4d(0, 1, 0, 0.7)
    gl.Begin(gl.QUADS)
    gl.Vertex2d(sx, sy)
    gl.Vertex2d(sx, sy + size)
    gl.Vertex2d(sx + size, sy + size)
    gl.Vertex2d(sx + size, sy)
    gl.End()
  }
  if cd.CanSpawnOthers {
    sx := float64(x) + 2 * border + size
    sy := float64(y) + border
    gl.Color4d(0.4, 0, 1, 0.7)
    gl.Begin(gl.QUADS)
    gl.Vertex2d(sx, sy)
    gl.Vertex2d(sx, sy + size)
    gl.Vertex2d(sx + size, sy + size)
    gl.Vertex2d(sx + size, sy)
    gl.End()
  }
  if cd.CanBeGoal {
    sx := float64(x) + border
    sy := float64(y) + 2 * border + size
    gl.Color4d(0.7, 0.7, 0, 0.7)
    gl.Begin(gl.QUADS)
    gl.Vertex2d(sx, sy)
    gl.Vertex2d(sx, sy + size)
    gl.Vertex2d(sx + size, sy + size)
    gl.Vertex2d(sx + size, sy)
    gl.End()
  }
}

type CellPanel struct {
  *gui.VerticalTable
  room *roomDef
  viewer *RoomViewer

  select_mode selectMode
  data map[string]bool
}

func MakeCellPanel(room *roomDef, viewer *RoomViewer) *CellPanel {
  var cp CellPanel
  cp.room = room
  cp.viewer = viewer
  cp.VerticalTable = gui.MakeVerticalTable()
  cp.data = make(map[string]bool)
  cell := reflect.TypeOf(CellData{})
  var options []string
  for i := 0; i < cell.NumField(); i++ {
    field := cell.Field(i)
    if len(field.Tag) > 0 && field.Type.Kind() == reflect.Bool {
      options = append(options, string(field.Tag))
    }
  }
  cp.VerticalTable.AddChild(gui.MakeCheckTextBox(options, 300, cp.data))
  return &cp
}

// Run when a cell is added to the selection, so that we can update the state of
// the check boxes.
func (w *CellPanel) updateAddCell(c CellData) {
  typ := reflect.TypeOf(c)
  val := reflect.ValueOf(c)
  for i := 0; i < typ.NumField(); i++ {
    field := typ.Field(i)
    if len(field.Tag) > 0 && field.Type.Kind() == reflect.Bool {
      if state,ok := w.data[string(field.Tag)]; ok && state != val.Field(i).Bool() {
        delete(w.data, string(field.Tag))
      }
    }
  }
}

func (w *CellPanel) updateInitWithCell(c CellData) {
  typ := reflect.TypeOf(c)
  val := reflect.ValueOf(c)
  for i := 0; i < typ.NumField(); i++ {
    field := typ.Field(i)
    if len(field.Tag) > 0 && field.Type.Kind() == reflect.Bool {
      w.data[string(field.Tag)] = val.Field(i).Bool()
    }
  }
}

// Run when a cell is removed from the selection, so that we can update the
// state of the check boxes.
func (w *CellPanel) updateRemoveCell() {
  if len(w.viewer.Selected.Cells) == 0 {
    w.updateInitWithCell(CellData{})
    return
  }
  for pos := range w.viewer.Selected.Cells {
    w.updateInitWithCell(w.room.Cell_data[pos.X][pos.Y])
    break
  }
  for pos := range w.viewer.Selected.Cells {
    w.updateAddCell(w.room.Cell_data[pos.X][pos.Y])
  }
}

// Run when the data on all selected cells might have been modified
func (w *CellPanel) updateModifyCells() {
  typ := reflect.TypeOf(CellData{})
  for i := 0; i < typ.NumField(); i++ {
    field := typ.Field(i)
    if len(field.Tag) > 0 && field.Type.Kind() == reflect.Bool {
      if state,ok := w.data[string(field.Tag)]; ok {
        for pos := range w.viewer.Selected.Cells {
          p := reflect.ValueOf(&w.room.Cell_data[pos.X][pos.Y])
          v := reflect.Indirect(p)
          v.Field(i).Set(reflect.ValueOf(state))
        }
      }
    }
  }
}

func (w *CellPanel) Respond(ui *gui.Gui, group gui.EventGroup) (consumed bool) {
  if w.VerticalTable.Respond(ui, group) {
    w.updateModifyCells()
    return true
  }
  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    w.select_mode = modeNoSelect
    w.viewer.Selected.Cells = make(map[CellPos]bool)
    w.updateRemoveCell()
    return true
  }

  cursor := group.Events[0].Key.Cursor()
  if cursor != nil && cursor.Name() == "Mouse" {
    bx,by := w.viewer.WindowToBoard(group.Events[0].Key.Cursor().Point())
    pos := CellPos{ int(bx), int(by) }
    if bx < 0 { pos.X-- }
    if by < 0 { pos.Y-- }
    if found,event := group.FindEvent(gin.MouseLButton); found {
      if event.Type == gin.Press {
        if w.viewer.Selected.Cells[pos] {
          w.select_mode = modeDeselect
        } else {
          w.select_mode = modeSelect
        }
      } else if event.Type == gin.Release {
        w.select_mode = modeNoSelect
      }
      consumed = true
    }
    if w.select_mode == modeSelect {
      if pos.InRange(w.room.Size) {
        if _,ok := w.viewer.Selected.Cells[pos]; !ok {
          w.viewer.Selected.Cells[pos] = true
          if len(w.viewer.Selected.Cells) == 1 {
            w.updateInitWithCell(w.room.Cell_data[pos.X][pos.Y])
          } else {
            w.updateAddCell(w.room.Cell_data[pos.X][pos.Y])
          }
        }
      }
    }
    if w.select_mode == modeDeselect {
      delete(w.viewer.Selected.Cells, pos)
      w.updateRemoveCell()
    }
  }
  return
}

func (w *CellPanel) Think(ui *gui.Gui, t int64) {
  w.VerticalTable.Think(ui, t)
}

func (w *CellPanel) Reload() {
}

func (w *CellPanel) Collapse() {
  w.select_mode = modeNoSelect
}

func (w *CellPanel) Expand() {
  w.viewer.SetEditMode(editCells)

  // If the size of the room has changed we'll want to clear out any selected
  // cells that no longer exist
  var pos []CellPos
  for k := range w.viewer.Selected.Cells {
    if !k.InRange(w.room.Size) {
      pos = append(pos, k)
    }
  }
  for _,k := range pos {
    delete(w.viewer.Selected.Cells, k)
  }
}
