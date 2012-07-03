function mantra(buff, condition)
	conditions = getConditions (Me)
	if not conditions [condition] then 
		doBasicAttack (buff, Me)
	end
end
				
function retaliate(melee)
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = Me.Info().LastEntityThatAttackedMe
		if exists (target) then
			if rangedDistBetweenEntities (Me, target) == 1 then
				return doBasicAttack (melee, target)
			else
				ps = AllPathablePoints (Me.Pos, pos (target), 1, 1)
				if table.getn (ps) > 0 then
					return doMove (ps, 1000)
				end
			end
		end
	end
end

function pursue(melee)
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = Me.Info().LastEntityThatIAttacked
		if exists (target) then
			if rangedDistBetweenEntities (Me, target) == 1 then
				return doBasicAttack (melee, target)
			else
				ps = AllPathablePoints (Me.Pos, pos (target), 1, 1)
				if table.getn (ps) > 0 then
				  return doMove (ps, 1000)
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

				

				

	