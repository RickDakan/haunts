function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function IsStoryMode()
  return true
end

function DoTutorials()
  --We should totally do some tutorials here.
  --It would be super cool.
end

function Side()
  if Net.Active() then
    return Net.Side()
  end
  return store.side
end

function Init(data)
  if Net.Active() then
    side_choices = {"Denizens"}
  else
    side_choices = Script.ChooserFromFile("ui/start/versus/side.json")
  end

  -- check data.map == "random" or something else
  Script.LoadHouse("Lvl_01_Haunted_House")  
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")

  store.side = side_choices[1]
  -- Side() = "Humans"
  if Side() == "Humans" or Net.Active() then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  else
    if Side() == "Denizens" then
      Script.BindAi("denizen", "human")
      Script.BindAi("minions", "minions.lua")
      Script.BindAi("intruder", "ch01/intruders.lua")
    end
    if Side() == "Intruders" then
      Script.BindAi("denizen", "ch01/denizens.lua")
      Script.BindAi("minions", "minions.lua")
      Script.BindAi("intruder", "human")
    end
  end

  store.nFirstWaypointDown = false
  store.nSecondWaypointDown = false


  store.waypoint_spawn = Script.GetSpawnPointsMatching("Waypoint1")
  store.Waypoint1 = Script.SpawnEntitySomewhereInSpawnPoints("Table", store.waypoint_spawn, false)
  Script.SetWaypoint("Waypoint1" , "intruders", store.Waypoint1.Pos, 3)

  -- StoreWaypoint("Waypoint1", "intruders", store.Waypoint1.Pos, 3, false)    
end

function intrudersSetup()
  if IsStoryMode() then
    intruder_names = {"Teen", "Occultist", "Ghost Hunter"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  -- else
  --   --permit all choices for normal vs play
  end

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
    if Side() == "Denizens" then
      filename = "ch01/" .. name .. ".lua"
      Script.BindAi(ent, filename)
    end
  end 
end

function denizensSetup()
  -- This creates a list of entities and associated point values.  The order
  -- the names are listed in here is the order they will appear to the user.
  if IsStoryMode() then
    ents = {
      {"Bosch", 1},
    }
  else
    --permit all choices for normal vs play.

  end
  
  if Side() == "Intruders" then
    master_spawn = Script.GetSpawnPointsMatching("Master_.*")
    ent = Script.SpawnEntitySomewhereInSpawnPoints("Bosch", master_spawn, false)
    placed = {ent}
    Script.BindAi(ent, "ch01/Bosch.lua")
  else
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Master_.*")
    placed = Script.PlaceEntities("Master_.*", ents, 1, 1)
  end

  if placed[1].Name == "Chosen One" then
    store.MasterName = "Chosen One"
    ServitorEnts = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  end
  if placed[1].Name == "Bosch" then
    store.MasterName = "Bosch"
    ServitorEnts = {
      {"Angry Shade", 1},
      {"Lost Soul", 1},
    }  
  end
--      {"Vengeful Wraith", 3},

  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  points = 6
  if Side() == "Intruders" then
    servitor_spawn = Script.GetSpawnPointsMatching("Servitors_Start1")
    while points > 0 do
      option = ServitorEnts[Script.Rand(table.getn(ServitorEnts))]
      if option[2] <= points then
        ent = Script.SpawnEntitySomewhereInSpawnPoints(option[1], servitor_spawn, true)
        Script.BindAi(ent, "ch01/" .. option[1] .. ".lua")
        points = points - option[2]
      end
    end
  else
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start1")
    Script.PlaceEntities("Servitors_Start1", ServitorEnts, 0, points)
  end
end

function RoundStart(intruders, round)
  print("SCRIPT: Round Start")
  store.execs = {}
  if round == 1 then
    if intruders then
      intrudersSetup() 
    else
      Script.DialogBox("ui/dialog/Lvl01/Opening_Denizens.json")
      Script.FocusPos(Script.GetSpawnPointsMatching("Master_Start")[1].Pos)
      denizensSetup()
    end
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")

    if IsStoryMode() then
      DoTutorials()
    end

    Script.EndPlayerInteraction()
    store.game = Script.SaveGameState()
    Net.UpdateState(store.game)
    return
  end

  if store.nFirstWaypointDown and not store.bSetup2Done then
    store.bSetup2Done = true
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start2")
    placed = Script.PlaceEntities("Servitors_Start2", ServitorEnts, 0, ValueForReinforce())
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
  end
  

  if store.nSecondWaypointDown and not store.bSetup3Done then
    store.bSetup3Done = true
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start3")
    placed = Script.PlaceEntities("Servitors_Start3", ServitorEnts, 0, ValueForReinforce())
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
  end

  spawns = Script.GetSpawnPointsMatching("Servitors_Start1")
  store.game = Script.SaveGameState()
  print("Update State Round/Intruders: ", round, intruders)
  Net.UpdateState(store.game)

  side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}
  SelectCharAtTurnStart(side)

  if Side() == "Humans" then
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("intruders")
    else
      Script.SetVisibility("denizens")
    end
    Script.ShowMainBar(true)
  else
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("intruders")
      Script.ShowMainBar(intruders)
    else
      Script.SetVisibility("denizens")
      Script.ShowMainBar(not intruders)
    end
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

