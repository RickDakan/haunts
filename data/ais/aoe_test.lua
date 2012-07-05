-- aoe_test
function aoePlaceAndAttack(attack, spec)
	gz = Utils.BestAoeAttackPos (attack, me().apCur - me().actions[attack].ap, spec)
	dsts = Utils.AllPathablePoints(pos(me()), gz, 1, me().actions[attack].range)
	Actions.Move(dsts, 1000)
	if Utils.RangedDistBetweenPositions (me().pos, gz) > me().actions[attack].range then
		return
	else
		Actions.AoeAttack(attack, gz)
	end
end

function Think()
	aoePlaceAndAttack("Abjuration", "enemies only")
	think()
end
