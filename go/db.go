package main

import (
    "database/sql"
    "log"
)

type databaseHandler struct {
    db *sql.DB
}

func (dbHandler *databaseHandler) executeSQLNoReturn(sqlStmt string) {
    _, err := dbHandler.db.Exec(sqlStmt)
    if err != nil {
        log.Fatal(err)
    }
}

func (dbHandler *databaseHandler) createUsersTable() {
    dbHandler.executeSQLNoReturn(`
	CREATE TABLE IF NOT EXISTS users (
		playerId TEXT PRIMARY KEY,
		name TEXT NOT NULL
	);
	`)
    log.Printf("Created users table")
}

func (dbHandler *databaseHandler) createGamesTable() {
    dbHandler.executeSQLNoReturn(`
	CREATE TABLE IF NOT EXISTS games (
        gameId TEXT PRIMARY KEY,
        mapType TEXT NOT NULL,
        mapSize TEXT NOT NULL,
        easyBots INT NOT NULL,
        normalBots INT NOT NULL,
        hardBots INT NOT NULL,
        crazyBots INT NOT NULL
    )
	`)
    log.Printf("Created games table")
}

func (dbHandler *databaseHandler) createUsersToGamesTable() {
    dbHandler.executeSQLNoReturn(`
    CREATE TABLE IF NOT EXISTS usersToGames (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        gameId TEXT NOT NULL,
        playerId TEXT NOT NULL,
        tribe TEXT NOT NULL,
        FOREIGN KEY (gameId) REFERENCES games(gameId),
        FOREIGN KEY (playerId) REFERENCES users(playerId),
        UNIQUE(gameId, playerId)
    );
    `)
    log.Printf("Created users to games table")
}

func (dbHandler *databaseHandler) createTables() {
    dbHandler.createUsersTable()
    dbHandler.createGamesTable()
    dbHandler.createUsersToGamesTable()
}

func (dbHandler *databaseHandler) gameExists(gameId string) bool {
    var exists bool
    err := dbHandler.db.
    QueryRow("SELECT EXISTS(SELECT 1 FROM games WHERE gameId = ?)", gameId).
    Scan(&exists)
    if err != nil {
        log.Fatal(err)
    }
    return exists
}

func (dbHandler *databaseHandler) getGameData(gameId string) *gameData {
    if !dbHandler.gameExists(gameId) {
        log.Fatal("Game does not exist: " + gameId)
    }
    sqlStmt := ` SELECT mapType, mapSize,
    easyBots, normalBots, hardBots, crazyBots
    FROM games WHERE gameId = ?
    `
    row := dbHandler.db.QueryRow(sqlStmt, gameId)
    var game gameData
    err := row.Scan(
        &game.MapType, &game.MapSize,
        &game.Bots.Easy, &game.Bots.Normal, &game.Bots.Hard, &game.Bots.Crazy,
    )
    if err != nil {
        log.Fatal(err)
    }
    return &game
}

