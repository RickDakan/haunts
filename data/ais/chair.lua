function Think()
  target = retaliate()
  if target ~= nil and Utils.RangedDistBetweenEntities (Me, target) <2 then
    moveWithinRangeAndAttack(1, "Smash", target)
  else
    target = pursue()
    if target == nil then
      target = targetAllyAttacker()
    end
    if target == nil then
      target = targetAllyTarget()
    end
    if target == nil then
      target = targetLowestStat("HpCur")
    end
    if target == nil then
      target = nearest()
    end 
    if target == nil then
      return
    end
    moveWithinRangeAndAttack (1, "Smash", target)
  end
end
