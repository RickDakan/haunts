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
  Script.LoadHouse("soundstage")  

  store.side = side_choices[1]
  -- store.side = "Humans"
  if store.side == "Humans" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end
  if store.side == "Denizens" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "ch01/intruders.lua")
  end
  if store.side == "Intruders" then
    Script.BindAi("denizen", "ch01/denizens.lua")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end

    denizen_spawn = Script.GetSpawnPointsMatching("denizen_spawn")
    intruder_spawn = Script.GetSpawnPointsMatching("intruder_spawn")

    Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Corpse", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Chosen One", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Eidolon", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Devotee", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Disciple", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Master of the Manse", denizen_spawn, false)
    --Script.SpawnEntitySomewhereInSpawnPoints("Transdimensional Maw", denizen_spawn, false)
    --Script.SpawnEntitySomewhereInSpawnPoints("Genius", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Golem", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Cultist", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Lost Soul", denizen_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Vengeful Wraith", denizen_spawn, false)

    Script.SpawnEntitySomewhereInSpawnPoints("Collector", intruder_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Detective", intruder_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Ghost Hunter", intruder_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Occultist", intruder_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Reporter", intruder_spawn, false)
    Script.SpawnEntitySomewhereInSpawnPoints("Teen", intruder_spawn, false)
end

function RoundStart(intruders, round)
  if round == 1 then
    Script.EndPlayerInteraction()
  end

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

  bSelected = false
  for _, ent in pairs(Script.GetAllEnts()) do
    Script.SetAp(ent, 100)
    if ent.Side.Intruder == intruders and not bSelected then
      Script.SelectEnt(ent)
      bSelected = true
    end
  end

  side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}
end

function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec
end


function RoundEnd(intruders, round)
  if round == 1 then
    return
  end

  if store.side == "Humans" then
    Script.ShowMainBar(false)
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
    if intruders then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end

    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.LoadGameState(store.game)

    for _, exec in pairs(store.execs) do
      bDone = false
      if not bDone then
        Script.DoExec(exec)
      end
    end
    store.execs = {}
  end
end