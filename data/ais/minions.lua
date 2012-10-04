function Think()
  print("A")
  minions = AllMinions()
  print("B")
  for _, minion in pairs(minions) do
  print("C", minion.Name)
    while IsActive(minion) do
  print("D")
      ExecMinion(minion)
  print("E")
    end
  print("F")
  end
end
