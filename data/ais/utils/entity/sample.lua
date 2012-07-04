function moveToAndMe()leeAttack(attack,target)
  print(Me())
  print(target)
  ps = AllPathablePoints(Me().Pos, target.Pos, 1, 1)
  res = DoMove(ps, 1000)
  if res == nil then
    return nil
  end
  return DoBasicAttack(attack, target)
end
