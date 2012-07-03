function think()
  -- This is just to check that there is someone around that we can see
  denizens = nearestNEntities (1, "denizen")
  if table.getn(denizens) == 0 then
    return
  end

  me_stats = getEntityStats(me())
  abj_stats = getAoeAttackStats(me(), "Abjuration")

  -- gz, or 'ground zero', is where we're going to center our aoe
  gz = bestAoeAttackPos("Abjuration", me_stats.apCur - abj_stats.ap, "enemies only")

  -- find all positions from which we could center our aoe on gz
  dsts = allPathablePoints(pos(me()), gz, 1, abj_stats.range)

  -- move to any one of the closest positions in dsts
  doMove(dsts, 1000)

  -- if we're still out of range then we'll just have to try again next turn
  if rangedDistBetweenPositions(pos(me()), gz) > abj_stats.range then
    return
  else
    doAoeAttack("Abjuration", gz)
  end

  -- More attacks if possible
  think()
end
think()