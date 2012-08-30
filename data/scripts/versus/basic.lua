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

function Init(data)
  side_choices = Script.ChooserFromFile("ui/start/versus/side.json")

  -- check data.map == "random" or something else
  Script.LoadHouse("Lvl_01_Haunted_House")  
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")
  Script.SetWaypoint("Kittens", "denizens", {X=20, Y=20}, 5)
  Script.SetWaypoint("Kittens", "denizens", {X=25, Y=25}, 3)
  -- This gives intruders vision of the Waypoint spawn points.  Probably
  -- you'll want to change this to only give vision of Waypoint1 at the
  -- start, and then update it appropriately as the game progresses.
  Script.SetVisibleSpawnPoints("intruders", ".*[Ww]ay.*")

  store.side = side_choices[1]
  if store.side == "Humans" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end
  if store.side == "Denizens" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "ch01/intruders.lua")
  end
  if store.side == "Intruders" then
    Script.BindAi("denizen", "denizens.lua")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end

  waypoint_spawn = Script.GetSpawnPointsMatching("Waypoint1")
  store.Waypoint1 = Script.SpawnEntitySomewhereInSpawnPoints("Waypoint", waypoint_spawn)

  store.nFirstWaypointDown = false
  store.nSecondWaypointDown = false
end

function intrudersSetup()
  print("Script: intruders setup")
  if IsStoryMode() then
    intruder_names = {"Teen", "Occultist", "Researcher"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  -- else
  --   --permit all choices for normal vs play
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn)
    if store.side == "Denizens" then
      Script.BindAi(ent, "ch01/" .. name .. ".lua")
    end
  --Don't understand hgear yet...halp!?
    Script.SetGear(ent, "Pre-loaded Playlist")
    -- PRETEND!
  end

  -- Choose entry point here.
  Script.SaveStore()
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
  
  Script.SetVisibility("denizens")
  setLosModeToRoomsWithSpawnsMatching("denizens", "Master_.*")

  -- Now we give the user a ui with which to place these entities.  The user
  -- will have 1 point to spend, and each of the options costs one point, so
  -- they will only place 1.  We will make sure they place exactly one.
  -- Also the "Master-.*" indicates that the entity can only be placed in
  -- spawn points that have a name that matches the regular expression
  -- "Master-.*", which means anything that begins with "Master-".  So
  -- "Master-BackRoom" and "Master-Center" both match, for example.
  placed = {}
  while table.getn(placed) == 0 do
    sps = Script.GetSpawnPointsMatching("Master_.*")
    Script.FocusPos(sps[1].Pos)
    placed = Script.PlaceEntities("Master_.*", ents, 1, 1)
    break
  end

  -- placed is an array containing all of the entities placed, in this case
  -- there will only be one, and we will use that one to determine what
  -- servitors to make available to the user to place.
  if placed[1].Name == "Chosen One" then
    store.MasterName = "Chosen One"
    store.ServitorEnts = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  end
  if placed[1].Name == "Bosch" then
    store.MasterName = "Bosch"
    store.ServitorEnts = {
      {"Angry Shade", 1},
      {"Lost Soul", 1},
    }  
  end
--      {"Vengeful Wraith", 3},
  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start1")
  placed = Script.PlaceEntities("Servitors_Start1", store.ServitorEnts, 0,6)
end

function RoundStart(intruders, round)
  print("Script: RoundStart", intruders, round)
  if round == 1 then
    if intruders then
      if store.side == "Humans" or store.side == "Intruders" then
        Script.DialogBox("ui/dialog/lvl1/Opening_Intruders.json")
      end
      intrudersSetup()     
    else
      if store.side == "Humans" or store.side == "Denizens" then
        Script.DialogBox("ui/dialog/lvl1/Opening_Denizens.json")
      end
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      if store.side == "Humans" or ((store.side == "Intruders") == intruders) then
        DoTutorials()
      end
    end

    Script.EndPlayerInteraction()

    return
  end

  if store.nFirstWaypointDown and not store.bSetup2Done then
    store.bSetup2Done = true
    -- denizensSetup()
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start2")
    placed = Script.PlaceEntities("Servitors_Start2", store.ServitorEnts, 0, ValueForReinforce())
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")    
  end

  if store.nSecondWaypointDown and not store.bSetup3Done then
    store.bSetup3Done = true
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start3")
    placed = Script.PlaceEntities("Servitors_Start3", store.ServitorEnts, 0, ValueForReinforce())
    print("DOODOODOO")
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
  end

  store.game = Script.SaveGameState()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder == intruders and not store.side == "Denizens" then
      Script.SelectEnt(ent)
      break
    end
  end


  Script.SetLosMode("intruders", "entities")
  Script.SetLosMode("denizens", "entities")
  if store.side == "Humans" then
    if intruders then
      Script.SetVisibility("intruders")
    else
      Script.SetVisibility("denizens")
    end
    Script.ShowMainBar(true)
  elseif store.side == "Intruders" then
    Script.ShowMainBar(intruders)
    Script.SetVisibility("intruders")
  elseif store.side == "Denizens" then
    Script.ShowMainBar(not intruders)
    Script.SetVisibility("denizens")
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

  return 6

  -- nTotalValueOnBoard = 0
  -- for _, ent in pairs(Script.GetAllEnts()) do
  --   for _, entValue in pairs(ents) do
  --     if ent.Name == entValue[1] then
  --       nTotalValueOnBoard = nTotalValueOnBoard + entValue[2]
  --     end 
  --   end
  -- end
  -- return ((6 - nTotalValueOnBoard) + store.nFirstWaypointDown + store.nSecondWaypointDown)
