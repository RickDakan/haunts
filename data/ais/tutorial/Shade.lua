function Think()
  while CrushIntruder(nil, nil, "Chill Touch", nil, nil) do
  end
  target = GetTarget()
  if target and Me.ApCur < Me.ApMax then
    HeadTowards(target.Pos)
  end
  return false
end
