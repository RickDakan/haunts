function mantra(buff, condition)
	conditions = getConditions (Me())
	if not conditions [condition] then 
		DoBasicAttack (buff, Me())
	end
end
				
function retaliate(melee)
	intruders = NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = Me().Info().LastEntityThatAttackedMe
		if Exists (target) then
			if RangedDistBetweenEntities (Me(), target) == 1 then
				return DoBasicAttack (melee, target)
			else
				ps = AllPathablePoints (Me().Pos, pos (target), 1, 1)
				if table.getn (ps) > 0 then
					return DoMove (ps, 1000)
				end
			end
		end
	end
end

function pursue(melee)
	intruders = NearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = Me().Info().LastEntityThatIAttacked
		if Exists (target) then
			if RangedDistBetweenEntities (Me(), target) == 1 then
				return DoBasicAttack (melee, target)
			else
				ps = AllPathablePoints (Me().Pos, pos (target), 1, 1)
				if table.getn (ps) > 0 then
				  return DoMove (ps, 1000)
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

				

				

	