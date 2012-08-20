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

  store.side = side_choices[1]
  if store.side == "Humans" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end
  if store.side == "Denizens" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "intruders.lua")
  end
  if store.side == "Intruders" then
    Script.BindAi("denizen", "denizens.lua")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end

  waypoint_spawn = Script.GetSpawnPointsMatching("Waypoint1")
  Waypoint1 = Script.SpawnEntitySomewhereInSpawnPoints("Waypoint", waypoint_spawn)

  nFirstWaypointDown = false
  nSecondWaypointDown = false
end

function intrudersSetup()

  if IsStoryMode() then
    intruder_names = {"Teen", "Occultist", "Researcher"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  -- else
  --   --permit all choices for normal vs play
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn)
    
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
    placed = Script.PlaceEntities("Master_.*", ents, 1, 1)
  end

  -- placed is an array containing all of the entities placed, in this case
  -- there will only be one, and we will use that one to determine what
  -- servitors to make available to the user to place.
  if placed[1].Name == "Chosen One" then
    MasterName = "Chosen One"
    ServitorEnts = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  end
  if placed[1].Name == "Bosch" then
    MasterName = "Bosch"
    ServitorEnts = {
      {"Angry Shade", 1},
      {"Lost Soul", 1},
    }  
  end
--      {"Vengeful Wraith", 3},

  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start1")
  placed = Script.PlaceEntities("Servitors_Start1", ServitorEnts, 0,6)
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      intrudersSetup()     
    else
      Script.DialogBox("ui/dialog/lvl1/Opening_Denizens.json")
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      DoTutorials()
    end

    Script.EndPlayerInteraction()

    return
  end

  if nFirstWaypointDown and not bSetup2Done then
    bSetup2Done = true
    --denizensSetup()
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start2")
    placed = Script.PlaceEntities("Servitors_Start2", ServitorEnts, 0, ValueForReinforce())
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")    
  end

  if nSecondWaypointDown and not bSetup3Done then
    bSetup3Done = true
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start3")
    placed = Script.PlaceEntities("Servitors_Start3", ServitorEnts, 0, ValueForReinforce())
    print("DOODOODOO")
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
  end

  store.game = Script.SaveGameState()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder == intruders then
      Script.SelectEnt(ent)
      break
    end
  end
  if store.side == "Humans" then
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("intruders")
    else
      Script.SetVisibility("denizens")
    end
    Script.ShowMainBar(true)
  else
    Script.ShowMainBar(intruders == (store.side == "Intruders"))
  end
end

function GetDistanceBetweenEnts(ent1, ent2)
  return (math.abs(ent1.Pos.X - ent2.Pos.X) + math.abs(ent1.Pos.Y - ent2.Pos.Y))
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
  -- return ((6 - nTotalValueOnBoard) + nFirstWaypointDown + nSecondWaypointDown)
end

function OnMove(ent, path)
  -- for _, ent in pairs(Script.GetAllEnts()) do
  -- end

  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec

  if  exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, Waypoint1) <= 3 and not nFirstWaypointDown then
    --The intruders got to the first waypoint.
    nFirstWaypointDown = 2 --2 because that's what we want to add to the deni's deploy 
    Script.SetHp(Waypoint1, 0)
    waypoint_spawn = Script.GetSpawnPointsMatching("Waypoint2")
    Waypoint2 = Script.SpawnEntitySomewhereInSpawnPoints("Waypoint", waypoint_spawn)
    Script.DialogBox("ui/dialog/lvl1/First_Waypoint_Down_Intruders.json") 
  end 

  if  exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, Waypoint2) <= 3 and not nSecondWaypointDown then
    --The intruders got to the second waypoint.
    nSecondWaypointDown = 2 --2 because that's what we want to add to the deni's deploy 
    Script.SetHp(Waypoint2, 0)
    waypoint_spawn = Script.GetSpawnPointsMatching("Waypoint3")
    Waypoint3 = Script.SpawnEntitySomewhereInSpawnPoints("Waypoint", waypoint_spawn)
    Script.DialogBox("ui/dialog/lvl1/Second_Waypoint_Down_Intruders.json")    
  end  

  if  exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, Waypoint3) <= 3 then
    --The intruders got to the third waypoint.  Game over, man.  Game over.
    Script.DialogBox("ui/dialog/lvl1/Victory_Intruders.json")
  end   

  if not AnyIntrudersAlive() then
    Script.DialogBox("ui/dialog/lvl1/Victory_Denizens.json")
  end 

end
 

function RoundEnd(intruders, round)
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

    if intruders then
      Script.DialogBox("ui/dialog/lvl1/pass_to_denizens.json")
      if nFirstWaypointDown and not bShowedFirstWaypointMessage then
        bShowedFirstWaypointMessage = true
        Script.DialogBox("ui/dialog/lvl1/First_Waypoint_Down_Denizens.json")
      end

      if nSecondWaypointDown and not bShowedSecondWaypointMessage then
        bShowedSecondWaypointMessage = true
        Script.DialogBox("ui/dialog/lvl1/Second_Waypoint_Down_Denizens.json")
      end
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/lvl1/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/lvl1/Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/lvl1/pass_to_intruders.json")
      end

      if bCountdownTriggered then
        nCountdown = nCountdown - 1
      end
    end

    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.LoadGameState(store.game)
    for _, exec in pairs(store.execs) do
      Script.DoExec(exec)
    end
    store.execs = {}
  end

  --if the Master is down, respawn him
  if not MasterIsAlive() then
    master_spawn = Script.GetSpawnPointsMatching("Master_Start")
    Script.SpawnEntitySomewhereInSpawnPoints(MasterName, master_spawn)    
  end

end

function MasterIsAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == MasterName and ent.HpCur > 0 then
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
