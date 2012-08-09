function Init()
   store.Ch03b = {}
   store.Ch03b.Spawnpoints_complete={}
   store.Ch03b.Spawnpoints = {
      "Ch03_Patient_One_Dies",
      "Ch03_Patient_Two_Dies",
      "Ch03_Patient_Three_Dies",
      "Ch03_Patient_Four_Dies"
   } 

  Script.DialogBox("ui/dialog/Ch03/Ch03_Fight_Intro.json")
  Script.LoadHouse("Chapter_03_b")

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
    --always bind one to human!
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Lt. Wilson Sax", intruder_spawn)
  
  patients_spawn = Script.GetSpawnPointsMatching("Patients-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Patient One", patients_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Patient Two", patients_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Patient Three", patients_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Patient Four", patients_spawn)

  maw_spawn = Script.GetSpawnPointsMatching("Maw-Start")
  for i = 1,4 do
    Script.SpawnEntitySomewhereInSpawnPoints("Transdimensional Maw", maw_spawn)
  end
  
  cultists_spawn = Script.GetSpawnPointsMatching("Cultists-Start")
  for i = 1,4 do
    Script.SpawnEntitySomewhereInSpawnPoints("Cultist", cultists_spawn)
  end

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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch03b.Spawnpoints, store.Ch03b.Spawnpoints_complete)
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
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch03b.Spawnpoints, store.Ch03b.Spawnpoints_complete)
  

  for _, ent in pairs (Script.GetAllEnts()) do
    if ent.Name("Patient One") and ent.HpCur == 0 and not store.Ch03b.Spawnpoints_complete["Ch03_Patient_One_Dies"] then
      store.Ch03b.Spawnpoints_complete["Ch03_Patient_One_Dies"] = true
 --     Script.FocusPos(ent.Pos)
      Script.DialogBox("Ch03_Patient_One_Dies.json")
    end
  end
 
    if ent.Name("Patient Two") and ent.HpCur == 0 and not store.Ch03b.Spawnpoints_complete["Ch03_Patient_Two_Dies"] then
      store.Ch03b.Spawnpoints_complete["Ch03_Patient_Two_Dies"] = true
--      Script.FocusPos(ent.Pos)
      Script.DialogBox("Ch03_Patient_Two_Dies.json")
    end

    if ent.Name("Patient Three") and ent.HpCur == 0 then and not store.Ch03b.Spawnpoints_complete["Ch03_Patient_Three_Dies"]
      store.Ch03b.Spawnpoints_complete["Ch03_Patient_Three_Dies"] = true
 --     Script.FocusPos(ent.Pos)
      Script.DialogBox("Ch03_Patient_Three_Dies.json")
    end

    if ent.Name("Patient Four") and ent.HpCur == 0 and not store.Ch03b.Spawnpoints_complete["Ch03_Patient_Four_Dies"] then
      store.Ch03b.Spawnpoints_complete["Ch03_Patient_Four_Dies"] = true
 --     Script.FocusPos(ent.Pos)
      Script.DialogBox("Ch03_Patient_Four_Dies.json")
    end
  end

-- -- SPAWN Tyree and have him WALK to PC
-- -- IF he gets near, trigger something
end

function RoundEnd(intruders, round)
  cultists_dead = true
  maws_dead = true
 
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Cultist" and ent.HpCur > 0 then
      cultists_dead = false
    end
  end

  for _, ent in pairs (Script.GetAllEnts()) do
    if ent.Name == "Transdimensional Maw" and ent.HpCur > 0 then
      maws_dead = false
    end
  end

  if cultists_dead == true and maws_dead == true then
    Script.DialogBox("Ch03_Doc_Tyree_Arrives.json")
    Script.StartScript("Ch03c.lua")
  end
end
