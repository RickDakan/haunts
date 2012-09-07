package game

import (
  "fmt"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/mrgnet"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/opengl/gl"
  "path/filepath"
  "sort"
  "time"
)

type onlineLayout struct {
  Title struct {
    X, Y    int
    Texture texture.Object
  }
  Background texture.Object
  Back       Button

  User TextEntry

  Text struct {
    String        string
    Size          int
    Justification string
  }

  Games struct {
    Up     Button
    Down   Button
    Scroll ScrollingRegion
  }
}

type OnlineMenu struct {
  layout  onlineLayout
  region  gui.Region
  buttons []ButtonLike
  games   []ButtonLike
  mx, my  int
  last_t  int64

  update_user  chan mrgnet.UpdateUserResponse
  update_alpha float64
  update_time  time.Time

  list_games chan mrgnet.ListGamesResponse
  list_time  time.Time
}

func InsertOnlineMenu(ui gui.WidgetParent) error {
  var sm OnlineMenu
  datadir := base.GetDataDir()
  err := base.LoadAndProcessObject(filepath.Join(datadir, "ui", "start", "online", "layout.json"), "json", &sm.layout)
  if err != nil {
    return err
  }
  sm.buttons = []ButtonLike{
    &sm.layout.Back,
    &sm.layout.Games.Up,
    &sm.layout.Games.Down,
    &sm.layout.User,
  }
  sm.layout.Back.f = func(interface{}) {
    ui.RemoveChild(&sm)
    InsertStartMenu(ui)
  }
  sm.layout.Games.Up.f = func(interface{}) {
    sm.layout.Games.Scroll.Up()
  }
  sm.layout.Games.Down.f = func(interface{}) {
    sm.layout.Games.Scroll.Down()
  }

  var net_id mrgnet.NetId
  fmt.Sscanf(base.GetStoreVal("netid"), "%d", &net_id)
  if net_id == 0 {
    net_id = mrgnet.NetId(mrgnet.RandomId())
    base.SetStoreVal("netid", fmt.Sprintf("%d", net_id))
  }

  sm.update_user = make(chan mrgnet.UpdateUserResponse)
  sm.layout.User.Button.f = func(interface{}) {
    var req mrgnet.UpdateUserRequest
    req.Name = sm.layout.User.Entry.text
    req.Id = net_id
    var resp mrgnet.UpdateUserResponse
    go func() {
      mrgnet.DoAction("user", req, &resp)
      sm.update_user <- resp
    }()
  }
  go func() {
    var resp mrgnet.UpdateUserResponse
    mrgnet.DoAction("user", mrgnet.UpdateUserRequest{Id: net_id}, &resp)
    sm.update_user <- resp
  }()
  // sm.layout.User.f = func(interface{}) {

  sm.list_games = make(chan mrgnet.ListGamesResponse)
  go func() {
    var resp mrgnet.ListGamesResponse
    mrgnet.DoAction("list", mrgnet.ListGamesRequest{Id: net_id, Unstarted: true}, &resp)
    sm.list_games <- resp
  }()

  ui.AddChild(&sm)
  return nil
}

func (sm *OnlineMenu) Requested() gui.Dims {
  return gui.Dims{1024, 768}
}

func (sm *OnlineMenu) Expandable() (bool, bool) {
  return false, false
}

func (sm *OnlineMenu) Rendered() gui.Region {
  return sm.region
}

type onlineButtonSlice []ButtonLike

func (o onlineButtonSlice) Len() int      { return len(o) }
func (o onlineButtonSlice) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o onlineButtonSlice) Less(i, j int) bool {
  return o[i].(*Button).Text.String < o[j].(*Button).Text.String
}

