package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jruiznavarro/wargamestactics/internal/ai"
	"github.com/jruiznavarro/wargamestactics/internal/game"
	"github.com/jruiznavarro/wargamestactics/internal/game/army"
	"github.com/jruiznavarro/wargamestactics/internal/game/board"
	"github.com/jruiznavarro/wargamestactics/internal/game/core"
)

//go:embed static
var staticFiles embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server is the web server for the AoS battle simulator.
type Server struct {
	registry *army.FactionRegistry
	dataDir  string
}

// NewServer creates a new web server.
func NewServer(registry *army.FactionRegistry, dataDir string) *Server {
	return &Server{
		registry: registry,
		dataDir:  dataDir,
	}
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()

	// Serve embedded static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("static files: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoints
	mux.HandleFunc("/api/factions", s.handleFactions)

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	log.Printf("AoS Web Simulator listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleFactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ids := s.registry.FactionIDs()
	type factionInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var factions []factionInfo
	for _, id := range ids {
		f := s.registry.GetFaction(id)
		if f != nil {
			factions = append(factions, factionInfo{ID: f.ID, Name: f.Name})
		}
	}
	json.NewEncoder(w).Encode(factions)
}

// wsMessage is a generic WebSocket message envelope.
type wsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket client connected: %s", conn.RemoteAddr())

	// This goroutine handles one game session per WebSocket connection.
	// Wait for a "start_game" message from the client.
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		var msg wsMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			sendError(conn, "invalid JSON")
			continue
		}

		switch msg.Type {
		case "start_game":
			s.runGameSession(conn, msg.Data)
			return // game session ended
		case "ping":
			sendJSON(conn, wsMessage{Type: "pong"})
		default:
			sendError(conn, fmt.Sprintf("unexpected message type: %s", msg.Type))
		}
	}
}

type startGameRequest struct {
	Mode     string `json:"mode"`     // "pvai" or "aivai"
	Faction1 string `json:"faction1"` // e.g. "seraphon"
	Faction2 string `json:"faction2"` // e.g. "tzeentch"
	Seed     int64  `json:"seed"`     // 0 = random
	Rounds   int    `json:"rounds"`   // max battle rounds
}

func (s *Server) runGameSession(conn *websocket.Conn, data json.RawMessage) {
	var req startGameRequest
	if err := json.Unmarshal(data, &req); err != nil {
		sendError(conn, "invalid start_game data")
		return
	}
	if req.Rounds <= 0 {
		req.Rounds = 5
	}
	if req.Seed == 0 {
		req.Seed = time.Now().UnixNano()
	}

	f1 := s.registry.GetFaction(req.Faction1)
	f2 := s.registry.GetFaction(req.Faction2)
	if f1 == nil || f2 == nil {
		sendError(conn, "unknown faction(s)")
		return
	}

	// Create game from battleplan
	bp := board.GetBattleplan(board.BattleplanTable1, 1)
	g := game.NewGameFromBattleplan(req.Seed, bp)

	var webPlayer *WebPlayer

	switch req.Mode {
	case "aivai":
		p1 := ai.NewAIPlayer(1, f1.Name)
		p2 := ai.NewAIPlayer(2, f2.Name)
		g.AddPlayer(p1)
		g.AddPlayer(p2)
	case "pvai":
		webPlayer = NewWebPlayer(1, "Player")
		p2 := ai.NewAIPlayer(2, f2.Name)
		g.AddPlayer(webPlayer)
		g.AddPlayer(p2)
	default:
		sendError(conn, "unsupported mode (use pvai or aivai)")
		return
	}

	// Setup armies
	setupFactionArmy(g, f1, 1)
	setupFactionArmy(g, f2, 2)
	army.RegisterFactionRules(g.Rules, f1, 1)
	army.RegisterFactionRules(g.Rules, f2, 2)
	if len(f1.Formations) > 0 {
		army.RegisterFormationRules(g.Rules, f1, 0, 1)
	}
	if len(f2.Formations) > 0 {
		army.RegisterFormationRules(g.Rules, f2, 0, 2)
	}
	g.RegisterTerrainRules()

	// Send game_started confirmation
	startInfo := map[string]interface{}{
		"battleplan": bp.Name,
		"faction1":   f1.Name,
		"faction2":   f2.Name,
		"seed":       req.Seed,
		"mode":       req.Mode,
	}
	startData, _ := json.Marshal(startInfo)
	sendJSON(conn, wsMessage{Type: "game_started", Data: startData})

	if req.Mode == "aivai" {
		s.runAIvAISession(conn, g, req.Rounds)
	} else {
		s.runPvAISession(conn, g, webPlayer, req.Rounds)
	}
}

func (s *Server) runAIvAISession(conn *websocket.Conn, g *game.Game, maxRounds int) {
	// Capture the initial log length so we can stream new entries
	lastLogIdx := 0

	// Run game in a goroutine and send state updates
	done := make(chan struct{})
	go func() {
		defer close(done)
		g.RunGame(maxRounds)
	}()

	// Poll game state periodically
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// Game finished — send final state
			s.sendGameState(conn, g, 0)
			s.sendNewLogs(conn, g, &lastLogIdx)
			s.sendGameOver(conn, g)
			return
		case <-ticker.C:
			s.sendGameState(conn, g, 0)
			s.sendNewLogs(conn, g, &lastLogIdx)
		}
	}
}

