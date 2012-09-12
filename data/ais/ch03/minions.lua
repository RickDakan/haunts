function Think()
  minions = AllDenizens()
  for _, ent in pairs(minions) do
    while IsActive(ent) do
      ExecMinions(ent)
    end
  end
  return false
end
