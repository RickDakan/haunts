function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

if not store.Ch03a then
  store.Ch03a = {}
  end
--
play_as_denizens = false
function Init()
   store.Ch03a = {}
   store.Ch03a.Spawnpoints_complete={}
   store.Ch03a.Spawnpoints = {
      "Bedroom-01",
      "Bedroom-02",
      "Spawn_Patients",
      "Trigger_Patients",
      "Armed_And_New_Map",
      "Vanish_Patient03",
      "Patient04-Spawn",
   } 

  Script.LoadHouse("Chapter_03_a")
  Script.DialogBox("ui/dialog/Ch03/Ch03_Intro.json") 

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
    --always bind one to human!
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Wilson Sax", intruder_spawn)
  
  patient01_spawn = Script.GetSpawnPointsMatching("Patient01-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patient01_spawn)
  
  patient02_spawn = Script.GetSpawnPointsMatching("Patient02-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patient02_spawn)

  patient03_spawn = Script.GetSpawnPointsMatching("Patient03-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patient03_spawn)
  ents = Script.GetAllEnts()
end
 

function RoundStart(intruders, round)
    Script.SetVisibility("intruders")
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.ShowMainBar(intruders ~= play_as_denizens)
end


function pointIsInSpawn(pos, sp)
  return pos.X >= sp.Pos.X and pos.X < sp.Pos.X + sp.Dims.Dx and pos.Y >= sp.Pos.Y and pos.Y < sp.Pos.Y + sp.Dims.Dy
end

function pointIsInSpawns(pos, regexp)
  sps = Script.GetSpawnPointsMatching(regexp)
  for _, sp in pairs(sps) do
    if pointIsInSpawn(pos, sp) then
      return true
    end
  end
  return false
end


function IsPosInUnusedSpawnpoint(pos, list, used)
  --name identifies spawnpoint
  for _, spawn in pairs(list) do
    if not used[spawn] and pointIsInSpawns(pos, spawn) then
      return spawn
    end
  end
  return nil
end


--THIS STOPS a unit in a spawn point not yet activated.
function OnMove(ent, path)
  if not ent.Side.Intruder then
    return table.getn(path)
  end
  for i, pos in pairs(path) do
    name = IsPosInUnusedSpawnpoint(pos, store.Ch03a.Spawnpoints, store.Ch03a.Spawnpoints_complete)
    if name then
      return i
    end
  end
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  if not exec.Ent.Side.Intruder then
    return
  end
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch03a.Spawnpoints, store.Ch03a.Spawnpoints_complete)
  if name == "Bedroom-01" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_Patient01_bedroom.json")
    store.Ch03a.Spawnpoints_complete["Bedroom-01"] = true
  end

  if name == "Bedroom-02" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_Patient02_bedroom.json")
    store.Ch03a.Spawnpoints_complete["Bedroom-01"] = true
  end

  if name == "Vanish_Patient03" then
    for _, ent in pairs(Script.GetAllEnts()) do
        if ent.Name == "Patient" then
          Script.PlayAnimations(ent, {"defend", "killed"})
          Script.SetHp(ent, 0)
          end
        end
    end
    store.Ch03a.Spawnpoints_complete["Vanish_Patient03"] = true
  end

  if name == "Patient04-Spawn" then
    patient04_spawn = Script.GetSpawnPointsMatching("Patient04-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Patient", patient04_spawn)
    store.Ch03a.Spawnpoints_complete["Patient04-Spawn"] = true
  end

  if name == "Trigger_Patients" then
    triggered_patients = Script.GetSpawnPointsMatching("Spawn_Patients")
    store.Ch03a.Spawnpoints_complete["Trigger_Patients"] = true
    for i = 1,4 do
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Patient", triggered_patients)
    end
  end

  if name == "Armed_And_New_Map" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_rec_room_arming.json")
    Script.StartScript("Ch03b.lua")
  end

end


function RoundEnd(intruders, round)
  print("end", intruders, round)
end

