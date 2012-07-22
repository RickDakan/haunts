function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function Init(data)
  side_choices = Script.ChooserFromFile("ui/start/versus/side.json")

  -- check data.map == "random" or something else
  Script.LoadHouse("versus-1")

  store.side = side_choices[1]
  if store.side == "Humans" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end
  if store.side == "Denizens" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "intruders.lua")
  end
  if store.side == "Intruders" then
    Script.BindAi("denizen", "denizens.lua")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end

  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")
  Script.SpawnEntitySomewhereInSpawnPoints("Teen", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Occultist", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Ghost Hunter", intruder_spawn)

  -- Temporary - just for testing:
  spawn = Script.GetSpawnPointsMatching("Master-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Chosen One", spawn)
  spawn = Script.GetSpawnPointsMatching("Servitors-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Corpse", spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Corpse", spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Corpse", spawn)
  -- End Temporary

  Script.SetLosMode("intruders", "entities")
  Script.SetLosMode("denizens", "entities")
  if store.side == "Intruders" then
    Script.SetVisibility("intruders")
  end
  if store.side == "Denizens" then
    Script.SetVisibility("denizens")
  end
  Script.ShowMainBar(true)
end

function RoundStart(intruders, round)
  if store.side == "Humans" then
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("intruders")
    else
      Script.SetVisibility("denizens")
    end
    Script.ShowMainBar(true)
  else
    Script.ShowMainBar(intruders == (store.side == "Intruders"))
  end
end


function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
end
 

function RoundEnd(intruders, round)
  if store.side == "Humans" then
    Script.ShowMainBar(false)
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
    if intruders then
      Script.DialogBox("ui/start/versus/pass_to_denizens.json")
    else
      Script.DialogBox("ui/start/versus/pass_to_intruders.json")
    end
  end
end

