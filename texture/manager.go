package texture

import (
  "image"
  "image/draw"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "runtime"
  "sync"
  "github.com/runningwild/glop/render"
  "github.com/runningwild/glop/memory"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/opengl/glu"
  "github.com/runningwild/mathgl"
)

type Object struct {
  Path base.Path

  // this path is the last one that was loaded, so that if we change Path then
  // we know to reload the texture.
  path base.Path
  data *Data
}
func (o *Object) Data() *Data {
  if o.data == nil || o.path != o.Path {
    o.data = LoadFromPath(string(o.Path))
    o.path = o.Path
  }
  return o.data
}

type Data struct {
  dx,dy int
  texture gl.Texture
}
func (d *Data) Dx() int {
  return d.dx
}
func (d *Data) Dy() int {
  return d.dy
}

var textureList uint
var textureListSync sync.Once
func setupTextureList() {
  textureListSync.Do(func() {
    render.Queue(func() {
      textureList = gl.GenLists(1)
      gl.NewList(textureList, gl.COMPILE)
      gl.Begin(gl.QUADS)
        gl.TexCoord2d(0, 0)
        gl.Vertex2i(0, 0)

        gl.TexCoord2d(0, -1)
        gl.Vertex2i(0, 1)

        gl.TexCoord2d(1, -1)
        gl.Vertex2i(1, 1)

        gl.TexCoord2d(1, 0)
        gl.Vertex2i(1, 0)
      gl.End()
      gl.EndList()
    })
  })
}
// Renders the texture on a quad at the texture's natural size.
func (d *Data) RenderNatural(x, y int) {
  d.Render(float64(x), float64(y), float64(d.dx), float64(d.dy))
}

func (d *Data) Render(x, y, dx, dy float64) {
  if textureList != 0 {
    var run, op mathgl.Mat4
    run.Identity()
    op.Translation(float32(x), float32(y), 0)
    run.Multiply(&op)
    op.Scaling(float32(dx), float32(dy), 1)
    run.Multiply(&op)

    gl.PushMatrix()
    gl.Enable(gl.TEXTURE_2D)
    d.Bind()
    gl.MultMatrixf(&run[0])
    gl.CallList(textureList)
    gl.PopMatrix()
  }
}

func (d *Data) RenderAdvanced(x, y, dx, dy, rot float64, flip bool) {
  if textureList != 0 {
    var run, op mathgl.Mat4
    run.Identity()
    op.Translation(float32(x), float32(y), 0)
    run.Multiply(&op)
    op.Translation(float32(dx/2), float32(dy/2), 0)
    run.Multiply(&op)
    op.RotationZ(float32(rot))
    run.Multiply(&op)
    if flip {
      op.Translation(float32(-dx/2), float32(-dy/2), 0)
      run.Multiply(&op)
      op.Scaling(float32(dx), float32(dy), 1)
      run.Multiply(&op)
    } else {
      op.Translation(float32(dx/2), float32(-dy/2), 0)
      run.Multiply(&op)
      op.Scaling(float32(-dx), float32(dy), 1)
      run.Multiply(&op)
    }
    gl.PushMatrix()
    gl.MultMatrixf(&run[0])
    gl.Enable(gl.TEXTURE_2D)
    d.Bind()
    gl.CallList(textureList)
    gl.PopMatrix()
  }
}

func (d *Data) Bind() {
  if d.texture == 0 {
    if error_texture == 0 {
      gl.Enable(gl.TEXTURE_2D)
      error_texture = gl.GenTexture()
      error_texture.Bind(gl.TEXTURE_2D)
      gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
      gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
      pink := []byte{ 255, 0, 255, 255 }
      glu.Build2DMipmaps(gl.TEXTURE_2D, 4, 1, 1, gl.RGBA, pink)
    }
    error_texture.Bind(gl.TEXTURE_2D)
  } else {
    d.texture.Bind(gl.TEXTURE_2D)
  }
}

