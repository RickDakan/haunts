
function Think()
  intruders = nearestNEntities(3, "intruder")

  stats = Me.
  stats = getAoeAttackStats(Me, "Abjuration")
  mystats = getEntityStats(Me)
  if intruders[1] then
    movement = mystats.apCur - stats.ap
    if movement < 0 then
      movement = 0
    end
    print("ooking for bst target")
    target = bestAoeAttackPos("Abjuration", movement, "enemies only")
    print("range", stats.range)
    print("target", target.x, target.y)
    if not (target.x == 0 and target.y == 0) then
      ps = allPathablePoints(Me.Pos, target, 1, stats.range)
      res = doMove(ps, 1000)
      print("taget", target.x, target.y)
      res = doAoeAttack("Abjuration", target)
      if res == nil then
        return nil
      end
    else
      ps = allPathablePoints(Me.Pos, pos(intruders[1]), stats.range, stats.range)
      doMove(ps, 1000)
      return
    end
  else
    return
  end
end
