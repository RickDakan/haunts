function mantra(buff, condition)
	conditions = getConditions (me())
	if not conditions [condition] then 
		doBasicAttack (buff, me())
	end
end
				
function retaliate(melee)
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(me()).lastEntityThatAttackedMe
		if exists (target) then
			if rangedDistBetweenEntities (me(), target) == 1 then
				return doBasicAttack (melee, target)
			else
				ps = allPathablePoints (pos(me()), pos (target), 1, 1)
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
		target = entityInfo(me()).lastEntityIAttacked
		if exists (target) then
			if rangedDistBetweenEntities (me(), target) == 1 then
				return doBasicAttack (melee, target)
			else
				ps = allPathablePoints (pos(me()), pos (target), 1, 1)
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

				

				

	