Script Functions
----------------

This table has script functions.

###__housename__ = Script.SelectHouse()
Pops up a ui allowing the user to select a house from all available houses.  
__housename__: Name of the house that the user selected.  


###Script.LoadHouse(__housename__)
Loads a fresh house.  
__housename__: Name of the house to load.  This will get rid of everything in the current house, if there is one.


###__dsts__ = Utils.AllPathablePoints(__src__, __dst__, __min__, __max__)
Finds all positions that a 1x1 unit could walk to from one position to nearby another.  
__src__: Where the path starts.  
__dst__: A point near where the path ends.  
__min__: Minimum distance from __dst__ that the path should end.  
__max__: Maximum distance from __dst__ that the path should end.  

__dsts__: An array of all points that can be reached by walking from __src__ to within __min__ and __max__ *ranged* distance of __dst__.  Assumes that a 1x1 unit is doing the walking.


###Script.ShowMainBar(__show__)
Shows/hides the main ui bar.  
__show__: Boolean indicating whether the main bar should be showing or not.  


###__ent__ = Script.SpawnEntityAtPosition(__name__, __pos__)
No description necessary.  
__name__: Name of the entity to spawn.  
__pos__: Position to spawn the entity at, this position must be empty or the entity will not spawn.  

__ent__: The entity that was just spawned, or nil if it could not be spawned.  


###__spawn_points__ = Script.GetSpawnPointsMatching(__regexp__)
Finds all spawn points that have a name matching a regexp.  
__regexp__: A string describing a regular expression.  Regular expressions are very powerful but can also get quite complicated.  For most purposes it is probably enough to know that <pre>".*"</pre> matches anything, so if your regexp is <pre>"Foo-.*"</pre> then you will match all strings that begin with "Foo-".  

__spawn_points__: An array of all spawn points whose names match __regexp__.  


###__ent__ = Script.SpawnEntitySomewhereInSpawnPoints(__name__, __spawn_point__)
Spawns an entity randomly in a set of spawn points.  
__name__: Name of the entity to spawn.  
__spawn_points__: An array of spawn points to spawn the entity in.  

__ent__: The entity that was spawned, or nil if it could not be spawned.


###__placed__ = Script.PlaceEntities(__regexp__, __points__, __ents__)
Provides an ui to the user to place entities in the house.  
__regexp__: A string describing a regular expression.  The spawn points whose names match __regexp__ will be available to the user to place the entities.  
__points__: The number of points available to the user to spend when placing these entities.  
__ents__: A table mapping entity name to point cost of that entity.  


###__room__ = Script.RoomAtPos(__pos__)
Finds the room that contains a position.  
__pos__: A point.  

__room__: The room containing the point.  


###__ents__ = Script.GetAllEnts()
Returns an array of all entities.  
__ents__: An array of entities.


###__choices__ = Script.DialogBox(__filename__)
Pops up a series of dialog boxes, specified by the file at __filename__.  
__filename__: Path to a file describing the series of dialog boxes to show to the user.  

__choices__: Array of choices that the user made.  The values in the array will correspond to the 'Id' value of the dialog boxes specify choices, in the order that they are shown to the user.  


###__choices__ = Script.PickFromN(__min__, __max__, __options__)
Pops up windows that allows the user to select one or more things from a list.  
__min__: Minimum options that the user must select.  
__max__: Maximum options that the user must select.  
__options__: Array of paths to icons to show the user.  

__choices__: Array of indices of options that the user chose.  


###__successful__ = Script.SetGear(__ent__, __gear__)
Sets the gear that __ent__ is using to __gear__.  
__ent__: An intruder entity, if __ent__ is not an intruder this function will do nothing.  
__gear__: The name of the gear for __ent__ to use.  

__successful__: True iff __ent__'s gear was set to __gear__.

###Script.BindAi(__target__, __source__)
Binds an ai to something.  
__target__: The thing to bind the ai to.  This can be an entity, or it can be one of the following strings: "denizen" "intruder" "minions".  
__source__: The path to the ai file to bind, or one of the following strings: "human" "net".  


###Script.SetVisibility(__side__)
Indicates what side's Pov the user views the game.
__side__: Either "denizens" or "intruders".  


###Script.SetLosMode(__side__, __mode__)
Sets what is visible to a given side.  
__side__: One of "denizens" or "intruders".  
__mode__: One of "none", "all", or "entities", or an array of rooms.  "none" will make everything dark, "all" makes everything visible, "entities" indicates that visibility will be determined by whatever entities are on that side (standard for gameplay).  If __mode__ is an array of rooms then visibility will be exactly those rooms.


###Script.EndPlayerInteraction()
Indicates that the current human player is done.  Ais will continue to perform their actions and then the turn will end.  


###Script.SaveStore()
Saves any values written to the store table.  This should be called any time data is written to those files to ensure that it is available later.  


