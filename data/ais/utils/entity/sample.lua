function moveToAndMeleeAttack(attack,target)
  print(Me)
  print(target)
  ps = allPathablePoints(Me.Pos, pos(target), 1, 1)
  res = doMove(ps, 1000)
  if res == nil then
    return nil
  end
  return doBasicAttack(attack, target)
end