function ValueForReinforce()
  --The denizens get to reinforce after each waypoint goes down.
  --They get 6 - (value of units on the board) + (2 for each dead waypoint)

  nValToReturn = 0
  nTotalValueOnBoard = 0
  for _, ent in pairs(Script.GetAllEnts()) do
    for _, entValue in pairs(ServitorEnts) do
      if ent.Name == entValue[1] then
        nTotalValueOnBoard = nTotalValueOnBoard + entValue[2]
      end 
    end
  end
  nValToReturn = (6 - nTotalValueOnBoard) 
  if store.nFirstWaypointDown then
    nValToReturn = nValToReturn + 2
  end
  if store.nSecondWaypointDown then
    nValToReturn = nValToReturn + 2
  end
  return nValToReturn
end

function OnMove(ent, path)
  for _, ent in pairs(Script.GetAllEnts()) do
  end

  return table.getn(path)
end

function SelectSpawn(SpawnName)
  math.randomseed(os.time())
  possible_spawns = Script.GetSpawnPointsMatching(SpawnName)
  bUsedOne = false   
  for _, spawn in pairs(possible_spawns) do
    if math.random(4) > 2 then
      return spawn
    end 
  end  
  return possible_spawns[1]      
end

-- Does any special processing from an exec.  This is in its own function because
-- it might be called during DoAction if the exec occurred locally, or in a
-- playback if the exec occurred remotely.
function checkExec(exec)
  print("SCRIPT: Checking Exec")
  print("SCRIPT:", exec.Ent)
  if exec.Ent then
    print("SCRIPT:", exec.Ent.Side.Intruder)
    print("SCRIPT: Ent - ", exec.Ent.Pos.X, exec.Ent.Pos.Y)
    print("SCRIPT: Ent - ", store.Waypoint1.Pos.X, store.Waypoint1.Pos.Y)
    print("SCRIPT:", GetDistanceBetweenEnts(exec.Ent, store.Waypoint1))
    print("SCRIPT:", store.nFirstWaypointDown)
  end
  if exec.Ent and exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint1) <= 3 and not store.nFirstWaypointDown then
    --The intruders got to the first waypoint.
    store.nFirstWaypointDown = 2 --2 because that's what we want to add to the deni's deploy 
    print("WAYPOINT SET MONKEY")
    store.waypoint_spawn = SelectSpawn("Waypoint2") 
    store.Waypoint2 = StoreSpawn("Chest",  store.waypoint_spawn.Pos)   
    Script.DialogBox("ui/dialog/Lvl01/First_Waypoint_Down_Intruders.json")
    Script.SetMusicParam("tension_level", 0.3) 

    StoreWaypoint("Waypoint1", "", "", "", true)
    StoreWaypoint("Waypoint2", "intruders", store.Waypoint2.Pos, 3, false)  
    Script.RemoveWaypoint("Waypoint1")
    Script.SetWaypoint("Waypoint2", "intruders", store.Waypoint2.Pos, 3)   
  end 


  if store.nFirstWaypointDown then
    if exec.Ent and exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint2) <= 3 and not store.nSecondWaypointDown then
      --The intruders got to the second waypoint.
      store.nSecondWaypointDown = 2 --2 because that's what we want to add to the deni's deploy 
      store.waypoint_spawn = SelectSpawn("Waypoint3") 
      store.Waypoint3 = StoreSpawn("Mirror", store.waypoint_spawn.Pos)
      Script.SetMusicParam("tension_level", 0.5) 
      Script.DialogBox("ui/dialog/Lvl01/Second_Waypoint_Down_Intruders.json")    

      StoreWaypoint("Waypoint2", "", "", "", true)
      StoreWaypoint("Waypoint3", "intruders", store.Waypoint3.Pos, 3, false) 
      Script.RemoveWaypoint("Waypoint2")
      Script.SetWaypoint("Waypoint3", "intruders", store.Waypoint3.Pos, 3)             
    end  
  end


  if store.nSecondWaypointDown then
    if exec.Ent and exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint3) <= 3 then
      --The intruders got to the third waypoint.  Game over, man.  Game over.
      Script.DialogBox("ui/dialog/Lvl01/Victory_Intruders.json")
      Script.SetMusicParam("tension_level", 0.7)
    end   
  end


  if not AnyIntrudersAlive() then
    Script.DialogBox("ui/dialog/Lvl01/Victory_Denizens.json")
  end 


  -- --after any action, if this ent's Ap is 0, we can select the next ent for them
  -- if exec.Ent.ApCur == 0 then 
  --   nextEnt = GetEntityWithMostAP(exec.Ent.Side)
  --   if nextEnt.ApCur > 0 then
  --     Script.SelectEnt(nextEnt)
  --   end
  -- end  
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec
  checkExec(exec)
