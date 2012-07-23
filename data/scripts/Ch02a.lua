function Init()
  if not store.Ch02 then
  store.Ch02 = {}
  end
   store.Ch02.Spawnpoints_complete{}
   store.Ch02.Spawnpoints = {
      "Tyrees_at_door",
      "Ch02_Cordelia_Dies",
      "Ch02_Sabina_Dies",
      "Ch02_Tyree_Ghost_Dies",      
   } 

  Script.LoadHouse("Chapter_02")
  Script.DialogBox("ui/dialog/Ch02/Ch02_Dialog01.json") 

  Script.BindAi("denizen", "human")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "intruders.lua")
    --always bind one to human!
end


function doDenizenSetup()
  ents = {
    {"Bosch", 1},
  }

  setLosModeToRoomsWithSpawnsMatching("denizens", "Bosch-.*")

  placed = {}
  while table.getn(placed) == 0 do
    placed = Script.PlaceEntities("Bosch-.*", 1, ents)
  end

  ents = {
      {"Lost Soul", 1},
      {"Poltergeist", 1},
      {"Angry Shade", 1},
    }
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors-.*")
  Script.PlaceEntities("Servitors-.*", 10, ents)

  if store.Ch01c.choice_a == "Greedy" then
 	ents = {
 		{"Tyree's Ghost", 1},
 	}
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors-.*")
	Script.PlaceEntities("Servitors-.*", 1, ents)
  end
end

if store.Ch01c.choice_a == "Discretion" then
  Script.DialogBox("ui/dialog/Ch02/Ch02_Dialog02.json")
end

if store.Ch01c.choice_a == "Greedy" then
  Script.DialogBox("ui/dialog/Ch02/Ch02_Dialog02_tyree.json")

elias_spawn = Script.GetSpawnPointsMatching("Elias-Start")
Script.SpawnEntitySomewhereInSpawnPoints ("Elias Tyree", elias_spawn)
Script.SetCondition ("Elias Tyree", "Determined", true)

cordelia_spawn = Script.GetSpawnPointsMatching("Cordelia-Start")
Script.SpawnEntitySomewhereInSpawnPoints ("Cordelia Tyree", cordelia_spawn)

sabina_spawn = Script.GetSpawnPointsMatching("Sabina-Start")
Script.SpawnEntitySomewhereInSpawnPoints ("Sabina Tyree", sabina_spawn)
end

function RoundStart(denizens, round)
    Script.SetVisibility("denizens")
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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch02.Spawnpoints, store.Ch02.Spawnpoints_complete)
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

function OnDenizensAction(intruders, round, exec)
 name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch02.Spawnpoints, store.Ch02.Spawnpoints_complete)
 

function OnIntrudersAction(intruders, round, exec)
 name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch02.Spawnpoints, store.Ch02.Spawnpoints_complete)
 
  if exec.Ent.Name("Cordelia Tyree") and ent.HpCur == 0 then
    store.Ch02.Spawnpoints_complete["Ch02_Cordelia_Dies"] = true
    Script.DialogBox("Ch02_Cordelia_Dies.json")
 end

  if exec.Ent.Name("Sabina Tyree") and ent.HpCur == 0 then
  store.Ch02.Spawnpoints_complete["Ch02_Sabina_Dies"] = true
  Script.DialogBox("Ch02_Sabina_Dies.json")
  end

  if name == "Tyrees_at_door" then
  store.Ch02.Spawnpoints_complete["Ch02_Sabina_Dies"] and store.Ch02.Spawnpoints_complete["Ch02_Cordelia_Dies"] then
    Script.DialogBox("Ch02_Elias_Alone.json")
  end
  
  if store.Ch02.Spawnpoints_complete["Ch02_Cordelia_Dies"] and store.Ch02.Spawnpoints_complete["Ch02_Sabina_Dies"] then
    Script.DialogBox("Ch02_Elias_and_Sabina.json")
  end
  if store.Ch02.Spawnpoints_complete["Ch02_Cordelia_Dies"] == nil and store.Ch02.Spawnpoints_complete["Ch02_Sabina_Dies"] == true then
    Script.DialogBox("Ch02_Elias_and_Cordelia.json")
  end
  if store.Ch02.Spawnpoints_complete["Ch02_Cordelia_Dies"] == nil and store.Ch02.Spawnpoints_complete["Ch02_Sabina_Dies"] == nil then
    Script.DialogBox("Ch02_Elias_and_Both.json")
    Script.SetCondition ("Elias Tyree", "Determined", false)
  end
  
  if store.Ch01c.choice_a == "Discretion" then
    Script.DialogBox("Ch02_Bosch_Choice_Without_Ghost.json")
      store.Ch02.choice_a = choice[1]
  end
  if store.Ch01c.choice_a == "Greedy" then
    Script.DialogBox("Ch02_Bosch_Choice_With_Ghost.json")
      store.Ch02.choice_a = choice[1]
  end
   
  if store.Ch02.choice_a == "Bosch in Golem" then
    golem_spawn = Script.GetSpawnPointsMatching("Golem_Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Bosch Golem", golem_spawn)
  end
  
  if store.Ch02.choice_a == "Pact Powers Golem" then
    golem_spawn = Script.GetSpawnPointsMatching("Golem_Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Pact Golem", golem_spawn)
  end

  if store.Ch02.choice_a == "Tyree Powers Golem" then
    golem_spawn = Script.GetSpawnPointsMatching("Golem_Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Tyree Golem", golem_spawn)
  end
end

-- NEED Tyree Golem Turns

--NEED ENDING DIALOGS







