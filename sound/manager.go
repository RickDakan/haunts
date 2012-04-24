package sound

import (
  "path/filepath"
  "github.com/runningwild/glop/sprite"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/fmod"
)

var(
  system *fmod.System
  music *fmod.Channel
  names map[string]string  // 'load' -> 'load_gun.ogg', stuff like that
)

func Init() {
  var err error
  system, err = fmod.CreateSystem()
  if err != nil {
    base.Error().Printf("Unable to create sound system: %v", err)
    return
  }
  err = system.Init(32, fmod.INIT_NORMAL, nil)
  if err != nil {
    base.Error().Printf("Unable to initialize sound system: %v", err)
    return
  }
  version, _ := system.GetVersion()
  err = base.LoadJson(filepath.Join(base.GetDataDir(), "sound", "names.json"), &names)
  if err != nil {
    base.Error().Printf("Unable to load names.json: %v", err)
    return
  }
  base.Log().Printf("Fmod version %x", version)
  sprite.SetTriggerFunc(trigger)
}

func MapSounds(m map[string]string) {
  for k,v := range m {
    names[k] = v
  }
}

func trigger(s *sprite.Sprite, name string) {
  base.Log().Printf("trigger: %p - %s", s, name)
  file, ok := names[name]
  if !ok {
    base.Error().Printf("Attempted to play unknown sound '%s'", name)
    return
  }
  path := filepath.Join(base.GetDataDir(), "sound", file)
  sound, err := system.CreateSound_FromFilename(path, fmod.MODE_LOOP_OFF)
  if err != nil {
    base.Error().Printf("Unable to load %s: %v", file, err)
    return
  }

  music, err = system.PlaySound(fmod.CHANNEL_FREE, sound, false)
  if err != nil {
    base.Error().Printf("Unable to play %s: %v", file, err)
    return
  }
  base.Log().Printf("Played sound: %s\n", file)
}

func SetBackgroundMusic(file string) {
  if file == "" {
    if music != nil {
      music.Stop()
      music = nil
    }
    return
  }
  path := filepath.Join(base.GetDataDir(), "sound", "music", file)
  sound, err := system.CreateSound_FromFilename(path, fmod.MODE_LOOP_NORMAL)
  if err != nil {
    base.Error().Printf("Unable to load %s: %v", file, err)
    return
  }

  music, err = system.PlaySound(fmod.CHANNEL_FREE, sound, false)
  if err != nil {
    base.Error().Printf("Unable to play %s: %v", file, err)
    return
  }
  cg, err := system.GetMasterChannelGroup()
  if err != nil {
    base.Error().Printf("Unable to set volume: %v", err)
    return
  }
  cg.SetVolume(0.1)
}




