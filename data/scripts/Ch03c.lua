function Init()
   store.Ch03c = {}
   store.Ch03c.Spawnpoints_complete={}
   store.Ch03c.Spawnpoints = {
      "Ch03_Part3_Intro",
      "Ch03_Patient_Two_Dies",
      "Ch03_Patient_Three_Dies",
      "Ch03_Patient_Four_Dies"
   } 

  Script.LoadHouse("Chapter_03_c")
  Script.DialogBox("ui/dialog/Ch03/Ch03_Part3_Intro.json")
  
  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
    --always bind one to human!
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Wilson Sax", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Director Tyree", intruder_spawn)
  
  maw_spawn = Script.GetSpawnPointsMatching("Maw-Start")
  for i = 1,4 do
    Script.SpawnEntitySomewhereInSpawnPoints("Transdimensional Maw", maw_spawn)
  end
  
  cultists_spawn = Script.GetSpawnPointsMatching("Cultists-Start")
  for i = 1,4 do
    Script.SpawnEntitySomewhereInSpawnPoints("Cultist", cultists_spawn)
  end


  wraith_spawn = Script.GetSpawnPointsMatching("Wraith-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Vengeful Wraith", maw_spawn)
  end
  
  shades_spawn = Script.GetSpawnPointsMatching("Shades-Start")
  for i = 1,4 do
    Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", cultists_spawn)
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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch03c.Spawnpoints, store.Ch03c.Spawnpoints_complete)
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
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch03c.Spawnpoints, store.Ch03c.Spawnpoints_complete)
end


function RoundEnd(intruders, round)
  cultists_dead = true
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Cultist" and ent.Stats.HpCur > 0 then
      cultists_dead = false
    end
  end

  maws_dead = true
  for _, ent in pairs (Script.GetAllEnts()) do
    if ent.Name == "Transdimensional Maw" and ent.Stats.HpCur > 0 then
      cultists_dead = false
    end
  end

  wraiths_dead = true
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Vengeful Wraith" and ent.Stats.HpCur > 0 then
      cultists_dead = false
    end
  end

  shades_dead = true
  for _, ent in pairs (Script.GetAllEnts()) do
    if ent.Name == "Angry Shade" and ent.Stats.HpCur > 0 then
      cultists_dead = false
    end
  end

  if cultists_dead == true and maws_dead == true and wraiths_dead == true and shades_dead == true then
    Script.DialogBox("Ch03_Tyree_Congrats.json")
    store.Ch03c.choice_a = choices[1]  
    Script.StartScript("Ch03d.lua")
  end
end
