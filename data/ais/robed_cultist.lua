function mantra(buff, condition)
	if not Me.Conditions[condition] then 
		Do.BasicAttack (buff, Me)
	end
end
				
function retaliate(melee)
	intruders = Utils.NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = Me.Info.LastEntityThatAttackedMe
		if Utils.Exists (target) then
			if Utils.RangedDistBetweenEntities (Me, target) == 1 then
				return Do.BasicAttack (melee, target)
			else
				ps = Utils.AllPathablePoints (Me.Pos, pos (target), 1, 1)
				if table.getn (ps) > 0 then
					return Do.Move (ps, 1000)
				end
			end
		end
	end
end

function pursue(melee)
	intruders = Utils.NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = Me.Info.LastEntityThatIAttacked
		if Utils.Exists (target) then
			if Utils.RangedDistBetweenEntities (Me, target) == 1 then
				return Do.BasicAttack (melee, target)
			else
				ps = Utils.AllPathablePoints (Me.Pos, pos (target), 1, 1)
				if table.getn (ps) > 0 then
				  return Do.Move (ps, 1000)
				end
			end
		end
	end
end

function attack()
	melee = "Sacrificial Blade"
	p = pursue(melee)
	r = retaliate(melee)
	return p or r
end

function think ()
	mantra("Cultic Mantra", "Focused")
	while attack() do
	end
end

think ()

				

				

	