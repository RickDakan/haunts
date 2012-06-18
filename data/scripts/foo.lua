function Init()
  map = selectMap()
  loadHouse(map)
  spawnDude("Angry Shade", 20, 20)
  spawnDude("Teen", 10, 20)

  -- This creates a list of entities and associated point values.  The order
  -- the names are listed in here is the order they will appear to the user.
  ents = {
    {"Angry Shade", 1},
    {"Disciple", 2},
    {"Devotee", 2},
    {"Eidolon", 3},
  }
  -- Now we give the user a ui with which to place these entities.  This user
  -- will have 10 points to spend and can only place the entities in spawn
  -- points with names that begin with "Foo-"
  dudes = placeDude("Foo-.*", 10, ents)

  for i, dude in pairs(dudes) do
    print(i, dude)
  end
  showMainBar(true)
  ents = getAllEnts()
end

function OnRound(intruders)
  print("Intruders", intruders)
end