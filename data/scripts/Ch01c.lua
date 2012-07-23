function Init()
  if not store.Ch01c then
    store.Ch01c = {}
  end
   store.Ch01c.Spawnpoints_complete = {}
   store.Ch01c.Spawnpoints = {
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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch01c.Spawnpoints, store.Ch01c.Spawnpoints_complete)
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
  --LOOK AT THIS AND LEARN!!!
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch01c.Spawnpoints, store.Ch01c.Spawnpoints_complete)
 
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
    print("Spawns:", table.getn(chair_spawn))
    ent = Script.SpawnEntitySomewhereInSpawnPoints("Vengeful Wraith", chair_spawn)
    print("ent:",ent)
    store.Ch01c.Spawnpoints_complete["Chair Trigger01"] = true
  end

  if name == "Foyer Trigger01" then
    angry_shade_spawn = Script.GetSpawnPointsMatching("Angry Shade Spawn02")
     for i = 1,5 do
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
     end
    store.Ch01c.Spawnpoints_complete["Foyer Trigger01"] = true
  end   

  if name == "Greathall Trigger01" then
    lost_soul_spawn = Script.GetSpawnPointsMatching("Lost Soul Spawn01")
      for i = 1,4 do
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Lost Soul", lost_soul_spawn)
     end   
    store.Ch01c.Spawnpoints_complete["Greathall Trigger01"] = true
  end

-- REPLACE HARRY WITH BOSCH

  if name == "Finale Trigger01" then
    choices = Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
    store.Ch01c.choice_a = choices[1]
    store.Ch01c.Spawnpoints_complete["Finale Trigger01"] = true
    

   
    if store.Ch01c.choice_a == "Greedy" then
      finale_shade_spawn = Script.GetSpawnPointsMatching ("Shade Finale")
      for i = 1,8 do
        Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", finale_shade_spawn)
      end
      for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Harry" then
          Script.SetHp(ent, 0)
        end
      end
      bosch_spawn = Script.GetSpawnPointsMatching ("Bosch")
      Script.SpawnEntitySomewhereInSpawnPoints("Angry Bosch", bosch_spawn)
    end
    

    if store.Ch01c.choice_a == "Discretion" then
      finale_shade_spawn = Script.GetSpawnPointsMatching ("Shade Finale")
      for i = 1,3 do
        Script.SpawnEntitySomewhereInSpawnPoints("Shade", finale_shade_spawn)
      end
      for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Harry" then
          Script.SetHp(ent, 0)
        end
      end
      bosch_spawn = Script.GetSpawnPointsMatching ("Bosch")
      Script.SpawnEntitySomewhereInSpawnPoints("Scornful Bosch", bosch_spawn)
      end
    end

    if name == "Exit" and store.Ch01c.choice_a == "Greedy" then 
      if exec.Ent.Name == "Caitlin" then
        Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog11.json")
      end
      if exec.Ent.Name == "Percy" then
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