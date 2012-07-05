function moveToAndMeleeAttack(attack,target)
  print(Me)
  print(target)
  ps = Utils.AllPathablePoints(Me.Pos, target.Pos, 1, 1)
  res = Actions.Move(ps, 1000)
  if res == nil then
    return nil
  end
  return Actions.BasicAttack(attack, target)
end
