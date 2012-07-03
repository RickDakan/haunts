ps = allPathablePoints(pos(me()), pos(me()), 1, 5)
r = randN(table.getn(ps))
target = ps[randN(table.getn(ps))]
a = {}
a[1] = target
res = doMove(a, 1000)
