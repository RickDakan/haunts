function Think()
  denizens = AllDenizens()
  for _, ent in pairs(denizens) do
    while IsActive(ent) do
      ExecDenizen(ent)
    end
  end
  return false
end
