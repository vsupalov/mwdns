package main

import (
    "golang.org/x/net/websocket"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "net/url"
    "strconv"
    "time"

    "github.com/vsupalov/mwdns/game"
    "github.com/vsupalov/mwdns/utils"
)

const (
    DEV_MODE            = false //whether development is currently going on - constant template reload
    MIN_RES             = false //whether minified js/less resources shall be linked in the templates

    DEFAULT_PAIR_COUNT  = 10
    DEFAULT_GAME_TYPE   = game.GAME_TYPE_CLASSIC
    DEFAULT_MAX_PLAYERS = 0
)

// template related variables
var (
    addr = flag.String("addr", "localhost:8080", "http service address")
    homeTempl = utils.CreateAutoTemplate("templates/startView.html", DEV_MODE)
    gameTempl = utils.CreateAutoTemplate("templates/gameView.html", DEV_MODE)

    // a nameless struct, to hold information to be rendered in the template (whether to assume minified js/css and the card type to use)
    templateStruct = struct {
        CardInformation *utils.CardInformationStruct
        Minified bool
    } {
        &utils.CardInformation,
        MIN_RES,
    }
)

// global control entity, accessible to handlers
var (
    gameManager = game.NewGameManager()
)

func homeHandler(c http.ResponseWriter, req *http.Request) {
    homeTempl.Execute(c, templateStruct)
}

func gameHandler(w http.ResponseWriter, req *http.Request) {
    gameId := req.URL.Query().Get("g")
    if gameId == "" || len(gameId) != 6 {
        tryCreateNewGame(w,req)
    } else {
        //game already exists.
        log.Println("Game", gameId, "requested")

        //redirect to the start page if the game does not exist
        _, err := gameManager.GetGame(gameId)
        if err != nil {
            log.Println("Game not found: ", gameId)
            errmsg := "The game you were trying to join (id: <b>" + gameId + "</b>) doesn't exist!"
            http.Redirect(w, req, "/?errmsg="+url.QueryEscape(errmsg), 303)
            return
        }
        gameTempl.Execute(w, templateStruct)
    }
}

/*
Accepts web socket connections from users and tries to associate them to a given open game.
*/
func wsHandler(ws *websocket.Conn) {
    // Get the game we're talking about, if it exists.
    gameId := ws.Request().URL.Query().Get("g")

    gameInstance, err := gameManager.GetGame(gameId)
    if err != nil {
        log.Println("Websocket request with invalid gameId: ", gameId)
        websocket.Message.Send(ws, `{"msg": "err_gameid", "gid": "`+gameId+`"}`)
        return
    }

    // Check if the game can take any more players.
    if gameInstance.Players.Len() >= gameInstance.MaxPlayers && gameInstance.MaxPlayers > 0 {
        log.Println("Game is already full")
        websocket.Message.Send(ws, fmt.Sprintf(`{"msg": "err_gamefull", "gid": "%v", "max": %v}`, gameId, gameInstance.MaxPlayers))
        return
    }

    // create player for the game
    player := game.NewPlayer(ws)
    // As this works with channels internally, we are alrighty TODO: check for error?
    gameInstance.AddPlayer(player)

    //TODO: always remove the player instead of setting him/her inactive? Rather mark as inactive?
    defer func() { gameInstance.RemovePlayer(player) }()

    go player.Writer()
    player.Reader()
}

/*
This function is called by the gameHandler to initiate a new game, when no game ID is given or a provided ID is judged as invalid.
Parses GET arguments of a request and tries to create a new game with the interpreted parameters.
*/
func tryCreateNewGame(w http.ResponseWriter, req *http.Request) {
    // Use the chance to clean up ded gaems.
    // This is the "reaper" form of cleanup: cleanup every now and then
    // (i.e. whenever there is a request to /game).
    gameManager.CleanGames()

    // create a new game if no game ID is given
    nCards, err := strconv.Atoi(req.URL.Query().Get("n"))
    if err != nil {
        log.Println("Invalid number of pairs", req.URL.Query().Get("n"), ", defaulting to ", DEFAULT_PAIR_COUNT)
        nCards = DEFAULT_PAIR_COUNT
    }
    nCards *= 2

    gameType, err := strconv.Atoi(req.URL.Query().Get("t"))
    if err != nil || (gameType != game.GAME_TYPE_CLASSIC && gameType != game.GAME_TYPE_RUSH) {
        log.Println("Invalid game type", req.URL.Query().Get("t"), ", defaulting to ", DEFAULT_GAME_TYPE)
        gameType = DEFAULT_GAME_TYPE
    }

    maxPlayers, err := strconv.Atoi(req.URL.Query().Get("m"))
    if err != nil {
        if req.URL.Query().Get("m") == "∞" {
            maxPlayers = 0
        } else {
            log.Println("Invalid max player count", req.URL.Query().Get("m"), ", defaulting to ", DEFAULT_MAX_PLAYERS)
            maxPlayers = DEFAULT_MAX_PLAYERS
        }
    }

    //ct (id from json - the card sizes are important)
    cardType, err := strconv.Atoi(req.URL.Query().Get("ct"))
    if err != nil {
        log.Println("Invalid card type parameter", req.URL.Query().Get("ct"), ", defaulting to ", 0)
        cardType = 0
    }

    //cl (tight grid = 0/loose grid/stack)
    cardLayout, err := strconv.Atoi(req.URL.Query().Get("cl"))
    if err != nil {
        log.Println("Invalid card layout parameter", req.URL.Query().Get("cl"), ", defaulting to ", 0)
        cardLayout = 0
    }

    //cr (no rotations = 0/some/lots)
    cardRotation, err := strconv.Atoi(req.URL.Query().Get("cr"))
    if err != nil {
        log.Println("Invalid card rotation parameter", req.URL.Query().Get("cr"), ", defaulting to ", 0)
        cardRotation = 0
    }

    gameId := gameManager.CreateNewGame(nCards, gameType, maxPlayers, cardType, cardLayout, cardRotation, req.URL)
    http.Redirect(w, req, fmt.Sprintf("/game?g=%v", gameId), 303)
}

// https://groups.google.com/forum/#!topic/golang-nuts/n-GjwsDlRco
func maxAgeHandler(seconds int, h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", seconds))
        h.ServeHTTP(w, r)
    })
}

func main() {
    log.Println("__Initial setup")
    rand.Seed(time.Now().UTC().UnixNano())
    log.Println("_Loading templates")
    homeTempl.Load()
    gameTempl.Load()

    log.Println("_Parsing card information from JSON file")
    utils.ParseCardInformation()

    sm := http.NewServeMux()
    sm.HandleFunc("/", homeHandler)
    sm.HandleFunc("/game", gameHandler)

    // for development, just 1 minute of caching
    const CACHE_SECONDS = 60
    sm.Handle("/static/", http.StripPrefix("/static/",
        maxAgeHandler(CACHE_SECONDS, http.FileServer(http.Dir("static")))))
    sm.Handle("/ws", websocket.Handler(wsHandler))

    flag.Parse()
    log.Printf("__Starting Server on '%v'\n", *addr)
    s := http.Server{Handler: sm, Addr: *addr}
    if err := s.ListenAndServe(); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
