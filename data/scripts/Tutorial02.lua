 function Init(data)
  -- check data.map == "random" or something else
  Script.LoadHouse("tutorial_denizens")
  Script.DialogBox("ui/dialog/tutorial/denizens_tutorial_1_unit_selection.json")  

  store.side = "Intruders"
  Script.BindAi("denizen", "human")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "tutorial/intruders.lua")

  store.nFirstWaypointDown = false
  store.nSecondWaypointDown = false


  swaypoint_spawn = Script.GetSpawnPointsMatching("Waypoint1")
  Script.SetWaypoint("Waypoint1", "denizens", swaypoint_spawn[1].Pos, 1)
end

function intrudersSetup()
  -- intruder_names = {"Teen", "Occultist", "Ghost Hunter"}
  -- intruder_spawn = Script.GetSpawnPointsMatching("intruders_start")

  -- for _, name in pairs(intruder_names) do
  --   ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  -- end 
end

function denizensSetup() 
  Script.SetVisibility("denizens")
  setLosModeToRoomsWithSpawnsMatching("denizens", "Master_Start")
    ents = {
    {"Bosch", 1},
    }
  placed = {}
  while table.getn(placed) == 0 do
  placed = Script.PlaceEntities("Master_.*", ents, 1, 1)
  end
  store.MasterName = "Bosch"
  ServitorEnts = 
  {
    {"Angry Shade", 1},
    {"Lost Soul", 1},
    {"Vengeful Wraith", 1}
  }  

  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start1")
  placed = Script.PlaceEntities("Servitors_Start1", ServitorEnts, 0, 6)
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      intrudersSetup() 
    else
      denizensSetup()
      Script.DialogBox("ui/dialog/tutorial/denizens_tutorial_2_movement.json")
    end
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")

    Script.EndPlayerInteraction()
    return 
  end

  Script.SetVisibility("denizens")
  if not intruders then
    Script.ShowMainBar(true)
  end

  if not store.nSecondWaypointDown then
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Side.Denizen and ent.Name ~= store.MasterName then
        Script.SetAp(ent, 0)
      end
    end
  end

  store.game = Script.SaveGameState()

  if not intruders then
    side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}
    SelectCharAtTurnStart(side)
  end
end

function GetDistanceBetweenEnts(ent1, ent2)
  v1 = ent1.Pos.X - ent2.Pos.X
  if v1 < 0 then
    v1 = 0-v1
  end
  v2 = ent1.Pos.Y - ent2.Pos.Y
  if v2 < 0 then
    v2 = 0-v2
  end
  return v1 + v2
end

function GetDistanceBetweenPoints(pos1, pos2)
  v1 = pos1.X - pos2.X
  if v1 < 0 then
    v1 = 0-v1
  end
  v2 = pos1.Y - pos2.Y
  if v2 < 0 then
    v2 = 0-v2
  end
  return v1 + v2
end

function OnMove(ent, path)
  for _, ent in pairs(Script.GetAllEnts()) do
  end

  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec

  if exec.Ent.Name == store.MasterName and GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Waypoint1")[1].Pos) <= 2 and not store.nFirstWaypointDown then
    --The intruders got to the first waypoint.
    store.nFirstWaypointDown = true

    Script.RemoveWaypoint("Waypoint1") 
    Script.SetWaypoint("Waypoint2", "denizens", SelectSpawn("Waypoint2").Pos, 1)  
    Script.DialogBox("ui/dialog/tutorial/denizens_tutorial_3_doors.json")
  end 


  if store.nFirstWaypointDown then
    if exec.Ent.Name == store.MasterName and GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Waypoint2")[1].Pos) <= 2 and not store.nSecondWaypointDown then
      --The intruders got to the second waypoint.
      store.nSecondWaypointDown = true 

      Script.RemoveWaypoint("Waypoint2")
      Script.SetWaypoint("Waypoint3", "denizens", SelectSpawn("Waypoint3").Pos, 1)  

      Script.DialogBox("ui/dialog/tutorial/denizens_tutorial_4_actions.json")

      filename = "tutorial/" .. "Detective" .. ".lua"
      spawns = Script.GetSpawnPointsMatching("intruders_start")
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Tutorial Detective", spawns, false)
      Script.BindAi(ent, filename)
    end  
  end


  if store.nSecondWaypointDown then
    if exec.Ent.Name == store.MasterName and GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Waypoint3")[1].Pos) <= 2 and not store.bThirdWaypointDown then
      store.bThirdWaypointDown = true
      Script.RemoveWaypoint("Waypoint3")
      Script.DialogBox("ui/dialog/tutorial/denizens_tutorial_5_actions_two.json")

      spawns = Script.GetSpawnPointsMatching("intruders_start2")
      filename = "tutorial/" .. "Occultist" .. ".lua"
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Tutorial Occultist", spawns, false)
      Script.BindAi(ent, filename)

      filename = "tutorial/" .. "Teen" .. ".lua"
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Tutorial Teen", spawns, false)
      Script.BindAi(ent, filename)
    end   
  end

  if store.bThirdWaypointDown and not AnyIntrudersAlive() then
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/tutorial/Finale_Tutorial_Denizens.json")
  end

  if not MasterIsAlive() then
    --respawn bosch
    master_spawn = Script.GetSpawnPointsMatching("Master_Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Bosch", master_spawn, false)   
  end 

  --after any action, if this ent's Ap is 0, we can select the next ent for them
  if exec.Ent.ApCur == 0 then 
    nextEnt = GetEntityWithMostAP(exec.Ent.Side)
    if nextEnt.ApCur > 0 then
      if exec.Action.Type ~= "Move" then
        Script.Sleep(2)
      end
      Script.SelectEnt(nextEnt)
    end
  end  
end

function SelectSpawn(SpawnName)
  possible_spawns = Script.GetSpawnPointsMatching(SpawnName)
  bUsedOne = false   
  for _, spawn in pairs(possible_spawns) do
    if Script.Rand(4) > 2 then
      return spawn
    end 
  end  
  return possible_spawns[1]      
end
 

function RoundEnd(intruders, round)
  if round == 1 then
    return
  end

  Script.ShowMainBar(false)

  store.execs = {}
end

function MasterIsAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == store.MasterName and ent.HpCur > 0 then
      return true
    end
  end
  return false  
end

function AnyIntrudersAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder and ent.HpCur > 0 then
      return true
    end
  end
  return false  
end

function SelectCharAtTurnStart(side)  
    --select the dood with the most AP
    Script.SelectEnt(GetEntityWithMostAP(side))
end

function GetEntityWithMostAP(side)
  entToSelect = nil
  for _, ent in pairs(Script.GetAllEnts()) do
    if (ent.Side.Intruder and side.Intruder) or (ent.Side.Denizen and side.Denizen) then   
      if entToSelect then    
        if entToSelect.ApCur < ent.ApCur then      
          entToSelect = ent
        end 
      else
        --first pass.  select this one.
        entToSelect = ent
      end
    end
  end
  return entToSelect
end

function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function GetMasterEnt()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == store.MasterName then
      return ent
    end
  end
end