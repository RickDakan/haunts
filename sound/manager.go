//

// +build !nosound
package sound

import (
  "path/filepath"
  "github.com/runningwild/haunts/base"
  fmod "github.com/runningwild/fmod/event"
)

var (
  system *fmod.System
)

func Init() {
  var err error
  system, err = fmod.EventSystemCreate()
  if err != nil {
    base.Error().Printf("Unable to create sound system: %v", err)
    return
  }

  err = system.Init(32, fmod.INIT_NORMAL, nil, fmod.EVENT_INIT_NORMAL)
  if err != nil {
    base.Error().Printf("Unable to initialize sound system: %v", err)
    return
  }
  version, _ := system.GetVersion()
  base.Log().Printf("Fmod version %x", version)

  err = system.SetMediaPath(filepath.Join(base.GetDataDir(), "sound") + "/")
  if err != nil {
    base.Error().Printf("Unable to set media path: %v\n", err)
    return
  }

  err = system.LoadPath("Haunts.fev", nil)
  if err != nil {
    base.Error().Printf("Unable to load fev: %v\n", err)
    return
  }
}

func PlaySound(name string) {
  if system == nil {
    return
  }
  sound, err := system.GetEvent(name, fmod.MODE_DEFAULT)
  if err != nil {
    base.Error().Printf("Unable to get event '%s': %v", name, err)
    return
  }
  sound.Start()
}
