--Chapter 2
function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end



play_as_denizens = true
function Init()
  if not store.Ch02 then
  store.Ch02 = {}
  end
   store.Ch02.Spawnpoints_complete{}
   store.Ch02.Spawnpoints = {
      "Tyrees_at_door",
   } 

  Script.LoadHouse("Chapter_02")
  Script.DialogBox("ui/dialog/Ch01/Ch02_Dialog01.json") 

  Script.BindAi("denizen", "human")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "intruders.lua")
    --always bind one to human!

  store.cordelia_dead = false
  store.sabina_dead = false
end


function doDenizenSetup()
  ents = {
    {"Bosch", 1},
  }

  setLosModeToRoomsWithSpawnsMatching("denizens", "Master-.*")

  placed = {}
  while table.getn(placed) == 0 do
    placed = Script.PlaceEntities("Master-.*", 1, ents)
  end
  ents = {
      {"Lost Soul", 1},
      {"Poltergeist", 1},
      {"Angry Shade", 1},
    }
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitor-.*")
  Script.PlaceEntities("Servitor-.*", 10, ents)

  if store.Ch01c.choice_a == "Greedy" then
 	ents = {
 		{"Tyree's Ghost", 1},
 	}
	Script.PlaceEntities("Servitor-.*", 1, ents)
	Script.SetLosMode("denizens", "none")
  end
end

if store.Ch01c.choice_a == "Discretion" then
  Script.DialogBox("ui/dialog/Ch02/Ch02_Dialog02.json")
end

if store.Ch01c.choice_a == "Greedy" then
--  Script.DialogBox("ui/dialog/Ch02/Ch02_Dialog02.json")




elias_spawn = Script.GetSpawnPointsMatching("Elias-Start")
Script.SpawnEntitySomewhereInSpawnPoints ("Elias Tyree", elias_spawn)

cordelia_spawn = Script.GetSpawnPointsMatching("Cordelia-Start")
Script.SpawnEntitySomewhereInSpawnPoints ("Cordelia Tyree", cordelia_spawn)

sabina_spawn = Script.GetSpawnPointsMatching("Sabina-Start")
Script.SpawnEntitySomewhereInSpawnPoints ("Sabina Tyree", sabina_spawn)

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


function OnAction(denizens, round, exec)

 if ent.Name("Cordelia Tyree") and ent.HpCur == 0 then
  store.cordelia_dead = true
  Script.DialogBox("Ch02_Cordelia_Dies")
 end

 if ent.Name("Sabina Tyree") and ent.HpCur == 0 then
  store.cordelia_dead = true
  Script.DialogBox("Ch02_Sabina_Dies")
 end

--saving Elias



function OnAction(intruders, round, exec)
  if name == "Tyrees_at_door" then
    if store.cordelia_dead == true and store.sabina_dead == true then
      Script.DialogBox("Ch02_Elias_Alone")
    end
    if store.cordelia_dead == true and store.sabina_dead == false then
      Script.DialogBox("Ch02_Elias_and_Sabina")
    end
    if store.cordelia_dead == false and store.sabina_dead == true then
      Script.DialogBox("Ch02_Elias_and_Cordelia")
    end
    if store.cordelia_dead == false and store.sabina_dead == false then
      Script.DialogBox("Ch02_Elias_and_Both")
    end
  if store.Ch01c.choice_a == "Discretion" then
    Script.DialogBox("Ch02_Bosch_Choice_Without_Ghost")
      store.Ch02.choice_a = choice[1]
  end
  if store.Ch01c.choice_a == "Greedy" then
    Script.DialogBox("Ch02_Bosch_Choice_With_Ghost")
      store.Ch02.choice_a = choice[1]
  end
  --choices determine spawn
  -- maybe  kill Bosch

end









