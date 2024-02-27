package main

import (
    "encoding/hex"
    "math/rand"
    crand "crypto/rand"
    "time"
    "log"
    "net/http"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "html/template"
)

const COOKIE_NAME = "playerId"
const DEFAULT_USERNAME = "New Player"

type application struct {
    db databaseHandler
    totalPlayers int
}

type botDifficulties struct {
    Easy int
    Normal int
    Hard int
    Crazy int
}

type gameData struct {
    MapType string
    MapSize string
    Bots botDifficulties
}

func main() {
    log.Println("Opening database ...")
    db, err := sql.Open("sqlite3", "database/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

    app := &application{
        db: databaseHandler{
            db: db,
        },
        totalPlayers: 4,
    }
    app.db.createTables()

    mux := http.NewServeMux()

    mux.HandleFunc("/", app.polytopiaHome)
    mux.HandleFunc("/g/", app.polytopiaGame)
    mux.HandleFunc("/generate-game", app.generateGame)
    mux.HandleFunc("/get-game", app.getGame)
    mux.HandleFunc("/profile", app.profile)
    mux.HandleFunc("/change-name", app.changeName)

    fileServer := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fileServer))

    wrappedMux := app.ensureUserExists(mux)

    certFile := "certs/polytopia_lynas_dev.crt"
    keyFile := "certs/polytopia_lynas_dev.key"

    log.Println("Listening ...")
    err = http.ListenAndServeTLS(":8443", certFile, keyFile, wrappedMux)
    if err != nil {
        log.Fatal(err)
    }
}

func (app *application) generateCookie() *http.Cookie {
    cookieValue := app.generateRandomHash(16)
    expiration := time.Now().AddDate(100, 0, 0) // 100 years
    cookie := &http.Cookie{
        Name: COOKIE_NAME,
        Value: cookieValue,
        Path: "/",
        Expires: expiration,
    }
    return cookie
}

func (app *application) ensureUserExists(next http.Handler) http.Handler {
    // Ensure the user cookie is set and that it exists in the database
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set cookie if it is not already set
        cookie, err := r.Cookie(COOKIE_NAME);
		if err != nil {
			cookie = app.generateCookie()
			http.SetCookie(w, cookie)
		}
        // Add to the database if it doesnt exist
        if !app.db.playerExists(cookie.Value) {
            name := DEFAULT_USERNAME
            app.db.createPlayer(cookie.Value, name)
        }
		next.ServeHTTP(w, r)
	})
}

func (app *application) getUserCookie(r *http.Request) *http.Cookie {
    cookie, err := r.Cookie(COOKIE_NAME)
    if err != nil {
        log.Fatal(err)
    }
    return cookie
}

func (app *application) profile(w http.ResponseWriter, r *http.Request) {
    type profileData struct {
        Name string
    }

    files := []string {
        "html/base.tmpl",
        "html/profile.tmpl",
    }

    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        log.Fatal(err)
    }

    cookie := app.getUserCookie(r)
    name := app.db.getName(cookie.Value)
    data := profileData{
        Name: name,
    }
    err = tmpl.Execute(w, data)
    if err != nil {
        log.Fatal(err)
    }
}

func (app *application) changeName(w http.ResponseWriter, r *http.Request) {
    // Change name and return new name
    if r.Method != http.MethodPost {
        http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
        return
    }
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusInternalServerError)
        return
    }
    name := r.FormValue("name")
    if name == "" {
        http.Error(w, "Name cannot be empty", http.StatusBadRequest)
        return
    }
    cookie := app.getUserCookie(r)
    app.db.updateName(cookie.Value, name)
    w.Write([]byte(name))
}

func (app *application) polytopiaHome(w http.ResponseWriter, r *http.Request) {
    files := []string {
        "html/base.tmpl",
    }

    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        log.Fatal(err)
    }

    err = tmpl.Execute(w, nil)
    if err != nil {
        log.Fatal(err)
    }
}

func (app *application) randomGame(players int) gameData {
    bots := app.getBotDifficulties(app.totalPlayers - players)
    return gameData{
        MapType: app.randomMapType(),
        MapSize: app.randomMapSize(),
        Bots: bots,
    }
}

func (app *application) makeNewGame(gameId string) {
    game := app.randomGame(0)
    app.db.createGame(gameId, &game)
}

func (app *application) regenerateGameBots(gameId string) {
    game := app.db.getGameData(gameId)
    numberOfPlayers := len(app.db.getGamePlayers(gameId))
    bots := app.getBotDifficulties(app.totalPlayers - numberOfPlayers)
    game.Bots = bots
    app.db.updateGame(gameId, game)
}

