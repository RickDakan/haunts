package actions

import (
  "encoding/gob"
  "path/filepath"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game/status"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
)

func registerSummonActions() map[string]func() game.Action {
  summons_actions := make(map[string]*SummonActionDef)
  base.RemoveRegistry("actions-summons_actions")
  base.RegisterRegistry("actions-summons_actions", summons_actions)
  base.RegisterAllObjectsInDir("actions-summons_actions", filepath.Join(base.GetDataDir(), "actions", "summons"), ".json", "json")
  makers := make(map[string]func() game.Action)
  for name := range summons_actions {
    cname := name
    makers[cname] = func() game.Action {
      a := SummonAction{ Defname: cname }
      base.GetObject("actions-summons_actions", &a)
      return &a
    }
  }
  return makers
}

func init() {
  game.RegisterActionMakers(registerSummonActions)
  gob.Register(&SummonAction{})
}

// Summon Actions target a single cell, are instant, and unreadyable.
type SummonAction struct {
  Defname string
  *SummonActionDef
  summonActionInst
}
type SummonActionDef struct {
  Name         string
  Kind         status.Kind
  Personal_los bool
  Ap           int
  Range        int
  Ent_name     string
  Animation    string
  Conditions   []string
  Texture      texture.Object
}
type summonActionInst struct {
  ent *game.Entity
  cx,cy int
  spawn *game.Entity
}
func (a *SummonAction) AP() int {
  return a.Ap
}
func (a *SummonAction) Pos() (int, int) {
  return 0, 0
}
func (a *SummonAction) Dims() (int, int) {
  return 0, 0
}
func (a *SummonAction) String() string {
  return a.Name
}
func (a *SummonAction) Icon() *texture.Object {
  return &a.Texture
}
func (a *SummonAction) Readyable() bool {
  return false
}
func (a *SummonAction) Preppable(ent *game.Entity, g *game.Game) bool {
  return ent.Stats.ApCur() >= a.Ap
}
func (a *SummonAction) Prep(ent *game.Entity, g *game.Game) bool {
  if !a.Preppable(ent, g) {
    return false
  }
  a.ent = ent
  return true
}
func (a *SummonAction) HandleInput(group gui.EventGroup, g *game.Game) game.InputStatus {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    bx,by := g.GetViewer().WindowToBoard(cursor.Point())
    bx += 0.5
    by += 0.5
    if bx < 0 { bx-- }
    if by < 0 { by-- }
    a.cx = int(bx)
    a.cy = int(by)
  }

  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if g.IsCellOccupied(a.cx, a.cy) {
      return game.Consumed
    }
    if a.Personal_los && !a.ent.HasLos(a.cx, a.cy, 1, 1) {
      return game.Consumed
    }
    if a.ent.Stats.ApCur() >= a.Ap {
      a.ent.Stats.ApplyDamage(-a.Ap, 0, status.Unspecified)
      a.spawn = game.MakeEntity(a.Ent_name, g)
      return game.ConsumedAndBegin
    }
    return game.Consumed
  }
  return game.NotConsumed
}
func (a *SummonAction) RenderOnFloor() {
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
  gl.Color4d(1.0, 0.2, 0.2, 0.8)
    gl.Vertex2i(a.cx + 0, a.cy + 0)
    gl.Vertex2i(a.cx + 0, a.cy + 1)
    gl.Vertex2i(a.cx + 1, a.cy + 1)
    gl.Vertex2i(a.cx + 1, a.cy + 0)
  gl.End()
}
func (a *SummonAction) Cancel() {
  a.summonActionInst = summonActionInst{}
}
func (a *SummonAction) Maintain(dt int64) game.MaintenanceStatus {
  if a.ent.Sprite().State() == "ready" {
    a.ent.TurnToFace(a.cx, a.cy)
    a.ent.Sprite().Command(a.Animation)
    a.spawn.Stats.OnBegin()
    a.ent.Game().SpawnEntity(a.spawn, a.cx, a.cy)
    return game.Complete
  }
  return game.InProgress
}
func (a *SummonAction) Interrupt() bool {
  return true
}

