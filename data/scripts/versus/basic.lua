function Init(data)
  -- level_choices = Script.ChooserFromFile("ui/start/versus/map_select.json")
  Script.LoadHouse("Lvl_01_Haunted_House") 
end

function RoundStart(intruders, round)
  Script.StartScript("Lvl04.lua")
  -- Script.StartScript(level_choices[1])
end
