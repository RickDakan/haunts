function Init(data)
  -- check data.map == "random" or something else
  Script.LoadHouse("tutorial_intruders")
  Script.DialogBox("ui/dialog/tutorial/intruders_tutorial_1_units_explained.json")  

  store.side = "Intruders"
  Script.BindAi("denizen", "tutorial/denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")

  store.nFirstWaypointDown = false
  store.nSecondWaypointDown = false


  swaypoint_spawn = Script.GetSpawnPointsMatching("Waypoint1")
  Script.SetWaypoint("Waypoint1", "intruders", swaypoint_spawn[1].Pos, 1)
end

function intrudersSetup()
  intruder_names = {"Teen", "Occultist", "Ghost Hunter"}
  intruder_spawn = Script.GetSpawnPointsMatching("intruders_start")

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end 
end

function denizensSetup() 
  store.MasterName = "Bosch"
  ServitorEnts = {
    {"Angry Shade", 1},
    {"Lost Soul", 1},
  }  
end

function RoundStart(intruders, round)
  side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}
  if round == 1 then
    if intruders then
      intrudersSetup() 
    else
      denizensSetup()
    end
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    SelectCharAtTurnStart(side)
    Script.EndPlayerInteraction()
    return 
  end

  Script.SetVisibility("intruders")
  if intruders then
  Script.ShowMainBar(true)
  end

  store.game = Script.SaveGameState()
  SelectCharAtTurnStart(side)
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

  if exec.Ent.Side.Intruder and GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Waypoint1")[1].Pos) <= 2 and not store.nFirstWaypointDown then
    --The intruders got to the first waypoint.
    store.nFirstWaypointDown = true

    Script.RemoveWaypoint("Waypoint1") 
    Script.SetWaypoint("Waypoint2", "intruders", SelectSpawn("Waypoint2").Pos, 1)  
    Script.DialogBox("ui/dialog/tutorial/intruders_tutorial_3_doors.json")
  end 


  if store.nFirstWaypointDown then
    if  exec.Ent.Side.Intruder and GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Waypoint2")[1].Pos) <= 2 and not store.nSecondWaypointDown then
      --The intruders got to the second waypoint.
      store.nSecondWaypointDown = true 

      Script.RemoveWaypoint("Waypoint2")
      Script.SetWaypoint("Waypoint3", "intruders", SelectSpawn("Waypoint3").Pos, 1)  

      Script.DialogBox("ui/dialog/tutorial/intruders_tutorial_4_actions.json")

      --store.bSpawnDudes = true
      filename = "tutorial/" .. "Shade" .. ".lua"
      spawns = Script.GetSpawnPointsMatching("Servitors_Start1")
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", spawns, false)
      Script.BindAi(ent, filename)
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", spawns, false)
      Script.BindAi(ent, filename)
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", spawns, false)
      Script.BindAi(ent, filename)

      filename = "tutorial/" .. "Wisp" .. ".lua"
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Lost Soul", spawns, false)
      Script.BindAi(ent, filename)
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Lost Soul", spawns, false)
      Script.BindAi(ent, filename)
    end  
  end


  if store.nSecondWaypointDown then
    if  exec.Ent.Side.Intruder and GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Waypoint3")[1].Pos) <= 2 and not store.bThirdWaypointDown then
      store.bThirdWaypointDown = true
      Script.DialogBox("ui/dialog/tutorial/intruders_tutorial_5_actions_two.json")
      spawns = Script.GetSpawnPointsMatching("Master_start")
      ent = Script.SpawnEntitySomewhereInSpawnPoints("Bosch", spawns, false)
      filename = "tutorial/" .. "Bosch" .. ".lua"
      Script.BindAi(ent, filename)
      Script.RemoveWaypoint("Waypoint3")
    end   
  end

  if store.bThirdWaypointDown and not MasterIsAlive() then
    --Script.Sleep(2)
    Script.DialogBox("ui/dialog/tutorial/Finale_Tutorial_Intruders.json")
  end

  if not AnyIntrudersAlive() then
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/Lvl01/Victory_Denizens.json")
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