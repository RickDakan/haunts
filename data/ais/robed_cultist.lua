function mantra(buff, condition)
	conditions = getConditions (me())
	if not conditions [condition] then 
		print ("buffing")
		doBasicAttack (buff, me())
	end
end
				
function retaliate(melee)
	intruders = nearestNEntity (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(me()).lastEntityThatAttackedMe
		if exists (target) then
			if rangedDistBetween (me(), target) == 1 then
				doBasicAttack (melee, target)
			end
		else
			ps = allPathablePoints (pos(me()), pos (target), 1, 10)
			if table.getn (ps) > 0 then
				doMove (ps, 1000)
			end
		end
	end
end

function pursue(melee)
	intruders = nearestNEntity (10, "intruder")
	for _, intruder in pairs (intruders) do
		target = entityInfo(me()).lastEntityIAttacked
		if exists (target) then
			if rangedDistBetween (me(), target) == 1 then
				doBasicAttack (melee, target)
			end
		else
			ps = allPathablePoints (pos(me()), pos (target), 1, 10)
			if table.getn (ps) > 0 then
				doMove (ps, 1000)
			end
		end
	end
end



function think ()
	melee = "Sacrificial Blade"
	buff = "Cultic Mantra"
	
	mantra("Cultic Mantra", "Focused")
--	pursue("Sacrificial Blade")
--	retaliate(melee)
end
think ()

				

				

	