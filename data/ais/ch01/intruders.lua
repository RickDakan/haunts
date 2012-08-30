function GetLeader(intruders, leader_order)
  for _, name in pairs(leader_order) do
    for _, ent in pairs(intruders) do
      if ent.Name == name then
        return ent
      end
    end
  end
  return nil
end

function Think()
  intruders = AllIntruders()
  leader = GetLeader(intruders, {"Occultist", "Researcher", "Teen"})
  for _, ent in pairs(intruders) do
    SetEntityMasterInfo(ent, "Name", ent.Name)
    SetEntityMasterInfo(ent, "Leader", leader.Name)
  end
  print("Master: leader is", leader.Name)
  while IsActive(leader) do
    print("Master: exec", leader.Name)
    ExecIntruder(leader)
  end
  for _, ent in pairs(intruders) do
    while IsActive(ent) do
      print("Master: exec", ent.Name)
      ExecIntruder(ent)
      print("SCRIPT: Execed", ent.Name)
    end
  end
  print("SCRIPT: Done with everyone")
  return false
end
