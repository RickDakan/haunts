
function Init()
  if not store.Ch05a then
    store.Ch05a = {}
  end
  store.Ch05a.Spawnpoints_complete={}
  store.Ch05a.Spawnpoints = {
    "Foyer",
    "Creepy1",
    "Creepy2",
    "Study Desk",
    "Private Entrance",
    "Creepy3",
    "Lab",
    "Coffin",
  } 


  Script.LoadHouse("Chapter_05_a")
    
  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
 
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  print("Found", table.getn(intruder_spawn), "points")
  Script.SpawnEntitySomewhereInSpawnPoints("Timothy K.", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Constance M.", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Samuel T.", intruder_spawn)
-- "ui/dialog/Ch05/ch05_Intro.json"
  Script.DialogBox("ui/dialog/Ch05/Ch05_Intro.json")
end
 

function RoundStart(intruders, round)
  Script.SetVisibility("intruders")
  Script.SetLosMode("intruders", "entities")
  Script.SetLosMode("denizens", "entities")
  Script.ShowMainBar(intruders)
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
  for _, spawn in pairs(list) do
    if not used[spawn] and pointIsInSpawns(pos, spawn) then
      return spawn
    end
  end
  return nil
end

function OnMove(ent, path)
  if not ent.Side.Intruder then
    return table.getn(path)
  end
  for i, pos in pairs(path) do
    name = IsPosInUnusedSpawnpoint(pos, store.Ch05a.Spawnpoints, store.Ch05a.Spawnpoints_complete)
    if name then
      return i
      --this stops them, if we don't stop them, then we need to store that it's true.
      --     store.Ch05a.Spawnpoints_complete["Ch01_Dialog04"] = true
    end
  end
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  if not exec.Ent.Side.Intruder then
    return
  end
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch05a.Spawnpoints, store.Ch05a.Spawnpoints_complete)

  if name == "Foyer" then
    Script.DialogBox("ui/dialog/Ch05/Ch05_Foyer.json")
    store.Ch05a.Spawnpoints_complete["Foyer"] = true
  end

  if name == "Study Desk" then
    Script.DialogBox("ui/dialog/Ch05/Ch05_Study_Desk.json")
    store.Ch05a.Spawnpoints_complete["Study Desk"] = true
  end

  if name == "Private Entrance" then
    Script.DialogBox("ui/dialog/Ch05/Ch05_Private_Entrance.json")
    store.Ch05a.Spawnpoints_complete["Private Entrance"] = true
  end

  if name == "Lab" then
    Script.DialogBox("ui/dialog/Ch05/Ch05_Lab.json")
    store.Ch05a.Spawnpoints_complete["Lab"] = true
  end

  if name == "Coffin" then
    Script.DialogBox("ui/dialog/Ch05/Ch05_Coffin.json")
    store.Ch05a.Spawnpoints_complete["Coffin"] = true
    -- Script.DialogBox("ui/dialog/Ch05/Ch05_Choose_Character")
    -- store.Ch01a.choice_a = choice[1]
  end

  if name == "Creepy1" then
    store.Ch05a.Spawnpoints_complete["Creepy1"] = true
  end

  if name == "Creepy2" then
    store.Ch05a.Spawnpoints_complete["Creepy2"] = true
  end

  if name == "Creepy3" then
    store.Ch05a.Spawnpoints_complete["Creepy3"] = true
  end
end

function RoundEnd(intruders, round)

end