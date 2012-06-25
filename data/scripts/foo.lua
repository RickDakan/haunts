function setLosModeToRoomsWithSpawnsMatching(pattern)
  sp = getSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = roomAtPos(spawn)
  end
  setLosMode("rooms", rooms)
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
function Init()
  map = selectMap()
  loadHouse(map)
  bindAi("denizen", "denizens.lua")
  bindAi("minions", "minions.lua")
  bindAi("intruder", "human")
end

function doDenizenSetup()
  -- This creates a list of entities and associated point values.  The order
  -- the names are listed in here is the order they will appear to the user.
  ents = {
    {"Chosen One", 1},
    {"Master of the Manse", 1},
  }

  setLosModeToRoomsWithSpawnsMatching("Master-.*")

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
  if placed[1] == "Chosen One" then
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
  setLosModeToRoomsWithSpawnsMatching("Servitor-.*")
  placed = placeEntities("Servitor-.*", 10, ents)

  setLosMode("none")
end

function doIntrudersSetup()
  modes = {}
  modes["Cleanse"] = "ui/explorer_setup/cleanse.png"
  modes["Relic"] = "ui/explorer_setup/relic.png"
  modes["Mystery"] = "ui/explorer_setup/mystery.png"
  r = pickFromN(1, 1, modes)
  for i,name in pairs(r) do
    print("picked", i, name)
  end

  intruder_spawn = getSpawnPointsMatching("Intruders-FrontDoor")
  spawnEntitySomewhereInSpawnPoints("Teen", intruder_spawn)
  spawnEntitySomewhereInSpawnPoints("Occultist", intruder_spawn)
  spawnEntitySomewhereInSpawnPoints("Ghost Hunter", intruder_spawn)

  ents = getAllEnts()

  for en, ent in pairs(ents) do
    print("Checking ent: ", ent.name)
    for i, gear in pairs(ent.gear) do
      print("gear", en, i, gear)
    end
    count = 0
    for _, _ in pairs(ent.gear) do
      count = count + 1
    end
    if count > 0 then
      r = pickFromN(1, 1, ent.gear)
      for i,name in pairs(r) do
        print("picked", i, name)
      end
    end
  end

  setLosMode("entities")
  showMainBar(true)
end

function RoundStart(intruders, turn)
  if turn == 1 then
    if intruders then
      doIntrudersSetup()
    else
      doDenizenSetup()
    end
    return
  end
  print("start", intruders, turn)
  -- if not intruders then
  --   spawn_points = getSpawnPointsMatching("Minion-.*")
  --   p = spawnEntitySomewhereInSpawnPoints("Angry Shade", spawn_points)
  --   print("Spawned Angry Shade at ", p.x, p.y)
  -- end
end

function RoundEnd(intruders, turn)
  print("end", intruders, turn)
end
