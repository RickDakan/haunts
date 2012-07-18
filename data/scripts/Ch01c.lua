function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end
--
play_as_denizens = false
function Init()
   store.Ch01c = {}
   store.Ch01c.Dialog_complete = {}
   store.Ch01c.Dialogs = {
      "Ch01_Dialog01",
      "Ch01_Dialog02",
      "Ch01_Dialog03",
      "Ch01_Dialog04",
      "Ch01_Dialog05",
      "Ch01_Dialog06",
      "Ch01_Dialog07",
      "Ch01_Dialog08",
      "Ch01_Dialog09",
      "Ch01_Dialog10",
      "Chair Trigger01",
      "Foyer Trigger01",
      "Greathall Trigger01",
      "Finale Trigger01",
      "Exit",
   } 

  Script.LoadHouse("Chapter_01_c")

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
 
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Caitlin", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Percy", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Harry", intruder_spawn)
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
  print("Regexp: '",regexp,"'")
  sps = Script.GetSpawnPointsMatching(regexp)
  print("Got ", "spawns for ", regexp)
  for _, sp in pairs(sps) do
    if pointIsInSpawn(pos, sp) then
      return true
    end
  end
  return false
end


function IsPosInUnusedSpawnpoint(pos, name, list)
  --name identifies spawnpoint
  for _, spawn in pairs(list) do
    if not used[name] and pointIsInSpawns(pos, name) then
      return name
    end
  end
  return nil
end

function OnMove(ent, path)
  if not ent.Side.Intruder then
    return table.getn(path)
  end
  for i, pos in pairs(path) do
    name = IsPosInUnusedSpawnpoint(pos, store.Spawnpoints, store.Spawnpoints_complete)
    if name then
      return i
    end
  end
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  if not exec.ent.Side.Intruder then
    return
  end
  name = IsPosInUnusedSpawnpoint(pos, store.Spawnpoints, store.Spawnpoints_complete)
  if name then
    dialog_path = "ui/dialog/Ch01/" .. name .. ".json"
    Script.DialogBox(dialog_path)
    store.Ch01.Spawnpoints_complete[name] = true 
   
    if name == "Chair Trigger01" then
    chair_spawn = Script.GetSpawnPointsMatching("Chair Spawn01")
    SpawnEntitySomewhereInSpawnPoints("Poltergeist", chair_spawn)
    end

    if name == "Foyer Trigger01" then
    angry_shade_spawn = Script.GetSpawnPointsMatching("Angry Shade Spawn02")
    SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
    SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
    SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
    SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
    SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
    end

    if name == "Greathall Trigger01" then
    lost_soul_spawn = Script.GetSpawnPointsMatching("Lost Soul Spawn01")
    SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
    SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
    SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
    SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
    SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
    end

    if name == "Finale Trigger01" then
      Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
        choices = Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
        store.Ch01c.choice_a = choices[1]
      bosch_spawn = Script.GetSpawnPointsMatching ("Bosch")
      intruder_spawn02 = Script.GetSpawnPointsMatching ("Intruders Spawn02")
      finale_shade_spawn = Script.GetSpawnPointsMatching ("Shade Finale")

    if store.Ch01c.choice_a == "Greedy" then
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)

    if store.Ch01c.choica_a == "Discretion" then
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)

    if name == "Exit" and store.Ch01c.choice_a == "Greedy" then
     
      if ent.Name == "Caitlin" then
        Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog11.json")
      end
      if ent.Name == "Percy" then
        Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog12.json")
      end
    end

    if name == "Exit" and store.Ch01c.choice_a == "Discretion" then
      Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog13.json")
    end
end