func (sm *OnlineMenu) Think(g *gui.Gui, t int64) {
  if sm.last_t == 0 {
    sm.last_t = t
    return
  }
  dt := t - sm.last_t
  sm.last_t = t
  sm.layout.Games.Scroll.Think(dt)
  if sm.mx == 0 && sm.my == 0 {
    sm.mx, sm.my = gin.In().GetCursor("Mouse").Point()
  }

  for {
    select {
    case resp := <-sm.update_user:
      sm.layout.User.Entry.text = resp.Name
      sm.update_alpha = 1.0
      sm.update_time = time.Now()
    case list := <-sm.list_games:
      if list.Err != "" {
        base.Error().Printf("Error getting listing: %v", list.Err)
        break
      }
      sm.games = sm.games[0:0]
      for i := range list.Games {
        base.Log().Printf("LIST: %v", list.Games[i].Name)
        var b Button
        b.Text.Justification = sm.layout.Text.Justification
        b.Text.Size = sm.layout.Text.Size
        b.Text.String = list.Games[i].Name
        var net_id mrgnet.NetId
        fmt.Sscanf(base.GetStoreVal("netid"), "%d", &net_id)
        b.f = func(interface{}) {
          var req mrgnet.JoinGameRequest
          req.Id = net_id
          req.Game_key = list.Ids[i]
          var resp mrgnet.JoinGameResponse
          mrgnet.DoAction("join", req, &resp)
        }
        sm.games = append(sm.games, &b)
        sort.Sort(onlineButtonSlice(sm.games))
        sm.layout.Games.Scroll.Height = int(base.GetDictionary(sm.layout.Text.Size).MaxHeight() * float64(len(sm.games)))
      }
    default:
      goto nomore
    }
  }
nomore:

  if sm.update_alpha > 0.0 && time.Now().Sub(sm.update_time).Seconds() >= 2 {
    sm.update_alpha = doApproach(sm.update_alpha, 0.0, dt)
  }

  for _, button := range sm.buttons {
    button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, dt)
  }

  if (gui.Point{sm.mx, sm.my}.Inside(sm.layout.Games.Scroll.Region())) {
    for _, button := range sm.games {
      button.Think(sm.region.X, sm.region.Y, sm.mx, sm.my, dt)
    }
  } else {
    for _, button := range sm.games {
      button.Think(sm.region.X, sm.region.Y, 0, 0, dt)
    }
  }
}

func (sm *OnlineMenu) Respond(g *gui.Gui, group gui.EventGroup) bool {
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    sm.mx, sm.my = cursor.Point()
  }
  if found, event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    for _, button := range sm.buttons {
      if button.handleClick(sm.mx, sm.my, nil) {
        return true
      }
    }
    for _, button := range sm.games {
      // TODO: Check for scroll bounds first
      if button.handleClick(sm.mx, sm.my, nil) {
        return true
      }
    }
  }
  hit := false
  for _, button := range sm.buttons {
    if button.Respond(group, nil) {
      hit = true
    }
  }
  inside := gui.Point{sm.mx, sm.my}.Inside(sm.layout.Games.Scroll.Region())
  if cursor == nil || inside {
    for _, button := range sm.games {
      // TODO: Check for scroll bounds first
      if button.Respond(group, nil) {
        hit = true
      }
    }
  }
  if hit {
    return true
  }
  return false
}

func (sm *OnlineMenu) Draw(region gui.Region) {
  sm.region = region
  gl.Color4ub(255, 255, 255, 255)
  sm.layout.Background.Data().RenderNatural(region.X, region.Y)
  title := sm.layout.Title
  title.Texture.Data().RenderNatural(region.X+title.X, region.Y+title.Y)
  for _, button := range sm.buttons {
    button.RenderAt(sm.region.X, sm.region.Y)
  }

  sm.layout.Games.Scroll.Region().PushClipPlanes()
  d := base.GetDictionary(sm.layout.Text.Size)
  sx := sm.layout.Games.Scroll.X
  sy := sm.layout.Games.Scroll.Top() - int(d.MaxHeight())
  for _, button := range sm.games {
    base.Log().Printf("Rendering at %d %d", sx, sy)
    button.RenderAt(sx, sy)
    sy -= int(d.MaxHeight())
  }
  sm.layout.Games.Scroll.Region().PopClipPlanes()

  gl.Color4ub(255, 255, 255, byte(255*sm.update_alpha))
  sx = sm.layout.User.Entry.X + sm.layout.User.Entry.Dx + 10
  sy = sm.layout.User.Button.Y
  d.RenderString("Name Updated", float64(sx), float64(sy), 0, d.MaxHeight(), gui.Left)
}

func (sm *OnlineMenu) DrawFocused(region gui.Region) {
}

func (sm *OnlineMenu) String() string {
  return "online menu"
}
