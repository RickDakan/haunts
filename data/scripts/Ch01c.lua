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
  if not store.Ch01c then
    store.Ch01c = {}
  end
   store.Ch01c = {}
   store.Ch01c.Dialog_complete = {}
   store.Ch01c.Dialogs = {
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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch01c.Spawnpoints, store.Ch01c.Spawnpoints_complete)
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
  name = IsPosInUnusedSpawnpoint(pos, store.Ch01c.Spawnpoints, store.Ch01c.Spawnpoints_complete)
  if name == "Ch01_Dialog07" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog07.json")
    store.Ch01c.Spawnpoints_complete["Ch01_Dialog07"] = true
  end

  if name == "Ch01_Dialog08" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog08.json")
    store.Ch01c.Spawnpoints_complete["Ch01_Dialog08"] = true
  end

  if name == "Ch01_Dialog09" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog09.json")
    store.Ch01c.Spawnpoints_complete["Ch01_Dialog09"] = true
  end

  if name == "Ch01_Dialog10" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
    store.Ch01c.Spawnpoints_complete["Ch01_Dialog10"] = true
  end

  if name == "Chair Trigger01" then
    chair_spawn = Script.GetSpawnPointsMatching("Chair Spawn01")
    SpawnEntitySomewhereInSpawnPoints("Poltergeist", chair_spawn)
    store.Ch01c.Spawnpoints_complete("Chair Trigger01") = true
  end

  if name == "Foyer Trigger01" then
    angry_shade_spawn = Script.GetSpawnPointsMatching("Angry Shade Spawn02")
     for i = 1,5 do
      ent = SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
     end
    store.Ch01c.Spawnpoints_complete("Foyer Trigger01") = true
  end   

  if name == "Greathall Trigger01" then
    lost_soul_spawn = Script.GetSpawnPointsMatching("Lost Soul Spawn01")
      for i = 1,4 do
      ent = SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
     end   
    store.Ch01c.Spawnpoints_complete("Greathall Trigger01") = true
  end

  if name == "Finale Trigger01" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
      choices = Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
      store.Ch01c.choice_a = choices[1]
    bosch_spawn = Script.GetSpawnPointsMatching ("Bosch")
    intruder_spawn02 = Script.GetSpawnPointsMatching ("Intruders Spawn02")
    finale_shade_spawn = Script.GetSpawnPointsMatching ("Shade Finale")
    store.Ch01c.Spawnpoints_complete("Finale Trigger01") = true
      if store.Ch01c.choice_a == "Greedy" then
        for i = 1,8 do
          SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
        end
      end

      if store.Ch01c.choica_a == "Discretion" then
        for i = 1,8 do
          SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
        end
      end
    end

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


function RoundEnd(intruders, round)
  if not intruders then
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Shade" then
        if ent.HpCur == 1 then
          Script.PlayAnimations(ent, {"defend", "killed"})
          Script.SetHp(ent, 0)
        end
        if ent.HpCur > 1 then 
          Script.SetHp(ent, 1)
        end
      end
    end
  end
--SOMETHING FOR GAME END
-- WIN and LOSE

end