package main

func Routes() {

	// Check Service
	SERVER.GET("/", HealthCheck)
	
	// List session play by user session
	SERVER.GET("/list",	SessionPlayList)

	// Create session play
	SERVER.POST("/init", SessionPlayInit)

	// Active Session play
	SERVER.PUT("/active/:sessionPlayId", SessionPlayActived)

	// Updated playing
	SERVER.PUT("/playing", SessionPlayWatching)

	// Kick session
	SERVER.DELETE("/kick", SessionPlayKickout)

	// Get Playlist
	SERVER.GET("/play/:movieId/:playId", GetPlayList)
	
}