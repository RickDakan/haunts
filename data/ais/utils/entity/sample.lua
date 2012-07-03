function moveToAndMeleeAttack(attack,target)
  print(me())
  print(target)
  ps = allPathablePoints(pos(me()), pos(target), 1, 1)
  res = doMove(ps, 1000)
  if res == nil then
    return nil
  end
  return doBasicAttack(attack, target)
end