end
 

function DoPlayback(state, execs)
  print("SCRIPT: DoPlayback - ", table.getn(execs))
  Script.LoadGameState(state)

  --focus the camera on somebody on each team.
  side2 = {Intruder = not intruders, Denizen = intruders, Npc = false, Object = false}  --reversed because it's still one side's turn when we're replaying their actions for the other side.
  Script.FocusPos(GetEntityWithMostAP(side2).Pos)

  for _, exec in pairs(execs) do
    print("SCRIPT: proc exec")
    bDone = false
    if exec.script_spawn then
      print("SRCIPT: Exec - Spawn")
      doSpawn(exec)
      bDone = true
    end
    if exec.script_despawn then
      print("SRCIPT: Exec - Despawn")
      deSpawn(exec)
      bDone = true
    end
    if exec.script_waypoint then
      print("SRCIPT: Exec - Waypoint")
      doWaypoint(exec)
      bDone = true
    end
    if not bDone then
      print("SRCIPT: Exec - Standard")
      Script.DoExec(exec)

      --will be used at turn start to try to reselect the last thing they acted with.
      if exec.Ent.Side == "intruders" then
        LastIntruderEnt = exec.Ent
      end
      if exec.Ent.Side == "denizens" then
        LastDenizenEnt = exec.Ent
      end
    end
    checkExec(exec)
  end
end

function denizensOnRound()
  if store.nFirstWaypointDown and not store.bShowedFirstWaypointMessage then
    store.bShowedFirstWaypointMessage = true
    Script.DialogBox("ui/dialog/Lvl01/First_Waypoint_Down_Denizens.json")
  end

  if store.nSecondWaypointDown and not store.bShowedSecondWaypointMessage then
    store.bShowedSecondWaypointMessage = true
    Script.DialogBox("ui/dialog/Lvl01/Second_Waypoint_Down_Denizens.json")
  end
end