end

function OnMove(ent, path)
  print("Script: OnMove")
  -- for _, ent in pairs(Script.GetAllEnts()) do
  -- end

  return table.getn(path)
end

function makeMyRemoveExec(ent)
  return {script="remove", ent=ent}
end
function makeMySpawnExec(name, spawn_pattern, store_name)
  return {script="spawn", name=name, pat=spawn_pattern, store_name=store_name}
end
function doMyExec(exec)
  if exec.script == "remove" then
    print("SCRIPT: remove", exec.ent.Name)
    Script.RemoveEnt(exec.ent)
  elseif exec.script == "spawn" then
    print("SCRIPT: spawn", exec.name)
    sps = Script.GetSpawnPointsMatching(exec.pat)
    ent = Script.SpawnEntitySomewhereInSpawnPoints(exec.name, sps)
    store[exec.store_name] = ent
  else
    Script.DoExec(exec)
  end
end

function OnAction(intruders, round, exec)
  print("Script: OnAction")
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec
  if exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint1) <= 3 and not store.nFirstWaypointDown then
    --The intruders got to the first waypoint.
    store.nFirstWaypointDown = 2 --2 because that's what we want to add to the deni's deploy 
    remove_exec = makeMyRemoveExec(store.Waypoint1)
    store.execs[table.getn(store.execs) + 1] = remove_exec
    doMyExec(remove_exec)
    spawn_exec = makeMySpawnExec("Waypoint", "Waypoint2", "Waypoint2")
    store.execs[table.getn(store.execs) + 1] = spawn_exec
    doMyExec(spawn_exec)
    if store.side == "Humans" or store.side == "Intruders" then
      Script.DialogBox("ui/dialog/lvl1/First_Waypoint_Down_Intruders.json") 
    end
  end 
  
  if exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint2) <= 3 and not store.nSecondWaypointDown then
    --The intruders got to the second waypoint.
    store.nSecondWaypointDown = 2 --2 because that's what we want to add to the deni's deploy 

    remove_exec = makeMyRemoveExec(store.Waypoint2)
    store.execs[table.getn(store.execs) + 1] = remove_exec
    doMyExec(remove_exec)
    spawn_exec = makeMySpawnExec("Waypoint", "Waypoint3", "Waypoint3")
    store.execs[table.getn(store.execs) + 1] = spawn_exec
    doMyExec(spawn_exec)
    if store.side == "Humans" or store.side == "Intruders" then
      Script.DialogBox("ui/dialog/lvl1/Second_Waypoint_Down_Intruders.json")    
    end
  end  

  if exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint3) <= 3 then
    --The intruders got to the third waypoint.  Game over, man.  Game over.
    Script.DialogBox("ui/dialog/lvl1/Victory_Intruders.json")
  end   

  if not AnyIntrudersAlive() then
    Script.DialogBox("ui/dialog/lvl1/Victory_Denizens.json")
  end 

end
 

function RoundEnd(intruders, round)
  print("Script: RoundEnd")
  if round == 1 then
    return
  end

  bSkipOtherChecks = false  --Resets this every round

  if store.side == "Humans" then
    Script.ShowMainBar(false)
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
    if intruders then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end
  elseif store.side == "Denizens" then
  elseif store.side == "Intruders" then
  end

  if store.side == "Humans" then
    if intruders then
      Script.DialogBox("ui/dialog/lvl1/pass_to_denizens.json")
    else
      Script.DialogBox("ui/dialog/lvl1/pass_to_intruders.json")
    end
  end

  if intruders then
    if store.side == "Humans" or store.side == "Intruders" then
      if store.nFirstWaypointDown and not bShowedFirstWaypointMessage then
        bShowedFirstWaypointMessage = true
        Script.DialogBox("ui/dialog/lvl1/First_Waypoint_Down_Denizens.json")
      end

      if store.nSecondWaypointDown and not bShowedSecondWaypointMessage then
        bShowedSecondWaypointMessage = true
        Script.DialogBox("ui/dialog/lvl1/Second_Waypoint_Down_Denizens.json")
      end
    end
  else
    if store.side == "Humans" or store.side == "Denizens" then
      if bCountdownTriggered then
        nCountdown = nCountdown - 1
      end
    end
  end

  if store.side == "Humans" then
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.LoadGameState(store.game)
    for _, exec in pairs(store.execs) do
      doMyExec(exec)
    end
    store.execs = {}
  end

  --if the Master is down, respawn him
  if not MasterIsAlive() then
    master_spawn = Script.GetSpawnPointsMatching("Master_Start")
    Script.SpawnEntitySomewhereInSpawnPoints(store.MasterName, master_spawn)    
  end

end

function MasterIsAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Bosch" and ent.HpCur > 0 then
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
