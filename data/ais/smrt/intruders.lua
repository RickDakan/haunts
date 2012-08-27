function Think()
  intruders = activeIntruders()
  for _, ent in pairs(intruders) do
    SetEntityMasterInfo(ent, "Name", ent.Name)
    if ent.Name == "Teen" then
      SetEntityMasterInfo(ent, "Leader", "true")
    end
  end
  while intruders[1] do
    execIntruder(intruders[1])
    intruders = activeIntruders()
  end
end
