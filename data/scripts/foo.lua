function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

play_as_denizens = false
function Init()
   store.Ch01 = {}
   store.Ch01.Dialog_complete = {}
   store.Ch01.Dialogs = {
      "Ch01_Dialog01",
      "Ch01_Dialog02",
      "Ch01_Dialog03",
      "Ch01_Dialog04",
      "Ch01_Dialog05",
      "Ch01_Dialog06",
      "Ch01_Dialog07",
      "Ch01_Dialog08",
      "Ch01_Dialog09",
      "Ch01_Dialog10",
   } 

while false do
    choices = Script.DialogBox("ui/dialog/sample.json")
    print("Choices made:")
    for _, choice in pairs(choices) do
      print("Chose: ", choice)
    end
  end


  Script.LoadHouse("dialog_test")
   
  sp = Script.GetSpawnPointsMatching("Foo-.*")

  if play_as_denizens then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "intruders.lua")
  else
    Script.BindAi("denizen", "denizens.lua")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
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
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitor-.*")
  placed = Script.PlaceEntities("Servitor-.*", 10, ents)

  -- set the denizens to not be able to see anything - it's not very
  -- significant since their turn is about to end anyway.
  Script.SetLosMode("denizens", "none")
end

function doIntrudersSetup()
  -- Let the intruders choose from among one of the three game modes.
  -- Currently the data isn't stored or used anywhere, but this gives you an
  -- idea of how a menu can be created.
  -- modes = {}
  -- modes["Cleanse"] = "ui/explorer_setup/cleanse.png"
  -- modes["Relic"] = "ui/explorer_setup/relic.png"
  -- modes["Mystery"] = "ui/explorer_setup/mystery.png"
  -- print("Last time picked: ",   store.mode)
  -- r = Script.PickFromN(1, 1, modes)
  -- for i,name in pairs(r) do
  --   print("This time picked: ", i, name)
  -- end
  -- store.mode = r[1]
  -- Script.SaveStore()
  -- -- Find the "Intruders-FrontDoor" spawn point and spawn a Teen, Occultist,
  -- -- and Ghost Hunter there.  Additionally we will mind the
  -- -- sample_aoe_occultist.lua ai to the occultist.
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")
  -- print("intruder spawn:", intruder_spawn)
  -- for k,v in pairs(intruder_spawn) do
  --   print("kv: ", k, v)
  -- end
  Script.SpawnEntitySomewhereInSpawnPoints("Teen", intruder_spawn)
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Occultist", intruder_spawn)
  Script.BindAi(ent, "sample_aoe_occultist.lua")
  Script.SpawnEntitySomewhereInSpawnPoints("Ghost Hunter", intruder_spawn)
  ents = Script.GetAllEnts()

  -- This lets you pick gear for each entity, you can uncomment this block
  -- if you want to turn it on.
  -- for en, ent in pairs(ents) do
  --   print("Checking ent: ", ent.Name)
  --   for i, gear in pairs(ent.GearOptions) do
  --     print("gear", en, i, gear)
  --   end
  --   count = 0
  --   for _, _ in pairs(ent.GearOptions) do
  --     count = count + 1
  --   end
  --   if count > 0 then
  --     r = Script.PickFromN(1, 1, ent.GearOptions)
  --     for i,name in pairs(r) do
  --       print("picked", i, name)
  --     end
  --     Script.SetGear(ent, r[1])
  --   end
  -- end
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      Script.SetVisibility("intruders")
      doIntrudersSetup()

      relic_spawn = Script.GetSpawnPointsMatching("Relic-.*")
      Script.SpawnEntitySomewhereInSpawnPoints("Scepter", relic_spawn)
      Script.SpawnEntitySomewhereInSpawnPoints("Tome", relic_spawn)
    else
      -- Script.SetVisibility("denizens")
      -- doDenizenSetup()
    end
    -- Make it clear that the players don't get to activate their units on the
    -- setup turn.
    Script.EndPlayerInteraction()
    return
  end

  -- The start of round 2 is when a player's visibility should be only what
  -- their units are able to see.  It is important to call setVisibility
  -- because it sets what the player can see, and it is important to call
  -- setLosMode because it sets what is technically visible to entities on
  -- each side.
  if round == 2 then
    if play_as_denizens then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
  end
  Script.ShowMainBar(intruders ~= play_as_denizens)

  -- This is sample code to spawn one angry shade at the start of each
  -- denizens' turn.
  -- if not intruders then
  --   spawn_points = Script.GetSpawnPointsMatching("Minion-.*")
  --   p = Script.SpawnEntitySomewhereInSpawnPoints("Angry Shade", spawn_points)
  --   print("Spawned Angry Shade at ", p.x, p.y)
  -- end
end


function pointIsInSpawn(pos, sp)
  return pos.X >= sp.Pos.X and pos.X < sp.Pos.X + sp.Dims.Dx and pos.Y >= sp.Pos.Y and pos.Y < sp.Pos.Y + sp.Dims.Dy
end

function pointIsInSpawns(pos, regexp)
  print("Regexp: '",regexp,"'")
  sps = Script.GetSpawnPointsMatching(regexp)
  print("Got ", "spawns for ", regexp)
  for _, sp in pairs(sps) do
    if pointIsInSpawn(pos, sp) then
      return true
    end
  end
  return false
end

function OnMove(ent, path)
  -- DIALOG
  -- We want to check if an intruder is about to walk through one of the
  -- waypoints we've set up.  Ideally we'd check first that ent.Side.Intruder
  -- is true, but that isn't working in this build (I've fixed it already though).
  -- path is an array of the points that the entity is about to walk through, so
  -- we check them in order to see if any overlap the next waypoint that should
  -- trigger.  If it does then we set the corresponding value in Dialog_complete
  -- to "DOIT", which is our signal to pop up the dialog box when the action
  -- completes.
  for _, dialog in pairs(store.Ch01.Dialogs) do
    -- We don't want to truncate movement for dialog that has already happened,
    -- so we check Dialog_complete before anything else.
    if not store.Ch01.Dialog_complete[dialog] then
      for i, pos in pairs(path) do
        if pointIsInSpawns(pos, dialog) then
          -- If we need to trigger the dialog then we set the appropriate value
          -- to "DOIT" and return the distance along this path that the ent
          -- should move.
          store.Ch01.Dialog_complete[dialog] = "DOIT"
          return i
        end
      end
    end
  end

  -- If we made it here then there was no dialog that needed to be shown, so
  -- we don't want to truncate the movement action, so we return the length
  -- of the path.
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- DIALOG
  -- Check for that value "DOIT", if we find it then we pop up the dialog box.
  for _, dialog in pairs(store.Ch01.Dialogs) do
   --  if not store.Ch01.Dialog_complete[dialog] then
   --   return
   -- end
    if store.Ch01.Dialog_complete[dialog] == "DOIT" then
      -- Note that the .. operator is string concatenation in lua.
      dialog_path = "ui/dialog/Ch01/" .. dialog .. ".json"
      Script.DialogBox(dialog_path)  -- pop up the dialog box
      store.Ch01.Dialog_complete[dialog] = true
      -- keep track so we don't do it again later
      return
    end
  end
end

function RoundEnd(intruders, round)
  print("end", intruders, round)
end
