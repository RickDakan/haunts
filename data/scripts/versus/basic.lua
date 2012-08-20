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

  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")
--  Script.PlayMusic("Haunts/Music/Adaptive/Bed 2")
--  Script.SetMusicParam("tension_level", 0.1)

  -- check data.map == "random" or something else
  Script.LoadHouse("Lvl_02_Basement_Lab")

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

  relic_spawn = Script.GetSpawnPointsMatching("Relic_Spawn")
  Relic = Script.SpawnEntitySomewhereInSpawnPoints("Relic", relic_spawn)
  Script.SelectEnt(Relic)
  --Sets the length of time the intruders have to get the master to the relic after the relic has been triggered.
  nCountdown = 5
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
    ents = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  end
  if placed[1].Name == "Bosch" then
    MasterName = "Bosch"
    ents = {
      {"Angry Shade", 1},
      {"Lost Soul", 1},
      {"Vengeful Wraith", 2},
     }  
  end


  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_.*")
  placed = Script.PlaceEntities("Servitors_.*", ents, 0, 10)
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      intrudersSetup()     
    else
      Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Denizens.json")
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      DoTutorials()
    end

    --Script.SetCondition (MasterName, "Lumbering", true)
    Script.EndPlayerInteraction()

    return
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

function OnMove(ent, path)
  -- for _, ent in pairs(Script.GetAllEnts()) do
  --   if ent.Name == "Relic" then
  --     Relic = ent
  --     break
  --   end
  -- end

  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec
 
  if exec.Action.Type == "Basic Attack" then
    if exec.Target.Name == MasterName and exec.Target.Hp <= 0 then
      --master is dead.  Intruders win.
      Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Victory_Intruders.json")
    end
  end

  if  exec.Ent.Side == "Intruder" and GetDistanceBetweenEnts(exec.Ent, Relic) <= 3 and not bCountdownTriggered then
    --The intruders got to the relic before the master.  They win.
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Victory_Intruders.json")
  end 

  if exec.Ent.Name == MasterName and GetDistanceBetweenEnts(exec.Ent, Relic) <= 3 and not bCountdownTriggered then
     bCountdownTriggered = true
     Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Countdown_Started_Denizens.json")
  end  
end
 

function RoundEnd(intruders, round)
  if round == 1 then
    return
  end

  bSkipOtherChecks = false  --Resets this every round

  if nCountdown == 0 then
    --game over, the denizens win.
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Victory_Denizens.json")
  end

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
      if bCountdownTriggered then
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Turns_Remaining_Denizens.json")
      else
        Script.DialogBox("ui/dialog/Lvl02/pass_to_denizens.json")
      end
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl02/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if bCountdownTriggered and not bShowedIntruderTimerMessage and not bSkipOtherChecks then
        bShowedIntruderTimerMessage = true
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Countdown_Started_Intruder.json")
        bSkipOtherChecks = true
      end

      if bCountdownTriggered and not bSkipOtherChecks then  --timer is triggered and we've already intro'd it
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Turns_Remaining_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl02/pass_to_intruders.json")
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
end