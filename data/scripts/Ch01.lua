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

  Script.LoadHouse("Chapter_01_a")
  Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog01.json") 

  sp = Script.GetSpawnPointsMatching("Foo-.*")

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
 
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")
  Script.SpawnEntitySomewhereInSpawnPoints("Caitlin", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Percy", intruder_spawn)
  ents = Script.GetAllEnts()
  end
end
 

function RoundStart(intruders, round)
    Script.SetVisibility("intruders")
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.ShowMainBar(intruders ~= play_as_denizens)
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