function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = getSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = roomAtPos(spawn)
  end
  setLosMode(side, "rooms", rooms)
end
-- Need the following functions
-- setLos("entities") -> sets los to all entities on the current side
-- setLos("rooms")
-- setLos("all")
-- endTurn()
-- spawnObjects -> maybe tag them with a string, like "relic"
-- 
-- need an OnAction() function that is called after every action is
-- completed.
play_as_denizens = true
function Init()
  while true do
    dialogBox("ui/dialog/sample.json")
  end
  map = selectMap()
  loadHouse(map)


  if play_as_denizens then
    bindAi("denizen", "human")
    bindAi("minions", "minions.lua")
    bindAi("intruder", "intruders.lua")
  else
    bindAi("denizen", "denizens.lua")
    bindAi("minions", "minions.lua")
    bindAi("intruder", "human")
  end
end

function doDenizenSetup()
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
    placed = placeEntities("Master-.*", 1, ents)
  end

  -- placed is an array containing all of the entities placed, in this case
  -- there will only be one, and we will use that one to determine what
  -- servitors to make available to the user to place.
  if placed[1].name == "Chosen One" then
    ents = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  else
    ents = {
      {"Corpse", 1},
      {"Vengeful Wraith", 1},
      {"Angry Shade", 1},
    }
  end

  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor-".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitor-.*")
  placed = placeEntities("Servitor-.*", 10, ents)

  -- set the denizens to not be able to see anything - it's not very
  -- significant since their turn is about to end anyway.
  setLosMode("denizens", "none")
end

function doIntrudersSetup()
  -- Let the intruders choose from among one of the three game modes.
  -- Currently the data isn't stored or used anywhere, but this gives you an
  -- idea of how a menu can be created.
  modes = {}
  modes["Cleanse"] = "ui/explorer_setup/cleanse.png"
  modes["Relic"] = "ui/explorer_setup/relic.png"
  modes["Mystery"] = "ui/explorer_setup/mystery.png"
  r = pickFromN(1, 1, modes)
  for i,name in pairs(r) do
    print("picked", i, name)
  end

  -- Find the "Intruders-FrontDoor" spawn point and spawn a Teen, Occultist,
  -- and Ghost Hunter there.  Additionally we will mind the
  -- sample_aoe_occultist.lua ai to the occultist.
  intruder_spawn = getSpawnPointsMatching("Intruders-FrontDoor")
  spawnEntitySomewhereInSpawnPoints("Teen", intruder_spawn)
  ent = spawnEntitySomewhereInSpawnPoints("Occultist", intruder_spawn)
  bindAi(ent, "sample_aoe_occultist.lua")
  spawnEntitySomewhereInSpawnPoints("Ghost Hunter", intruder_spawn)
  ents = getAllEnts()

  -- This lets you pick gear for each entity, you can uncomment this block
  -- if you want to turn it on.
  -- for en, ent in pairs(ents) do
  --   print("Checking ent: ", ent.name)
  --   for i, gear in pairs(ent.gear) do
  --     print("gear", en, i, gear)
  --   end
  --   count = 0
  --   for _, _ in pairs(ent.gear) do
  --     count = count + 1
  --   end
  --   if count > 0 then
  --     r = pickFromN(1, 1, ent.gear)
  --     for i,name in pairs(r) do
  --       print("picked", i, name)
  --     end
  --     setGear(ent, r[1])
  --   end
  -- end
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      setVisibility("intruders")
      doIntrudersSetup()
    else
      setVisibility("denizens")
      doDenizenSetup()
    end
    -- Make it clear that the players don't get to activate their units on the
    -- setup turn.
    endPlayerInteraction()
    return
  end

  -- The start of round 2 is when a player's visibility should be only what
  -- their units are able to see.  It is important to call setVisibility
  -- because it sets what the player can see, and it is important to call
  -- setLosMode because it sets what is technically visible to entities on
  -- each side.
  if round == 2 then
    if play_as_denizens then
      setVisibility("denizens")
    else
      setVisibility("intruders")
    end
    setLosMode("intruders", "entities")
    setLosMode("denizens", "entities")
  end
  showMainBar(intruders ~= play_as_denizens)

  -- This is sample code to spawn one angry shade at the start of each
  -- denizens' turn.
  -- if not intruders then
  --   spawn_points = getSpawnPointsMatching("Minion-.*")
  --   p = spawnEntitySomewhereInSpawnPoints("Angry Shade", spawn_points)
  --   print("Spawned Angry Shade at ", p.x, p.y)
  -- end
end

function RoundEnd(intruders, round)
  print("end", intruders, round)
end
