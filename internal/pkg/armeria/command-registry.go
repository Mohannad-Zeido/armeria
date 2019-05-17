package armeria

func RegisterGameCommands(state *GameState) {
	commands := []*Command{
		{
			Name: "login",
			Help: "Logs your character into the game world.",
			Permissions: &CommandPermissions{
				RequireNoCharacter: true,
			},
			Arguments: []*CommandArgument{
				{
					Name: "character",
				},
				{
					Name: "password",
				},
			},
			Handler: handleLoginCommand,
		},
		{
			Name: "look",
			Help: "Looks at something.",
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Handler: handleLookCommand,
		},
		{
			Name: "say",
			Help: "Says something to everyone in the same room as you.",
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Arguments: []*CommandArgument{
				{
					Name:             "text",
					IncludeRemaining: true,
				},
			},
			Handler: handleSayCommand,
		},
		{
			Name: "move",
			Help: "Moves your character into a connecting room.",
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Arguments: []*CommandArgument{
				{
					Name: "direction",
				},
			},
			Handler: handleMoveCommand,
		},
		{Name: "north", Alias: "move north"},
		{Name: "south", Alias: "move south"},
		{Name: "east", Alias: "move east"},
		{Name: "west", Alias: "move west"},
		{Name: "up", Alias: "move up"},
		{Name: "down", Alias: "move down"},
		{
			Name: "room",
			Help: "Allows you to manage rooms.",
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Subcommands: []*Command{
				{
					Name:    "edit",
					Help:    "Opens the editor panel for the current room.",
					Handler: handleRoomEditCommand,
				},
				{
					Name: "set",
					Help: "Allows you to set a room attribute. Leave value empty to revert to default.",
					Arguments: []*CommandArgument{
						{
							Name: "property",
						},
						{
							Name:             "value",
							IncludeRemaining: true,
							Optional:         true,
						},
					},
					Handler: handleRoomSetCommand,
				},
				{
					Name: "create",
					Help: "Creates a new room in a particular direction.",
					Arguments: []*CommandArgument{
						{
							Name: "direction",
						},
					},
					Handler: handleRoomCreateCommand,
				},
			},
		},
		{
			Name: "save",
			Help: "Writes the in-memory game data to disk.",
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Handler: handleSaveCommand,
		},
		{
			Name: "reload",
			Help: "Updates, builds and reloads the server, client, or both.",
			Arguments: []*CommandArgument{
				{
					Name: "component",
				},
			},
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Handler: handleReloadCommand,
		},
		{
			Name: "map",
			Help: "Asks the server to re-render the minimap for the area.",
			Permissions: &CommandPermissions{
				RequireCharacter: true,
			},
			Handler: handleMapCommand,
		},
	}

	for _, cmd := range commands {
		state.commandManager.RegisterCommand(cmd)
	}
}
