-- wisp new
-- look at moving away from everyone


function Think()


  shrieker_trigger = Script.GetSpawnPointsMatching("Trigger_Shrieker")
  if not store.Ch05b.Spawnpoints_complete[shrieker_trigger] then
    --the shrieker has not yet been triggered.  Chill.
    break
  end

--Doesn't work... just gonna have him disappear and reappear.
  -- shrieker_trigger = Script.GetSpawnPointsMatching("Trigger_Shrieker")
  -- boss_trigger = Script.GetSpawnPointsMatching("Boss_Start")
  -- if store.Ch05b.Spawnpoints_complete[shrieker_trigger] and not store.Ch05b.Spawnpoints_complete[boss_trigger] then
  --   --the shrieker has been triggered, and the boss encounter has not.  Move toward the boss room.
  --   dsts[0].X = boss_trigger[0].Pos.X
  --   dsts[0].Y = boss_trigger[0].Pos.Y
  --   Do.Move(dsts, 1000)
  -- end 




  intruder = Utils.NearestNEntities (1, "intruder")[1]
  if Utils.RangedDistBetweenEntities (Me, intruder) <3 then
    moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
  else 
    moveWithinRangeAndAttack (4, "Ectoplasmic Discharge", intruder)
  end
end



