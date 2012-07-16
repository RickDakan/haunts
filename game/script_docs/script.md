Script Functions
----------------

This table has script functions.
------

###_housename_ = Script.__StartScript__(_path_)
Terminates this script and starts a new one, specified by _path_.  
_path_: Path of the script to run, relative to the scripts directory.  

------

###_housename_ = Script.__SelectHouse__()
Pops up a ui allowing the user to select a house from all available houses.  
_housename_: Name of the house that the user selected.  

------

###Script.__LoadHouse__(_housename_)
Loads a fresh house.  
_housename_: Name of the house to load.  This will get rid of everything in the current house, if there is one.

------

###_dsts_ = Utils.AllPathablePoints(_src_, _dst_, _min_, _max_)
Finds all positions that a 1x1 unit could walk to from one position to nearby another.  
_src_: Where the path starts.  
_dst_: A point near where the path ends.  
_min_: Minimum distance from _dst_ that the path should end.  
_max_: Maximum distance from _dst_ that the path should end.  

_dsts_: An array of all points that can be reached by walking from _src_ to within _min_ and _max_ *ranged* distance of _dst_.  Assumes that a 1x1 unit is doing the walking.

------

###Script.__ShowMainBar__(_show_)
Shows/hides the main ui bar.  
_show_: Boolean indicating whether the main bar should be showing or not.  

------

###_ent_ = Script.__SpawnEntityAtPosition__(_name_, _pos_)
No description necessary.  
_name_: Name of the entity to spawn.  
_pos_: Position to spawn the entity at, this position must be empty or the entity will not spawn.  

_ent_: The entity that was just spawned, or nil if it could not be spawned.  

------

###_spawn_points_ = Script.__GetSpawnPointsMatching__(_regexp_)
Finds all spawn points that have a name matching a regexp.  
_regexp_: A string describing a regular expression.  Regular expressions are very powerful but can also get quite complicated.  For most purposes it is probably enough to know that <pre>".*"</pre> matches anything, so if your regexp is <pre>"Foo-.*"</pre> then you will match all strings that begin with "Foo-".  

_spawn_points_: An array of all spawn points whose names match _regexp_.  

------

###_ent_ = Script.__SpawnEntitySomewhereInSpawnPoints__(_name_, _spawn_point_)
Spawns an entity randomly in a set of spawn points.  
_name_: Name of the entity to spawn.  
_spawn_points_: An array of spawn points to spawn the entity in.  

_ent_: The entity that was spawned, or nil if it could not be spawned.

------

###_placed_ = Script.__PlaceEntities__(_regexp_, _points_, _ents_)
Provides an ui to the user to place entities in the house.  
_regexp_: A string describing a regular expression.  The spawn points whose names match _regexp_ will be available to the user to place the entities.  
_points_: The number of points available to the user to spend when placing these entities.  
_ents_: A table mapping entity name to point cost of that entity.  

------

###_room_ = Script.__RoomAtPos__(_pos_)
Finds the room that contains a position.  
_pos_: A point.  

_room_: The room containing the point.  

------

###_ents_ = Script.__GetAllEnts__()
Returns an array of all entities.  
_ents_: An array of entities.

------

###_choices_ = Script.__DialogBox__(_filename_)
Pops up a series of dialog boxes, specified by the file at _filename_.  
_filename_: Path to a file describing the series of dialog boxes to show to the user.  

_choices_: Array of choices that the user made.  The values in the array will correspond to the 'Id' value of the dialog boxes specify choices, in the order that they are shown to the user.  

------

###_choices_ = Script.__PickFromN__(_min_, _max_, _options_)
Pops up windows that allows the user to select one or more things from a list.  
_min_: Minimum options that the user must select.  
_max_: Maximum options that the user must select.  
_options_: Array of paths to icons to show the user.  

_choices_: Array of indices of options that the user chose.  

------

###_successful_ = Script.__SetGear__(_ent_, _gear_)
Sets the gear that _ent_ is using to _gear_.  
_ent_: An intruder entity, if _ent_ is not an intruder this function will do nothing.  
_gear_: The name of the gear for _ent_ to use.  
_successful_: True iff _ent_'s gear was set to _gear_.  

------

###Script.__SetCondition__(_ent_, _name_, _set_)
Sets whether or not _ent_ has the condition named _name_.  
_ent_: The entity to apply/remote this condition from.  
_name_: The name of the condition.  
_set_: A boolean, true if this condition should be applied, false if it should be removed.  If you re-apply a condition it will reset its duration, if you remove a condition that _ent_ doesn't have nothing will happen.  

------

###Script.__SetPosition__(_ent_, _pos_)
Moves _ent_ to _pos_.  
_ent_: The entity to move.  
_pos_: The position to move _ent_ to.  

------

###Script.__BindAi__(_target_, _source_)
Binds an ai to something.  
_target_: The thing to bind the ai to.  This can be an entity, or it can be one of the following strings: "denizen" "intruder" "minions".  
_source_: The path to the ai file to bind, or one of the following strings: "human" "net".  

------

###Script.__SetVisibility__(_side_)
Indicates what side's Pov the user views the game.
_side_: Either "denizens" or "intruders".  

------

###Script.__SetLosMode__(_side_, _mode_)
Sets what is visible to a given side.  
_side_: One of "denizens" or "intruders".  
_mode_: One of "none", "all", or "entities", or an array of rooms.  "none" will make everything dark, "all" makes everything visible, "entities" indicates that visibility will be determined by whatever entities are on that side (standard for gameplay).  If _mode_ is an array of rooms then visibility will be exactly those rooms.

------

###Script.__EndPlayerInteraction__()
Indicates that the current human player is done.  Ais will continue to perform their actions and then the turn will end.  

------

###Script.__SaveStore__()
Saves any values written to the store table.  This should be called any time data is written to those files to ensure that it is available later.  


