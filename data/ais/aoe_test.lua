-- aoe_test
function aoePlaceAndAttack(attack, spec)
	gz = BestAoeAttackPos (attack, me().apCur - me().actions[attack].ap, spec)
	dsts = AllPathablePoints(pos(me()), gz, 1, me().actions[attack].range)
	DoMove(dsts, 1000)
	if RangedDistBetweenPositions (me().pos, gz) > me().actions[attack].range then
		return
	else
		DoAoeAttack(attack, gz)
	end
end

function Think()
	aoePlaceAndAttack("Abjuration", "enemies only")
	think()
end