func (app *application) addPlayerToGame(gameId string, playerId string) {
    tribe := app.getRandomTribe()
    log.Printf("Adding player %s to game %s with tribe %s and regenerating bots", playerId, gameId, tribe)
    app.db.addPlayerToGame(playerId, gameId, tribe)
    app.regenerateGameBots(gameId)
}

func (app *application) playerInGame(playerId string, gameId string) bool {
    players := app.db.getGamePlayers(gameId)
    for _, player := range players {
        if player == playerId {
            return true
        }
    }
    return false
}

func (app *application) polytopiaGame(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        id = app.generateRandomHash(8)
        http.Redirect(w, r, "/g?id=" + id, http.StatusFound)
        return
    }

    cookie := app.getUserCookie(r)

    if !app.db.gameExists(id) {
        app.makeNewGame(id)
    }
    if !app.playerInGame(cookie.Value, id) {
        app.addPlayerToGame(id, cookie.Value)
    }

    files := []string {
        "html/base.tmpl",
        "html/polytopia.tmpl",
    }

    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        log.Fatal(err)
    }

    type gameData struct {
        ID string
    }

    err = tmpl.Execute(w, gameData{ID: id})
    if err != nil {
        log.Fatal(err)
    }
}

func (app *application) getGame(w http.ResponseWriter, r *http.Request) {
    type playerTribe struct {
        Name string
        Tribe string
    }
    type resultData struct {
        Game gameData
        Players []playerTribe
    }

    gameId := r.URL.Query().Get("id")
    if !app.db.gameExists(gameId) {
        http.Error(w, "Invalid id", http.StatusBadRequest)
        return
    }
    game := app.db.getGameData(gameId)
    playerIds := app.db.getGamePlayers(gameId)
    log.Print(playerIds)
    var players []playerTribe
    for _, playerId := range playerIds {
        name := app.db.getName(playerId)
        tribe := app.db.getTribe(playerId, gameId)
        players = append(players, playerTribe{name, tribe})
    }
    result := resultData{*game, players}

    tmpl, err := template.ParseFiles("html/game-result.tmpl")
    if err != nil {
        log.Fatal(err)
    }
    err = tmpl.Execute(w, result)
    if err != nil {
        log.Fatal(err)
    }
}

func (app *application) generateGame(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
        http.StatusMethodNotAllowed)
        return
    }
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusInternalServerError)
        return
    }

    id := r.FormValue("id")
    numberOfPlayers := len(app.db.getGamePlayers(id))
    // In future, this is where we respect the parameters given
    game := app.randomGame(numberOfPlayers)
    app.db.updateGame(id, &game)
    players := app.db.getGamePlayers(id)
    for _, player := range players {
       app.db.updateTribe(player, id, app.getRandomTribe())
    }
    log.Printf("Regenerated game %s", id)
}

func (app *application) generateRandomHash(n int) string {
    bytes := make([]byte, n)
    _, err := crand.Read(bytes)
    if err != nil {
        log.Fatal(err)
    }
    return hex.EncodeToString(bytes)
}

func (app *application) getBotDifficulties(n int) botDifficulties {
    var bots botDifficulties
    for i := 0; i < n; i++ {
        switch rand.Intn(4) {
        case 0:
            bots.Easy++
        case 1:
            bots.Normal++
        case 2:
            bots.Hard++
        case 3:
            bots.Crazy++
        }
    }
    return bots
}

func (app *application) getRandomListElement(list []string) string {
    return list[rand.Intn(len(list))]
}

func (app *application) randomMapType() string {
    return app.getRandomListElement([]string{
        "Dryland",
        "Lakes",
        "Pangea",
        "Continents",
        "Archipelago",
        "Water World",
    })
}

func (app *application) randomMapSize() string {
    return app.getRandomListElement([]string{
        "Tiny",
        "Small",
        "Normal",
    })
}

func (app *application) getRandomTribe() string {
    return app.getRandomListElement([]string{
        "Xin-xi",
        "Imperius",
        "Bardur",
        "Oumaji",
        "Kickoo",
        "Hoodrick",
        "Luxidoor",
        "Vengir",
        "Zebasi",
        "Ai-Mo",
        "Quetzali",
        "Yadakk",
        "Aquarion",
        "Elyrion",
        "Polaris",
        "Cymanti",
    })
}