var (
  manager Manager
  error_texture gl.Texture
)

func init() {
  manager.registry = make(map[string]*Data)
}

type Manager struct {
  registry map[string]*Data
}

func Reload() {
  manager.Reload()
}
func (m *Manager) Reload() {
}

func LoadFromPath(path string) *Data {
  return manager.LoadFromPath(path)
}

func finalizeData(d *Data) {
  render.Queue(func() {
    d.texture.Delete()
  })
}

type loadRequest struct {
  path string
  data *Data
}
var load_requests chan loadRequest
var load_count int
var load_mutex sync.Mutex
const load_threshold = 1000*1000
func init() {
  load_requests = make(chan loadRequest, 100)
  for i := 0; i < 4; i++ {
    go loadTextureRoutine()
  }
}

// This routine waits for a filename and a data object, then loads the texture
// in that file into that object.  This is so that only one texture is being
// loaded at a time, it prevents us from hammering the filesystem and also
// makes sure we aren't using up a ton of memory all at once.
func loadTextureRoutine() {
  for req := range load_requests {
    handleLoadRequest(req)
  }
}

func handleLoadRequest(req loadRequest) {
  f,_ := os.Open(req.path)
  im,_,err := image.Decode(f)
  f.Close()
  if err != nil {
    base.Warn().Printf("Unable to decode texture '%s': %v", req.path, err)
    return
  }
  gray := true
  dx := im.Bounds().Dx()
  dy := im.Bounds().Dy()
  for i := 0; i < dx; i++ {
    for j := 0; j < dy; j++ {
      r, g, b, _ := im.At(i, j).RGBA()
      if r != g || g != b {
        gray = false
        break
      }
    }
    if !gray {
      break
    }
  }
  var canvas draw.Image
  var pix []byte
  if gray {
    ga := NewGrayAlpha(im.Bounds())
    pix = ga.Pix
    canvas = ga
  } else {
    pix = memory.GetBlock(4*req.data.dx*req.data.dy)
    canvas = &image.RGBA{pix, 4*req.data.dx, im.Bounds()}
  }
  draw.Draw(canvas, im.Bounds(), im, image.Point{}, draw.Src)
  load_mutex.Lock()
  load_count += len(pix)
  manual_unlock := false
  // This prevents us from trying to send too much to opengl in a single
  // frame.  If we go over the threshold then we hold the lock until we're
  // done sending data to opengl, then other requests will be free to
  // queue up and they will run on the next frame.
  if load_count < load_threshold {
    load_mutex.Unlock()
  } else {
    manual_unlock = true
  }
  render.Queue(func() {
    gl.Enable(gl.TEXTURE_2D)
    req.data.texture = gl.GenTexture()
    req.data.texture.Bind(gl.TEXTURE_2D)
    gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    if gray {
      glu.Build2DMipmaps(gl.TEXTURE_2D, gl.LUMINANCE_ALPHA, req.data.dx, req.data.dy, gl.LUMINANCE_ALPHA, pix)
    } else {
      glu.Build2DMipmaps(gl.TEXTURE_2D, gl.RGBA, req.data.dx, req.data.dy, gl.RGBA, pix)
    }
    memory.FreeBlock(pix)
    if manual_unlock {
      load_count = 0
      load_mutex.Unlock()
    }
    runtime.SetFinalizer(req.data, finalizeData)
  })
}

func (m *Manager) LoadFromPath(path string) *Data {
  setupTextureList()
  if data,ok := m.registry[path]; ok {
    return data
  }
  base.Log().Printf("Loading %s\n", path)
  var data Data
  m.registry[path] = &data

  f,err := os.Open(path)
  if err != nil {
    base.Warn().Printf("Unable to load texture '%s': %v", path, err)
    return &data
  }
  config,_,err := image.DecodeConfig(f)
  f.Close()
  data.dx = config.Width
  data.dy = config.Height

  load_requests <- loadRequest{path, &data}
  return &data
}