func (s *Server) runPvAISession(conn *websocket.Conn, g *game.Game, wp *WebPlayer, maxRounds int) {
	lastLogIdx := 0

	// Run game in a background goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		g.RunGame(maxRounds)
	}()

	// Main loop: multiplex between WebPlayer state requests and incoming WebSocket messages
	for {
		select {
		case <-done:
			s.sendGameState(conn, g, wp.ID())
			s.sendNewLogs(conn, g, &lastLogIdx)
			s.sendGameOver(conn, g)
			return

		case prompt := <-wp.stateCh:
			// The game is asking for a player command — send state to client
			s.sendNewLogs(conn, g, &lastLogIdx)
			s.sendPrompt(conn, prompt)

			// Wait for client command
			for {
				_, raw, err := conn.ReadMessage()
				if err != nil {
					log.Printf("WebSocket read error: %v", err)
					close(wp.commandCh)
					return
				}

				var msg wsMessage
				if err := json.Unmarshal(raw, &msg); err != nil {
					sendError(conn, "invalid JSON")
					continue
				}

				if msg.Type == "command" {
					var wsCmd WSCommand
					if err := json.Unmarshal(msg.Data, &wsCmd); err != nil {
						sendError(conn, "invalid command data")
						continue
					}

					cmd := ParseWSCommand(wp.ID(), wsCmd)
					if cmd == nil {
						sendError(conn, "could not parse command")
						continue
					}

					wp.commandCh <- cmd
					break
				} else if msg.Type == "ping" {
					sendJSON(conn, wsMessage{Type: "pong"})
				}
			}
		}
	}
}

func (s *Server) sendGameState(conn *websocket.Conn, g *game.Game, playerID int) {
	view := g.View(playerID)
	data, _ := json.Marshal(view)
	sendJSON(conn, wsMessage{Type: "game_state", Data: data})
}

func (s *Server) sendNewLogs(conn *websocket.Conn, g *game.Game, lastIdx *int) {
	if len(g.Log) > *lastIdx {
		newLogs := g.Log[*lastIdx:]
		*lastIdx = len(g.Log)
		data, _ := json.Marshal(newLogs)
		sendJSON(conn, wsMessage{Type: "battle_log", Data: data})
	}
}

func (s *Server) sendPrompt(conn *websocket.Conn, prompt *PlayerPrompt) {
	data, _ := json.Marshal(prompt)
	sendJSON(conn, wsMessage{Type: "prompt", Data: data})
}

func (s *Server) sendGameOver(conn *websocket.Conn, g *game.Game) {
	result := map[string]interface{}{
		"winner":        g.Winner,
		"victoryPoints": g.VictoryPoints,
		"isOver":        g.IsOver,
	}
	for _, p := range g.Players {
		if p.ID() == g.Winner {
			result["winnerName"] = p.Name()
		}
	}
	data, _ := json.Marshal(result)
	sendJSON(conn, wsMessage{Type: "game_over", Data: data})
}

func sendJSON(conn *websocket.Conn, msg wsMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

func sendError(conn *websocket.Conn, errMsg string) {
	data, _ := json.Marshal(map[string]string{"message": errMsg})
	sendJSON(conn, wsMessage{Type: "error", Data: data})
}

// setupFactionArmy mirrors the CLI's setupFactionArmy function.
func setupFactionArmy(g *game.Game, faction *army.Faction, ownerID int) {
	baseY := 6.0
	if ownerID == 2 {
		baseY = 38.0
	}

	pointsSpent := 0
	pointsLimit := 1000
	xPos := 10.0
	isFirstHero := true

	for _, ws := range faction.Warscrolls {
		if !ws.HasKeyword("Hero") || pointsSpent+ws.Points > pointsLimit {
			continue
		}
		pos := core.Position{X: xPos, Y: baseY}
		u := g.CreateUnitFromSpec(ws.Name, ownerID, ws.ToCoreStats(), ws.ToCoreWeapons(),
			ws.UnitSize, pos, ws.BaseSizeInches(),
			ws.ToCoreKeywords(), ws.WardSave, ws.PowerLevel,
			ws.ToCoreSpells(), ws.ToCorePrayers())
		applyAbilities(u, &ws)
		u.FactionKeyword = faction.ID
		u.Tags = append([]string{}, ws.Tags...)
		if isFirstHero {
			u.IsGeneral = true
			isFirstHero = false
		}
		army.RegisterWarscrollAbilityRules(g.Rules, u, &ws)
		pointsSpent += ws.Points
		xPos += 8.0
		break
	}

	for _, ws := range faction.Warscrolls {
		if ws.HasKeyword("Hero") || pointsSpent+ws.Points > pointsLimit {
			continue
		}
		if xPos > 55.0 {
			break
		}
		pos := core.Position{X: xPos, Y: baseY}
		u := g.CreateUnitFromSpec(ws.Name, ownerID, ws.ToCoreStats(), ws.ToCoreWeapons(),
			ws.UnitSize, pos, ws.BaseSizeInches(),
			ws.ToCoreKeywords(), ws.WardSave, ws.PowerLevel,
			ws.ToCoreSpells(), ws.ToCorePrayers())
		applyAbilities(u, &ws)
		u.FactionKeyword = faction.ID
		u.Tags = append([]string{}, ws.Tags...)
		army.RegisterWarscrollAbilityRules(g.Rules, u, &ws)
		pointsSpent += ws.Points
		xPos += 8.0
	}
}

func applyAbilities(u *core.Unit, ws *army.Warscroll) {
	for _, ab := range ws.Abilities {
		switch ab.Effect {
		case "ward":
			if ab.Value > 0 && (u.WardSave == 0 || ab.Value < u.WardSave) {
				u.WardSave = ab.Value
			}
		case "strikeFirst":
			u.StrikeOrder = core.StrikeFirst
		case "strikeLast":
			u.StrikeOrder = core.StrikeLast
		}
	}
}

