Utils Functions
---------------

This table has functions that allows an entity ai to query the game for information.  Generally, you will not be able to get information that a human player wouldn't be able to get in the same situation.


###_dsts_ = Utils.__AllPathablePoints__(_src_, _dst_, _min_, _max_)
_src_: Where the path starts.  
_dst_: A point near where the path ends.  
_min_: Minimum distance from _dst_ that the path should end.  
_max_: Maximum distance from _dst_ that the path should end.  

_dsts_: An array of all points that can be reached by walking from _src_ to within _min_ and _max_ *ranged* distance of _dst_.  Assumes that a 1x1 unit is doing the walking.

------

###_center_, _hits_ = Utils.__BestAoeAttackPos__(_attack_, _extra_dist_, _spec_)
_attack_: Name of the aoe attack to use.  
_extra_dist_: Maximum extra distance to move before using the attack.  
_spec_: A value indicating if it is ok to hit allied units.  "allies ok", "minions ok", and "enemies only" are the acceptable values.

_center_: Where to place the aoe for maximum effect (i.e. maximum number of enemy entities hit), might need to move first to get within range.  
_hits_: An array containing all of the entities that would be hit by the aoe if it is centered on _center_.

------

###_exists_ = Utils.__Exists__(_ent_)
_ent_: The entity to query.

_exists_: True iff _ent_ is an Entity object, still exists, and has positive health.  Note that this means that you can pass nil to this function and it will return false (rather than throwing an error).

------

###_ents_ = Utils.__NearestNEntities__(_max_, _kind_)
_max_: Maximum number of entities to return.  
_kind_: What entities to look for.  The following values are accetpable: "intruder" "denizen" "minion" "servitor" "master" "non-minion" "non-servitor" "non-master" "all".  

_ents_: An array containing the nearest _max_ entities in LoS that match _kind_.  If there are fewer than _max_ entities then as many as possible will be returned.

------

###_dist_ = Utils.__RangedDistBetweenPositions__(_p1_, _p2_)
_p1_: A point.  
_p2_: Another point.  

_dist_: The ranged distance between _p1_ and _p2_.

------

###_dist_ = Utils.__RangedDistBetweenEntities__(_e1_, _e2_)
_e1_: An entity.  
_e2_: Another entity.  

_dist_: The ranged distance between _e1_ and _e2_.  Note that if either entity is larger than 1x1 this might not return the same value as Utils.__RangedDistBetweenPositions__(_e1_.Pos, _e2_.Pos)

