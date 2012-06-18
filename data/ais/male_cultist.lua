

-- if there's someone next to him without Agony, inject him
-- move away from someone
-- shoot them!


--check to see if adjacent people have Agony - if they do, he wants to move away


function moveAndAttack (attack, target)
	ps = allPathablePoints (pos(me()), pos(target), 1, getBasicAttackStats(me(), attack).range)
	res = doMove (ps, 1000)
	if exists(target) then
		doBasicAttack(attack, target)
	end
end
	
--keep distance

function keepDistanceAndAttack (distance, attack, target)
	ps = allPathablePoints (pos(me()), pos(target), distance, getBasicAttackStats(me(), attack).range)
	doMove (ps, 1000)
	if exists(target) then
		doBasickAttack(attack, target)
	end
end



function think()
	intruders = nearestNEntities (10, "intruder")
	for _, intruder in pairs (intruders) do
		if rangedDistBetweenEntities (me(), intruder) >1 then
			break
		end
		if nil == getConditions (intruder) ["Agony"] then
			moveAndAttack("Inject", intruder)
		end
	end
end
think()


	