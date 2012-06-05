while true do
  intruders = nearestNEntities(3, "intruder")
  if intruders[1] then
    mypos = pos(me())
    hispos = pos(intruders[1])
    ps = allPathablePoints(mypos, hispos, 2, 4)
    index = 1
    while ps[index] do
      print(ps[index])
      index = index + 1
    end
    res = doBasicAttack("Chill Touch", intruders[1])
  end
end
