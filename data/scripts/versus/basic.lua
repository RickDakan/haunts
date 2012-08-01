function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function Init(data)
  side_choices = Script.ChooserFromFile("ui/start/versus/side.json")

  -- check data.map == "random" or something else
  Script.LoadHouse("versus-1")

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

  -- Script.SetLosMode("intruders", "entities")
  -- Script.SetLosMode("denizens", "entities")
  -- if store.side == "Intruders" then
  --   Script.SetVisibility("intruders")
  -- end
  -- if store.side == "Denizens" then
  --   Script.SetVisibility("denizens")
  -- end
  -- Script.ShowMainBar(true)
end

function intrudersSetup()
  store.goal = Script.ChooserFromFile("ui/start/versus/goals.json")
  if store.goal == "Cleanse" then
  end
  if store.goal == "Mystery" then
  end
  if store.goal == "Relic" then
  end

  intruder_names = {"Teen", "Occultist", "Ghost Hunter"}
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn)
    Script.SetGear(ent, "Pre-loaded Playlist")
    -- PRETEND!
  end

  -- Choose entry point here.

  Script.SaveStore()
end

function denizensSetup()
  -- This creates a list of entities and associated point values.  The order
  -- the names are listed in here is the order they will appear to the user.
  ents = {
    {"Chosen One", 1},
    {"Master of the Manse", 1},
  }

  setLosModeToRoomsWithSpawnsMatching("denizens", "Master-.*")

  -- Now we give the user a ui with which to place these entities.  The user
  -- will have 1 point to spend, and each of the options costs one point, so
  -- they will only place 1.  We will make sure they place exactly one.
  -- Also the "Master-.*" indicates that the entity can only be placed in
  -- spawn points that have a name that matches the regular expression
  -- "Master-.*", which means anything that begins with "Master-".  So
  -- "Master-BackRoom" and "Master-Center" both match, for example.
  placed = {}
  while table.getn(placed) == 0 do
    placed = Script.PlaceEntities("Master-.*", 1, ents)
  end

  -- placed is an array containing all of the entities placed, in this case
  -- there will only be one, and we will use that one to determine what
  -- servitors to make available to the user to place.
  if placed[1].Name == "Chosen One" then
    ents = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  else
    ents = {
      {"Corpse", 1},
      {"Vengeful Wraith", 1},
      {"Poltergeist", 1},
      {"Angry Shade", 1},
    }
  end

  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor-".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors-.*")
  placed = Script.PlaceEntities("Servitors-.*", 10, ents)
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      intrudersSetup()
    else
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
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


function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec
end
 

function RoundEnd(intruders, round)
  if round == 1 then
    return
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
      Script.DialogBox("ui/start/versus/pass_to_denizens.json")
    else
      Script.DialogBox("ui/start/versus/pass_to_intruders.json")
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

