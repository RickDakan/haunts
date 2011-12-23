package house

import (
  "glop/gui"
)

type CellPanel struct {
  *gui.VerticalTable
  room *Room
  viewer *RoomViewer

  selecting_cells selectMode
  selected_cells map[int]bool
}

func MakeCellPanel(room *Room, viewer *RoomViewer) *CellPanel {
  var cp CellPanel
  cp.room = room
  cp.viewer = viewer
  cp.VerticalTable = gui.MakeVerticalTable()
  cp.selected_cells = make(map[int]bool)
  return &cp
}

func (w *CellPanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if w.VerticalTable.Respond(ui, group) {
    return true
  }
  // if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
  //   if w.wall_texture != nil {
  //     w.viewer.SetTempWallTexture(nil)
  //     w.wall_texture = nil
  //   }
  //   return true
  // }
  // if found,event := group.FindEvent(gin.MouseLButton); found {
  //   if w.wall_texture != nil && (event.Type == gin.Press || (event.Type == gin.Release && w.drop_on_release)) {
  //     w.viewer.SetTempWallTexture(nil)
  //     w.viewer.AddWallTexture(w.wall_texture)
  //     w.room.WallTextures = append(w.room.WallTextures, w.wall_texture)
  //     w.wall_texture = nil
  //   } else if w.wall_texture == nil && event.Type == gin.Press {
  //     pos,height := w.viewer.WindowToWall(event.Key.Cursor().Point())
  //     w.wall_texture = w.viewer.SelectWallTextureAt(event.Key.Cursor().Point())
  //     if w.wall_texture != nil {
  //       w.prev_wall_texture = new(WallTexture)
  //       *w.prev_wall_texture = *w.wall_texture
  //     } else {
  //       index := int(pos * float32(len(w.room.WallData)))
  //       if index >= 0 && index < len(w.room.WallData) {
  //         if w.selected_cells[index] {
  //           delete(w.selected_cells, index)
  //           w.selecting_cells = deselect
  //         } else {
  //           w.selected_cells[index] = true
  //           w.selecting_cells = selectOn
  //         }
  //       }
  //     }
  //     w.room.WallTextures = algorithm.Choose(w.room.WallTextures, func(a interface{}) bool {
  //       return a.(*WallTexture) != w.wall_texture
  //     }).([]*WallTexture)
  //     w.drop_on_release = true
  //     if w.wall_texture != nil {
  //       w.drag_anchor.pos = pos - w.wall_texture.Pos
  //       w.drag_anchor.height = height - float32(w.wall_texture.Height)
  //     }
  //   } else if event.Type == gin.Release {
  //     w.selecting_cells = selectOff
  //   }
  //   return true
  // }
  // cursor := group.Events[0].Key.Cursor()
  // if cursor != nil && cursor.Name() == "Mouse" {
  //   if w.selecting_cells == selectOn || w.selecting_cells == deselect {
  //     pos,_ := w.viewer.WindowToWall(cursor.Point())
  //     index := int(pos * float32(len(w.room.WallData)))
  //     if index >= 0 && index < len(w.room.WallData) {
  //       if w.selecting_cells == selectOn {
  //         w.selected_cells[index] = true
  //         w.selecting_cells = selectOn
  //       } else {
  //         delete(w.selected_cells, index)
  //         w.selecting_cells = deselect
  //       }
  //     }
  //   }
  // }
  return false
}

func (w *CellPanel) Think(ui *gui.Gui, t int64) {
  // if w.wall_texture != nil {
  //   mx,my := gin.In().GetCursor("Mouse").Point()
  //   pos,height := w.viewer.WindowToWall(mx, my)
  //   w.wall_texture.Pos = pos - w.drag_anchor.pos
  //   w.wall_texture.Height = height - w.drag_anchor.height
  // }
  // w.viewer.selected_cells = w.selected_cells
  // w.VerticalTable.Think(ui, t)

}

func (w *CellPanel) Collapse() {
  w.selecting_cells = selectOff
}

func (w *CellPanel) Expand() {

}
