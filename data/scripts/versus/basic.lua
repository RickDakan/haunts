function Init(data)
<<<<<<< HEAD
 level_choices = Script.ChooserFromFile("ui/start/versus/map_select.json")
  Script.LoadHouse("Lvl_01_Haunted_House") 
=======
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
>>>>>>> Jonathan/devel
end

function RoundStart(intruders, round)
  Script.StartScript(level_choices[1])
end
