Utils Functions
---------------

This table has functions that allows an entity ai to query the game for information.  Generally, you will not be able to get information that a human player wouldn't be able to get in the same situation.

###__dsts__ = Utils.AllPathablePoints(__src__, __dst__, __min__, __max__)
__src__: Where the path starts.  
__dst__: A point near where the path ends.  
__min__: Minimum distance from __dst__ that the path should end.  
__max__: Maximum distance from __dst__ that the path should end.  

__dsts__: An array of all points that can be reached by walking from __src__ to within __min__ and __max__ *ranged* distance of __dst__.  Assumes that a 1x1 unit is doing the walking.


###__center__, __hits__ = Utils.BestAoeAttackPos(__attack__, __extra_dist__, __spec__)
__attack__: Name of the aoe attack to use.  
__extra_dist__: Maximum extra distance to move before using the attack.  
__spec__: A value indicating if it is ok to hit allied units.  "allies ok", "minions ok", and "enemies only" are the acceptable values.

__center__: Where to place the aoe for maximum effect (i.e. maximum number of enemy entities hit), might need to move first to get within range.  
__hits__: An array containing all of the entities that would be hit by the aoe if it is centered on __center__.


###__exists__ = Utils.Exists(__ent__)
__ent__: The entity to query.

__exists__: True iff __ent__ is an Entity object, still exists, and has positive health.  Note that this means that you can pass nil to this function and it will return false (rather than throwing an error).


###__ents__ = Utils.NearestNEntities(__max__, __kind__)
__max__: Maximum number of entities to return.  
__kind__: What entities to look for.  The following values are accetpable: "intruder" "denizen" "minion" "servitor" "master" "non-minion" "non-servitor" "non-master" "all".  

__ents__: An array containing the nearest __max__ entities in LoS that match __kind__.  If there are fewer than __max__ entities then as many as possible will be returned.


###__dist__ = Utils.RangedDistBetweenPositions(__p1__, __p2__)
__p1__: A point.  
__p2__: Another point.  

__dist__: The ranged distance between __p1__ and __p2__.


###__dist__ = Utils.RangedDistBetweenEntities(__e1__, __e2__)
__e1__: An entity.  
__e2__: Another entity.  

__dist__: The ranged distance between __e1__ and __e2__.  Note that if either entity is larger than 1x1 this might not return the same value as Utils.RangedDistBetweenPositions(__e1__.Pos, __e2__.Pos)


NearbyUnexploredRoom
RoomPath
RoomContaining
AllDoorsBetween
AllDoorsOn
DoorPositions
DoorIsOpen
RoomPositions
