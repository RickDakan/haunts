package house

import (
  "runtime"
  "github.com/runningwild/glop/render"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/opengl/glu"
)

const LosMinVisibility = 64
const LosVisibilityThreshold = 200

// A LosTexture is defined over a square portion of a grid, and if a pixel is
// non-black it indicates that there is visibility to that pixel from the
// center of the texture.  The texture is a square with a size that is a power
// of two, so the center is defined as the pixel to the lower-left of the
// actual center of the texture.
type LosTexture struct {
  pix []byte
  p2d [][]byte
  tex gl.Texture
  x,y int

  // The texture needs to be created on the render thread, so we use this to
  // get the texture after it's been made.
  rec chan gl.Texture
}

func losTextureFinalize(lt *LosTexture) {
  render.Queue(func() {
    gl.Enable(gl.TEXTURE_2D)
    lt.tex.Delete()
  })
}

// Creates a LosTexture with the specified size, which must be a power of two.
func MakeLosTexture(size int) *LosTexture {
  var lt LosTexture
  lt.pix = make([]byte, size*size)
  lt.p2d = make([][]byte, size)
  lt.rec = make(chan gl.Texture, 1)
  for i := 0; i < size; i++ {
    lt.p2d[i] = lt.pix[i * size : (i+1) * size]
  }

  render.Queue(func() {
    gl.Enable(gl.TEXTURE_2D)
    tex := gl.GenTexture()
    tex.Bind(gl.TEXTURE_2D)
    gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    glu.Build2DMipmaps(gl.TEXTURE_2D, gl.RGBA, len(lt.p2d), len(lt.p2d), gl.ALPHA, lt.pix)
    lt.rec <- tex
  })

  runtime.SetFinalizer(&lt, losTextureFinalize)

  return &lt
}

// If the texture has been created this returns true, otherwise it checks for
// the finished texture and returns true if it is available, false otherwise.
func (lt *LosTexture) ready() bool {
  if lt.tex != 0 { return true }
  select {
    case lt.tex = <-lt.rec:
      return true
    default:
  }
  return false
}

// Updates OpenGl with any changes that have been made to the texture.  The
// coordinates given specify the new position of the center of the texture.
// OpenGl calls in this method are run on the render thread
func (lt *LosTexture) Remap(x,y int) {
  lt.x = x
  lt.y = y
  if !lt.ready() { return }
  render.Queue(func() {
    gl.Enable(gl.TEXTURE_2D)
    lt.tex.Bind(gl.TEXTURE_2D)
    glu.Build2DMipmaps(gl.TEXTURE_2D, gl.RGBA, len(lt.p2d), len(lt.p2d), gl.ALPHA, lt.pix)
  })
}

// Binds the texture, not run on the render thread
func (lt *LosTexture) Bind() {
  lt.ready()
  lt.tex.Bind(gl.TEXTURE_2D)
}

func (lt *LosTexture) Size() int {
  return len(lt.p2d)
}

// Clears the texture so that all pixels are set to the specified value
func (lt *LosTexture) Clear(v byte) {
  for i := range lt.pix {
    lt.pix[i] = v
  }
}

// Returns the coordinates of the minimum and maximum corners of the region
// that this texture covers.
func (lt *LosTexture) Region() (x,y,x2,y2 int) {
  x = lt.x - len(lt.p2d) / 2 + 1
  y = lt.y - len(lt.p2d) / 2 + 1
  x2 = lt.x + len(lt.p2d) / 2
  y2 = lt.y + len(lt.p2d) / 2
  return
}

// Gets the value at the specified texel, taking into account the offset
// that the texture is currently positioned at.
func (lt *LosTexture) Get(x,y int) byte {
  return lt.p2d[y - lt.y + len(lt.p2d) / 2 - 1][x - lt.x + len(lt.p2d) / 2 - 1]
}

// Sets the texel at x,y to val, taking into account the offset that the
// texture is currently positioned at.
func (lt *LosTexture) Set(x,y int, val byte) {
  lt.p2d[y - lt.y + len(lt.p2d) / 2 - 1][x - lt.x + len(lt.p2d) / 2 - 1] = val
}

// Returns a convenient 2d slice over the texture, as well as the coordinates
// of the lower-left pixel.
func (lt *LosTexture) Pix() (pix [][]byte, x,y int) {
  return lt.p2d, lt.x, lt.y
}
