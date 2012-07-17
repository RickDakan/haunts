--dialog_test.lua

function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function Init()
	store.ch1 = {}
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
	Script.SaveStore()
	
	Script.BindAi("intruder", "human")
	Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
	

function doIntrudersSetup()
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")
  Script.SpawnEntitySomewhereInSpawnPoints("Teen", intruder_spawn)
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Occultist", intruder_spawn)
  Script.BindAi(ent, "sample_aoe_occultist.lua")
  Script.SpawnEntitySomewhereInSpawnPoints("Ghost Hunter", intruder_spawn)
  ents = Script.GetAllEnts()
end


function RoundStart(intruders, round)
    Script.SetVisibility("intruders")
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

end
function OnMove(ent, path)
  for _, dialog in pairs(store.Ch01.Dialogs) do
    if not store.Ch01.Dialog_complete[dialog] then
      for i, pos in pairs(path) do
        if pointIsInSpawns(pos, dialog) then
          store.Ch01.Dialog_complete[dialog] = "DOIT"
          return i
        end
      end
    end
  end
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- DIALOG
  -- Check for that value "DOIT", if we find it then we pop up the dialog box.
  for _, dialog in pairs(store.Ch01.Dialogs) do
 --   if not store.Ch1.Dialog_complete[dialog] then
 
 --     return
 --   end
    if store.Ch01.Dialog_complete[dialog] == "DOIT" then
      -- Note that the .. operator is string concatenation in lua.
      dialog_path = "ui/dialog/ch1/" .. dialog .. ".json"
      Script.DialogBox(dialog_path)  -- pop up the dialog box
      store.Ch1.Dialog_complete[dialog] = true
      -- keep track so we don't do it again later
      return
    end
  end
end

