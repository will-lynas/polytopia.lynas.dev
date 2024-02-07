package main

import (
    "net/http"
    "fmt"
    "strconv"
    "math/rand"
)

const totalPlayers = 9

func getRandomListElement(list []string) string {
    return list[rand.Intn(len(list))]
}

func polytopiaHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        http.ServeFile(w, r, "static/polytopia.html")
        return
    }
    if r.Method != http.MethodPost {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
    }
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusInternalServerError)
        return
    }
    numPlayers, err := strconv.Atoi(r.FormValue("num-players"))
    if err != nil {
        http.Error(w, "Error parsing number of players",
        http.StatusInternalServerError)
        return
    }
    numBots := rand.Intn(totalPlayers - numPlayers + 1)
    mapTypes := r.Form["map-types"]
    if len(mapTypes) == 0 {
        http.Error(w, "No map types selected", http.StatusInternalServerError)
        return
    }
    mapType := getRandomListElement(mapTypes)
    mapSizes := r.Form["map-sizes"]
    if len(mapSizes) == 0 {
        http.Error(w, "No map sizes selected", http.StatusInternalServerError)
        return
    }
    mapSize := getRandomListElement(mapSizes)
    tribes := r.Form["tribes"]
    if len(tribes) == 0 {
        http.Error(w, "No tribes selected", http.StatusInternalServerError)
        return
    }
    botDifficulties := r.Form["bot-difficulties"]
    if len(botDifficulties) == 0 {
        http.Error(w, "No bot difficulty selected", http.StatusInternalServerError)
        return
    }

    responseContent := fmt.Sprintf(
        `Map type: %s<br>
        Map size: %s`,
        mapType, mapSize)

    responseContent += "<br><br>Bots: "
    for i := 0; i < numBots; i++ {
        responseContent += fmt.Sprintf(" %s", getRandomListElement(botDifficulties))
    }

    responseContent += "<br><br>Tribes:<br>"
    for i := 0; i < numPlayers; i++ {
        tribe := getRandomListElement(tribes)
        responseContent += fmt.Sprintf(
            `Player %d: <span class="spoiler">%-16s</span><br>`,
            i+1, tribe)
        }

        w.Header().Set("Content-Type", "text/html")
        fmt.Fprint(w, responseContent)
    }