function RoundEnd(intruders, round)
  print("SCRIPT: Round End:", Net.Side())
  print("Update Execs Round/Intruders: ", round, intruders)
  Net.UpdateExecs(store.execs)
  if Net.Active() then
    if Side() == "Denizens" then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end
    Script.ShowMainBar(false)
    Net.Wait()
    -- cur = Script.SaveGameState()
    state, execs = Net.LatestStateAndExecs()
    DoPlayback(state, execs)
    if Side() == "Denizens" then
      denizensOnRound()
    end
    Script.ShowMainBar(true)
    return
  end

  if round == 1 then
    return
  end

  bSkipOtherChecks = false  --Resets this every round

  if Side() == "Humans" then
    Script.ShowMainBar(false)
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end

    if intruders then
      Script.DialogBox("ui/dialog/Lvl01/pass_to_denizens.json")
      denizensOnRound()
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl01/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl01/Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl01/pass_to_intruders.json")
      end
    end

    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    DoPlayback(store.game, store.execs)
  end


  --if the Master is down, respawn him
  if intruders then
    if not MasterIsAlive() then
      master_spawn = Script.GetSpawnPointsMatching("Master_Start")
      if store.MasterName == "Bosch" then
        store.MasterName = "Bosch's Ghost"
        store.bUsingGhostBosch = true 
        Script.DialogBox("ui/dialog/Lvl01/Bosch_Rises_Denizens.json")
        store.bBoschRespawnedTellIntruders = true
      end
      Script.SpawnEntitySomewhereInSpawnPoints(store.MasterName, master_spawn, false)
    end
  else
    if store.bBoschRespawnedTellIntruders then
      Script.DialogBox("ui/dialog/Lvl01/Bosch_Rises_Intruders.json")
      store.bBoschRespawnedTellIntruders = false --keep this dialogue from getting triggered ever again
    end
  end

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
  bDone = false
  if LastIntruderEnt then
    if side.Intruder then
   -- if LastIntruderEnt.Side == side then
      Script.SelectEnt(LastIntruderEnt)
      bDone = true
    end
  end  
  if LastDenizenEnt and not bDone then
    if side.Denizen then
--    if LastDenizenEnt.Side == side then      
      Script.SelectEnt(LastDenizenEnt)
      bDone = true
    end  
  end   

  if not bDone then
    --select the dood with the most AP
    Script.SelectEnt(GetEntityWithMostAP(side))
  end  
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

function GetSpawnsFromListWhereNoLoS(spawns)
  GoodSpawns = {}
  for _, possibleSpawn in pairs(spawns) do
    --nasty set of loops here.
    bBadSpawn = false
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Side.Intruder then
        for _, visibleSpawn in pairs(Script.GetLos(ent)) do
          if (visibleSpawn.Pos.X == possibleSpawn.Pos.X) and (visibleSpawn.Pos.Y == possibleSpawn.Pos.Y) then
            bBadSpawn = true
            break
          end 
        end
      end
      if bBadSpawn then
        --no sense in continuing for this possible spawn.  We already know it's bad.
        break
      end
    end
    if not bBadSpawn then
        GoodSpawns[table.getn(GoodSpawns) + 1] = possibleSpawn
    end
  end
  
  return GoodSpawns
end

function StoreSpawn(entName, spawnPos)
  spawn_exec = {script_spawn=true, name=entName, pos=spawnPos}
  store.execs[table.getn(store.execs) + 1] = spawn_exec
  return doSpawn(spawn_exec)
end

function doSpawn(spawnExec)
  return Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
  deSpawn(despawn_exec)
end

function deSpawn(despawnExec)
  if despawnExec.entity.HpMax then  --can only kill things that have hp
    Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
    Script.SetHp(despawnExec.entity, 0)
  end
  DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
end

function StoreWaypoint(wpname, wpside, wppos, wpradius, wpremove)
  waypoint_exec = {script_waypoint=true, name=wpname, side=wpside, pos=wppos, radius=wpradius, remove=wpremove}
  store.execs[table.getn(store.execs) + 1] = waypoint_exec
  doWaypoint(waypoint_exec)
end

function doWaypoint(waypointExec)
  if waypointExec.remove then
    return Script.RemoveWaypoint(waypointExec.name)
  else
    return Script.SetWaypoint(waypointExec.name, waypointExec.side, waypointExec.pos, waypointExec.radius)
  end
end