func (dbHandler *databaseHandler) createGame(gameId string, game *gameData) {
    if dbHandler.gameExists(gameId) {
        log.Fatal("Game already exists: " + gameId)
    }
    sqlStmt := `
    INSERT INTO games (
        gameId, mapType, mapSize,
        easyBots, normalBots, hardBots, crazyBots
    )
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
    _, err := dbHandler.db.Exec(
        sqlStmt,
        gameId, game.MapType, game.MapSize,
        game.Bots.Easy, game.Bots.Normal, game.Bots.Hard, game.Bots.Crazy,
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created new game %s", gameId)
}

func (dbHandler *databaseHandler) updateGame(gameId string, game *gameData) {
    if !dbHandler.gameExists(gameId) {
        log.Fatal("Game does not exist: " + gameId)
    }
    sqlStmt := `
    UPDATE games SET
    mapType = ?, mapSize = ?,
    easyBots = ?, normalBots = ?,
    hardBots = ?, crazyBots = ?
    WHERE gameId = ?
    `
    _, err := dbHandler.db.Exec(
        sqlStmt,
        game.MapType, game.MapSize,
        game.Bots.Easy, game.Bots.Normal, game.Bots.Hard, game.Bots.Crazy,
        gameId,
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Updated game %s", gameId)
}

func (dbHandler *databaseHandler) getGamePlayers(gameID string) []string {
    sqlStmt := "SELECT playerId FROM usersToGames WHERE gameId = ?"
    rows, err := dbHandler.db.Query(sqlStmt, gameID)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    var players []string
    for rows.Next() {
        var playerID string
        err = rows.Scan(&playerID)
        if err != nil {
            log.Fatal(err)
        }
        players = append(players, playerID)
    }
    return players
}

func (dbHandler *databaseHandler) getPlayerGames(playerID string) []string {
    sqlStmt := "SELECT gameId FROM usersToGames WHERE playerId = ?"
    rows, err := dbHandler.db.Query(sqlStmt, playerID)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    var games []string
    for rows.Next() {
        var gameID string
        err = rows.Scan(&gameID)
        if err != nil {
            log.Fatal(err)
        }
        games = append(games, gameID)
    }
    return games
}

func (dbHandler *databaseHandler) updateTribe(playerId string, gameId string, tribe string) {
    sqlStmt := "UPDATE usersToGames SET tribe = ? WHERE playerId = ? AND gameId = ?"
    _, err := dbHandler.db.Exec(sqlStmt, tribe, playerId, gameId)
    if err != nil {
        log.Fatal(err)
    }
    name := dbHandler.getName(playerId)
    log.Printf("Updating tribe for %s (%s) in game %s to %s", name, playerId, gameId, tribe)
}

func (dbHandler *databaseHandler) addPlayerToGame(playerId string, gameId string, tribe string) {
    sqlStmt := "INSERT INTO usersToGames (gameId, playerId, tribe) VALUES (?, ?, ?)"
    _, err := dbHandler.db.Exec(sqlStmt, gameId, playerId, tribe)
    if err != nil {
        log.Fatal(err)
    }
    name := dbHandler.getName(playerId)
    log.Printf("Adding %s (%s) to game %s with tribe %s", name, playerId, gameId, tribe)
}

func (dbHandler *databaseHandler) getTribe(playerId string, gameId string) string {
    sqlStmt := "SELECT tribe FROM usersToGames WHERE gameId = ? AND playerId = ?"
    var tribe string
    err := dbHandler.db.QueryRow(sqlStmt, gameId, playerId).Scan(&tribe)
    if err != nil {
        log.Fatal(err)
    }
    return tribe

}

func (dbHandler *databaseHandler) playerExists(playerID string) bool {
    var exists bool
    err := dbHandler.db.
    QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE playerId = ?)", playerID).
    Scan(&exists)
    if err != nil {
        log.Fatal(err)
    }
    return exists
}

func (dbHandler *databaseHandler) createPlayer(playerID string, name string) {
    if dbHandler.playerExists(playerID) {
        log.Fatal("Player already exists: " + playerID)
    }
    sqlStmt := `
    INSERT INTO users (playerID, name) VALUES (?, ?);
    `
    _, err := dbHandler.db.Exec(sqlStmt, playerID, name, name)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created player %s with playerID %s", name, playerID)
}

func (dbHandler *databaseHandler) updateName(playerID string, name string) {
    if !dbHandler.playerExists(playerID) {
        log.Fatal("Player does not exist: " + playerID)
    }
    previousName := dbHandler.getName(playerID)
    sqlStmt := `
    UPDATE users SET name = ? WHERE playerId = ?
    `
    _, err := dbHandler.db.Exec(sqlStmt, name, playerID)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Changing name of %s from %s to %s", playerID, previousName, name)
}

func (dbHandler *databaseHandler) getName(playerId string) string {
    if !dbHandler.playerExists(playerId) {
        log.Fatal("Player does not exist: " + playerId)
    }
    var name string
    err := dbHandler.db.
    QueryRow("SELECT name FROM users WHERE playerId = ?", playerId).
    Scan(&name)
    if err != nil {
        log.Fatal(err)
    }
    return name
}
